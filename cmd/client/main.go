package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/packr"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/device"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/kafka"
	"github.com/ESG-USA/Auklet-Client-C/message"
	"github.com/ESG-USA/Auklet-Client-C/schema"
)

type client struct {
	app  *app.App
	prod *kafka.Producer
}

func newclient(args []string) *client {
	c := &client{app: app.New(args)}
	go api.CreateOrGetDevice(device.MacHash, c.app.ID)
	return c
}

func newAgentServer(app *app.App) agent.Server {
	addr := "/tmp/auklet-" + strconv.Itoa(os.Getpid())
	handlers := map[string]agent.Handler{
		"profile": func(data []byte) (kafka.Message, error) {
			return schema.NewProfile(data, app)
		},
		"event": func(data []byte) (kafka.Message, error) {
			app.Cmd.Wait()
			log.Printf("app %v exited with error signal", app.Path)
			return schema.NewErrorSig(data, app)
		},
		"log": func(data []byte) (kafka.Message, error) {
			return schema.NewLog(data)
		},
	}
	return agent.NewServer(addr, handlers)
}

func (c *client) createPipeline() {
	if err := os.MkdirAll(".auklet/message", 0777); err != nil {
		errorlog.Print(err)
	}
	server := newAgentServer(c.app)
	watcher := message.NewExitWatcher(server, c.app)
	limiter := message.NewDataLimiter(watcher, c.app.ID)
	queue := message.NewQueue(limiter)
	c.prod = kafka.NewProducer(queue)
	pollConfig := func() {
		poll := func() {
			dl := api.GetDataLimit(c.app.ID).Config
			go func() { server.Configure() <- dl.EmissionPeriod }()
			go func() { limiter.Configure() <- dl.Cellular }()
		}
		poll()
		for _ = range time.Tick(time.Hour) {
			poll()
		}
	}

	go server.Serve()
	go watcher.Serve()
	go limiter.Serve()
	go queue.Serve()
	go pollConfig()
}

func (c *client) run() {
	if !c.app.IsReleased {
		// not released. Start the app, but don't serve it.
		if err := c.app.Start(); err == nil {
			c.app.Wait()
		}
		os.Exit(0)
	}

	c.createPipeline()
	err := c.app.Start()
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
	format := "License for %v\n-------------------------\n%v\n"
	fmt.Printf(format, "Auklet Client", licensesBox.String("LICENSE"))
	for _, l := range licenses {
		if l != "LICENSE" {
			ownerName := strings.Split(l, "--")
			fmt.Println("\n\n")
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
	newclient(args).run()
}
