package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gobuffalo/packr"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/message"
	"github.com/ESG-USA/Auklet-Client-C/schema"
	"github.com/ESG-USA/Auklet-Client-C/version"
)

var (
	userVersion  string
	viewLicenses bool
)

func init() {
	log.SetFlags(log.Lmicroseconds)
	flag.StringVar(&userVersion, "version", "", "user-defined version string")
	flag.BoolVar(&viewLicenses, "licenses", false, "view OSS licenses")
}

func main() {
	flag.Parse()
	if viewLicenses {
		licenses()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	log.Printf("Auklet Client version %s (%s)\n", version.Version, version.BuildDate)
	apply(config.Get())
	c, err := newclient(args[0], args[1:]...)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.run(); err != nil {
		log.Fatal(err)
	}
}

func apply(cfg config.Config) {
	api.BaseURL = cfg.BaseURL
	if !cfg.LogInfo {
		log.SetOutput(ioutil.Discard)
	}
	if !cfg.LogErrors {
		errorlog.SetOutput(ioutil.Discard)
	}
}

func newclient(name string, args ...string) (*client, error) {
	exec, err := app.NewExec(args[0], args[1:]...)
	if err != nil {
		return nil, err
	}
	return &client{
		exec:        exec,
		userVersion: userVersion,
	}, nil
}

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
	fmt.Printf("view OSS licenses: %v -licenses\n", os.Args[0])
}

func licenses() {
	licensesBox := packr.NewBox("./licenses")
	licenses := licensesBox.List()
	// Print the Auklet license first, then iterate over all the others.
	format := "License for %v\n-------------------------\n%v"
	fmt.Printf(format, "Auklet Client", licensesBox.String("LICENSE"))
	for _, l := range licenses {
		if l != "LICENSE" {
			ownerName := strings.Split(l, "--")
			fmt.Printf("\n\n\n")
			header := fmt.Sprintf("package: %v/%v", ownerName[0], ownerName[1])
			fmt.Printf(format, header, licensesBox.String(l))
		}
	}
}

type client struct {
	creds       *api.Credentials
	certs       *tls.Config
	addr        string
	exec        *app.Exec
	userVersion string
}

func (c *client) run() error {
	if err := api.Do(api.Release{c.exec.CheckSum()}); err != nil {
		errorlog.Print(err)
		// not released. Start the app, but don't serve it.
		return c.exec.Run()
	}

	if err := c.exec.Connect(); err != nil {
		return err
	}

	if !c.prepare() {
		return nil
	}

	return c.runPipeline()
}

func (c *client) prepare() bool {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		creds, err := api.GetCredentials(".auklet/identification")
		if err != nil {
			// TODO: send this over MQTT
			errorlog.Print(err)
		}
		c.creds = creds
	}()
	go func() {
		defer wg.Done()
		addr := new(api.BrokerAddress)
		if err := api.Do(addr); err != nil {
			// TODO: send this over MQTT
			errorlog.Print(err)
		}
		c.addr = addr.Address
	}()
	go func() {
		defer wg.Done()
		certs := new(api.Certificates)
		if err := api.Do(certs); err != nil {
			// TODO: send this over MQTT
			errorlog.Print(err)
		}
		c.certs = certs.TLSConfig
	}()
	wg.Wait()
	return c.addr != "" && c.certs != nil && c.creds != nil
}

func (c *client) runPipeline() error {
	dir := ".auklet/message"

	producer, err := broker.NewMQTTProducer(c.addr, c.certs, c.creds)
	if err != nil {
		return err
	}

	persistor := broker.NewPersistor(dir)
	loader := broker.NewMessageLoader(dir)
	logger := agent.NewLogger(c.exec.AppLogs)
	server := agent.NewServer(c.exec.AgentData, c.exec.Decoder)
	agentMessages := agent.NewMerger(logger, server)
	converter := schema.NewConverter(agentMessages, persistor, c.exec, c.creds.Username, c.userVersion)
	requester := agent.NewPeriodicRequester(c.exec.AgentData, server.Done)
	merger := message.NewMerger(converter, loader, requester)
	limiter := message.NewDataLimiter(merger, message.FilePersistor{".auklet/datalimit.json"})

	pollConfig := func() {
		poll := func() {
			dl := new(api.DataLimit)
			if err := api.Do(dl); err != nil {
				// TODO: send this over MQTT
				errorlog.Print(err)
				return
			}
			go func() { requester.Configure() <- dl.EmissionPeriod }()
			go func() { limiter.Conf <- dl.Cellular }()
		}
		poll()
		for _ = range time.Tick(time.Hour) {
			poll()
		}
	}
	go pollConfig()

	producer.Serve(limiter)
	return nil
}
