package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/spf13/afero"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	backend "github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/device"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/message"
	"github.com/ESG-USA/Auklet-Client-C/schema"
	"github.com/ESG-USA/Auklet-Client-C/version"
)

func main() {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	flags.Usage = func() {
		fmt.Printf("Usage of %v:\n", os.Args[0])
		fmt.Println("All non-flag arguments are treated as a command to run.")
		flags.PrintDefaults()
	}
	var (
		userVersion  string
		viewLicenses bool
		noNetwork    bool
	)
	flags.StringVar(&userVersion, "version", "", "user-defined version string")
	flags.BoolVar(&viewLicenses, "licenses", false, "view OSS licenses")
	flags.BoolVar(&noNetwork, "no-network", false, "disable network communication")
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	if viewLicenses {
		licenses()
		os.Exit(0)
	}

	if len(flags.Args()) == 0 {
		flags.Usage()
		os.Exit(1)
	}

	log.Printf("Auklet Client version %s (%s)\n", version.Version, version.BuildDate)
	e, err := app.NewExec(flags.Args()[0], flags.Args()[1:]...)
	if err != nil {
		log.Fatal(err)
	}

	// choose pipeline type
	var pipeline interface {
		run(exec) error
	}

	switch noNetwork {
	case true:
		pipeline = dumper{}
	case false:
		pipeline, err = newclient(userVersion)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := pipeline.run(e); err != nil {
		log.Fatal(err)
	}
}

func configureLogs(env config.Getenv) {
	log.SetFlags(log.Lmicroseconds)
	if !env.LogInfo() {
		log.SetOutput(ioutil.Discard)
	}
	if !env.LogErrors() {
		errorlog.SetOutput(ioutil.Discard)
	}
}

func licenses() {
	box := packr.NewBox("./licenses")
	// Print the Auklet license first, then iterate over all the others.
	format := "License for %v\n-------------------------\n%v"
	fmt.Printf(format, "Auklet Client", box.String("LICENSE"))
	for _, l := range box.List() {
		if l != "LICENSE" {
			header := "package: " + strings.Replace(l, "--", "/", 1)
			fmt.Printf("\n\n\n"+format, header, box.String(l))
		}
	}
}

type exec interface {
	schema.ExitSignalApp
	Connect() error
	Run() error
	AgentData() io.ReadWriter
	Decoder() *json.Decoder
	AppLogs() io.Reader
}

type dumper struct{}

func (dumper) run(e exec) error {
	if err := e.Connect(); err != nil {
		return err
	}

	server := agent.NewServer(e.AgentData(), e.Decoder())
	logger := agent.NewLogger(e.AppLogs())
	agent.NewPeriodicRequester(e.AgentData(), server.Done, nil)
	for m := range agent.Merge(server, logger).Output() {
		// dump the contents
		fmt.Printf(`type: %v
data: %v
error: %v

`, m.Type, string(m.Data), m.Error)
	}
	return nil
}

type client struct {
	msgPath      string // directory for storing unsent messages
	limPersistor message.Persistor
	api          interface {
		dataLimiter
		Release(string) error
	}
	userVersion string
	username    string
	appID       string
	macHash     string
	producer    interface{ Serve(broker.MessageSource) }
	fs          broker.Fs
}

func newclient(userVersion string) (*client, error) {
	env := config.OS
	configureLogs(env)
	appID := env.AppID()
	macHash := device.IfaceHash()
	fs := afero.NewOsFs()

	api := backend.API{
		BaseURL: env.BaseURL(version.Version),
		Key:     env.APIKey(),
		AppID:   appID,
		MacHash: macHash,

		CredsPath: ".auklet/identification",
		Fs:        fs,

		ReleasesEP:     backend.ReleasesEP,
		CertificatesEP: backend.CertificatesEP,
		DevicesEP:      backend.DevicesEP,
		ConfigEP:       backend.ConfigEP,
		DataLimitEP:    backend.DataLimitEP,
	}

	cfg, err := broker.NewConfig(api)
	if err != nil {
		return nil, err
	}

	producer, err := broker.NewMQTTProducer(cfg)
	if err != nil {
		return nil, err
	}

	return &client{
		msgPath:      ".auklet/message",
		limPersistor: message.FilePersistor{Path: ".auklet/datalimit.json"},
		api:          api,
		userVersion:  userVersion,
		appID:        appID,
		macHash:      macHash,
		producer:     producer,
		fs:           fs,
	}, nil
}

func (c *client) run(exec exec) error {
	err := c.api.Release(exec.CheckSum())
	if err != nil {
		errorlog.Print(err)
		// not released. Start the app, but don't serve it.
		return exec.Run()
	}

	if c.producer == nil {
		return nil
	}

	if err := exec.Connect(); err != nil {
		return err
	}

	cfg := pollConfig(c.api) // dataLimiter

	// main source of messages
	server := agent.NewServer(exec.AgentData(), exec.Decoder())

	c.producer.Serve(
		message.NewDataLimiter(
			c.limPersistor,
			cfg.limiter,
			schema.NewConverter(
				schema.Config{
					Monitor:     device.NewMonitor(),
					Persistor:   broker.NewPersistor(c.msgPath, c.fs, cfg.persistor),
					App:         exec, // schema.ExitSignalApp
					Username:    c.username,
					UserVersion: c.userVersion,
					AppID:       c.appID,
					MacHash:     c.macHash,
				},
				server,
				agent.NewLogger(exec.AppLogs()),
			),
			broker.NewMessageLoader(c.msgPath, c.fs),
			agent.NewPeriodicRequester(
				exec.AgentData(),
				server.Done,
				cfg.requester,
			),
		),
	)
	return nil
}

type dataLimiter interface {
	DataLimit() (*backend.DataLimit, error)
}

type configChans struct {
	requester chan int
	limiter   chan backend.CellularConfig
	persistor chan *int64
}

// pollConfig periodically polls the backend for data-limiting parameters and
// sends them on its output channels.
func pollConfig(api dataLimiter) configChans {
	c := configChans{
		requester: make(chan int, 1),
		limiter:   make(chan backend.CellularConfig, 1),
		persistor: make(chan *int64, 1),
	}

	go func() {
		poll := func() {
			dl, err := api.DataLimit()
			if err != nil {
				errorlog.Print(err)
				return
			}
			c.persistor <- dl.Storage
			c.requester <- dl.EmissionPeriod
			c.limiter <- dl.Cellular
		}

		poll()
		for _ = range time.Tick(time.Hour) {
			poll()
		}
	}()

	return c
}
