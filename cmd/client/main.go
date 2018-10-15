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
	run(fs, api, os.Args[1:], appID, macHash)
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

func run(fs broker.Fs, api backend.API, args []string, appID, macHash string) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	usage := func() {
		fmt.Printf("Usage of %v:\n", os.Args[0])
		fmt.Println("All non-flag arguments are treated as a command to run.")
		flags.PrintDefaults()
	}
	flags.SetOutput(os.Stdout)
	flags.Usage = usage
	var userVersion string
	var viewLicenses bool
	flags.StringVar(&userVersion, "version", "", "user-defined version string")
	flags.BoolVar(&viewLicenses, "licenses", false, "view OSS licenses")
	if err := flags.Parse(args); err != nil {
		log.Fatal(err)
	}

	if viewLicenses {
		licenses()
		os.Exit(0)
	}

	if len(flags.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	log.Printf("Auklet Client version %s (%s)\n", version.Version, version.BuildDate)
	exec, err := app.NewExec(flags.Args()[0], flags.Args()[1:]...)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := broker.NewConfig(api)
	if err != nil {
		errorlog.Print(err)
	}

	producer, err := broker.NewMQTTProducer(cfg)
	if err != nil {
		errorlog.Print(err)
	}

	c := client{
		msgPath:      ".auklet/message",
		limPersistor: message.FilePersistor{Path: ".auklet/datalimit.json"},
		api:          api,
		exec:         exec,
		userVersion:  userVersion,
		appID:        appID,
		macHash:      macHash,
		producer:     producer,
		fs:           fs,
	}

	if err := c.run(); err != nil {
		log.Fatal(err)
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

type client struct {
	msgPath      string // directory for storing unsent messages
	limPersistor message.Persistor
	api          interface {
		dataLimiter
		Release(string) error
	}
	exec        exec
	userVersion string
	username    string
	appID       string
	macHash     string
	producer    interface{ Serve(broker.MessageSource) }
	fs          broker.Fs
}

func (c *client) run() error {
	err := c.api.Release(c.exec.CheckSum())
	if err != nil {
		errorlog.Print(err)
		// not released. Start the app, but don't serve it.
		return c.exec.Run()
	}

	if c.producer == nil {
		return nil
	}

	if err := c.exec.Connect(); err != nil {
		return err
	}

	c.runPipeline()
	return nil
}

func (c *client) runPipeline() {
	cfg := pollConfig(c.api) // dataLimiter

	// main source of messages
	server := agent.NewServer(c.exec.AgentData(), c.exec.Decoder())

	c.producer.Serve(
		message.NewDataLimiter(
			c.limPersistor,
			cfg.limiter,
			schema.NewConverter(
				schema.Config{
					Monitor:     device.NewMonitor(),
					Persistor:   broker.NewPersistor(c.msgPath, c.fs, cfg.persistor),
					App:         c.exec, // schema.ExitSignalApp
					Username:    c.username,
					UserVersion: c.userVersion,
					AppID:       c.appID,
					MacHash:     c.macHash,
				},
				server,
				agent.NewLogger(c.exec.AppLogs()),
			),
			broker.NewMessageLoader(c.msgPath, c.fs),
			agent.NewPeriodicRequester(
				c.exec.AgentData(),
				server.Done,
				cfg.requester,
			),
		),
	)
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
