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
}

var (
	osExit = os.Exit
	osArgs = os.Args
)

func main() {
	osExit(run(osArgs))
}

func run(args []string) int {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&userVersion, "version", "", "user-defined version string")
	fs.BoolVar(&viewLicenses, "licenses", false, "view OSS licenses")
	if err := fs.Parse(args); err != nil {
		log.Print(err)
		return 1
	}

	if viewLicenses {
		licenses()
		return 0
	}

	if len(args) == 0 {
		usage()
		return 1
	}

	log.Printf("Auklet Client version %s (%s)\n", version.Version, version.BuildDate)
	apply(config.Get())
	return startClient(args)
}

func startClient(args []string) int {
	c, err := newclient(args[0], args[1:]...)
	if err != nil {
		log.Print(err)
		return 1
	}
	if err := c.run(); err != nil {
		log.Print(err)
		return 1
	}
	return 0
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
	exec, err := app.NewExec(name, args...)
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

var (
	apiDo = api.Do
)

func (c *client) run() error {
	if err := apiDo(api.Release{c.exec.CheckSum()}); err != nil {
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
		if err := apiDo(addr); err != nil {
			// TODO: send this over MQTT
			errorlog.Print(err)
		}
		c.addr = addr.Address
	}()
	go func() {
		defer wg.Done()
		certs := new(api.Certificates)
		if err := apiDo(certs); err != nil {
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
			if err := apiDo(dl); err != nil {
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
