package main

import (
	"crypto/tls"
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

type client struct {
	creds *api.Credentials
	certs *tls.Config
	addr  string
	exec  *app.Exec
}

var dir = ".auklet/message"
var persistor = broker.NewPersistor(dir)

func newclient(exec *app.Exec) *client {
	c := &client{
		exec: exec,
	}
	return c
}

func (c *client) runPipeline() {
	producer, err := broker.NewMQTTProducer(c.addr, c.certs, c.creds.Username, c.creds.Password, c.creds.Org)
	if err != nil {
		log.Fatal(err)
	}
	loader := broker.NewMessageLoader(dir)
	logger := agent.NewLogger(c.exec.Logs())
	server := agent.NewServer(c.exec.Data(), c.exec.Decoder())
	agentMessages := agent.NewMerger(logger, server)
	converter := schema.NewConverter(agentMessages, persistor, c.exec)
	requester := agent.NewPeriodicRequester(c.exec.Data(), server.Done)
	watcher := message.NewExitWatcher(converter, c.exec, persistor)
	merger := message.NewMerger(watcher, loader, requester)
	limiter := message.NewDataLimiter(merger, message.FilePersistor{".auklet/datalimit.json"})

	pollConfig := func() {
		poll := func() {
			dl, err := api.GetDataLimit()
			if err != nil {
				// TODO: send this over MQTT
				errorlog.Print(err)
				return
			}
			go func() { requester.Configure() <- dl.EmissionPeriod }()
			go func() { limiter.Configure() <- dl.Cellular }()
		}
		poll()
		for _ = range time.Tick(time.Hour) {
			poll()
		}
	}
	go pollConfig()

	producer.Serve(limiter)
}

func (c *client) prepare() bool {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		creds, err := api.GetCredentials()
		if err != nil {
			// TODO: send this over MQTT
			errorlog.Print(err)
		}
		c.creds = creds
	}()
	go func() {
		defer wg.Done()
		addr, err := api.GetBrokerAddr()
		if err != nil {
			// TODO: send this over MQTT
			errorlog.Print(err)
		}
		c.addr = addr
	}()
	go func() {
		defer wg.Done()
		certs, err := api.Certificates()
		if err != nil {
			// TODO: send this over MQTT
			errorlog.Print(err)
		}
		c.certs = certs
	}()
	wg.Wait()
	return c.addr != "" && c.certs != nil && c.creds != nil
}

func (c *client) run() {
	if err := api.Release(c.exec.CheckSum()); err != nil {
		errorlog.Print(err)
		// not released. Start the app, but don't serve it.
		if err := c.exec.Start(); err != nil {
			log.Fatal(err)
		}
		c.exec.Wait()
		return
	}

	if err := c.exec.AddSockets(); err != nil {
		log.Fatal(err)
	}

	if err := c.exec.Start(); err != nil {
		log.Fatal(err)
	}

	if err := c.exec.GetAgentVersion(); err != nil {
		log.Fatal(err)
	}

	if !c.prepare() {
		return
	}

	c.runPipeline()
}

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
	fmt.Printf("view OSS licenses: %v --licenses\n", os.Args[0])
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

func getConfig() config.Config {
	if version.Version == "local-build" {
		return config.LocalBuild()
	}
	return config.ReleaseBuild()
}

func init() {
	log.SetFlags(log.Lmicroseconds)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	if args[0] == "--licenses" {
		licenses()
		os.Exit(1)
	}
	log.Printf("Auklet Client version %s (%s)\n", version.Version, version.BuildDate)
	cfg := getConfig()
	api.BaseURL = cfg.BaseURL
	if !cfg.LogInfo {
		log.SetOutput(ioutil.Discard)
	}
	if !cfg.LogErrors {
		errorlog.SetOutput(ioutil.Discard)
	}
	exec, err := app.NewExec(args[0], args[1:]...)
	if err != nil {
		log.Fatal(err)
	}
	newclient(exec).run()
}
