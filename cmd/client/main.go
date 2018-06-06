package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ESG-USA/Auklet-Client/agent"
	"github.com/ESG-USA/Auklet-Client/api"
	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/config"
	"github.com/ESG-USA/Auklet-Client/device"
	"github.com/ESG-USA/Auklet-Client/errorlog"
	"github.com/ESG-USA/Auklet-Client/kafka"
	"github.com/ESG-USA/Auklet-Client/message"
	"github.com/ESG-USA/Auklet-Client/schema"
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
	logHandler := func(msg []byte) (kafka.Message, error) {
		return schema.NewAppLog(msg, c.app)
	}
	logger := agent.NewLogger("/tmp/auklet-log-" + strconv.Itoa(os.Getpid()), logHandler)
	server := newAgentServer(c.app)
	watcher := message.NewExitWatcher(server, c.app)
	merger := message.NewMerger(logger, watcher)
	limiter := message.NewDataLimiter(merger, c.app.ID)
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

	go logger.Serve()
	go server.Serve()
	go merger.Serve()
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
}

func getConfig() config.Config {
	if Version == "local-build" {
		return config.LocalBuild()
	}
	return config.ReleaseBuild()
}

func init() {
	log.SetFlags(log.Lmicroseconds)
	log.Printf("Auklet Client version %s (%s)\n", Version, BuildDate)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
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
