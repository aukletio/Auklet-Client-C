package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/packr"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/device"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/message"
	"github.com/ESG-USA/Auklet-Client-C/schema"
)

type server interface {
	Serve()
}

var application *app.App

type client struct {
	prod server
}

var dir = ".auklet/message"
var persistor = broker.NewPersistor(dir)

func newclient() *client {
	c := &client{}
	go api.CreateOrGetDevice(device.MacHash, application.ID())
	return c
}

func (c *client) createPipeline() {
	loader := broker.NewMessageLoader(dir)
	logger := agent.NewLogger(application.Logs())
	server := agent.NewServer(application.Data())
	agentMessages := agent.NewMerger(logger, server)
	converter := schema.NewConverter(agentMessages, persistor, application)
	requester := agent.NewPeriodicRequester(application.Data())
	watcher := message.NewExitWatcher(converter, application, persistor)
	merger := message.NewMerger(watcher, loader, requester)
	limiter := message.NewDataLimiter(merger, message.FilePersistor{".auklet/datalimit.json"})
	c.prod = broker.NewProducer(limiter)

	pollConfig := func() {
		poll := func() {
			dl := api.GetDataLimit(application.ID()).Config
			go func() { requester.Configure() <- dl.EmissionPeriod }()
			go func() { limiter.Configure() <- dl.Cellular }()
		}
		poll()
		for _ = range time.Tick(time.Hour) {
			poll()
		}
	}
	go pollConfig()
}

func (c *client) run() {
	if !application.IsReleased() {
		// not released. Start the app, but don't serve it.
		if err := application.Start(); err == nil {
			application.Wait()
		}
		os.Exit(0)
	}

	c.createPipeline()
	err := application.Start()
	if err != nil {
		os.Exit(1)
	}
	c.prod.Serve()
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
	if Version == "local-build" {
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
	log.Printf("Auklet Client version %s (%s)\n", Version, BuildDate)
	cfg := getConfig()
	api.BaseURL = cfg.BaseURL
	if !cfg.LogInfo {
		log.SetOutput(ioutil.Discard)
	}
	if !cfg.LogErrors {
		errorlog.SetOutput(ioutil.Discard)
	}
	application = app.New(args)
	newclient().run()
}
