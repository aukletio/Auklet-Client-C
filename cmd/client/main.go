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
	api := backend.API{
		BaseURL: env.BaseURL(version.Version),
		Key:     env.APIKey(),
		AppID:   appID,
		MacHash: macHash,

		ReleasesEP:     backend.ReleasesEP,
		CertificatesEP: backend.CertificatesEP,
		DevicesEP:      backend.DevicesEP,
		ConfigEP:       backend.ConfigEP,
		DataLimitEP:    backend.DataLimitEP,
	}
	run(api, os.Args[1:], appID, macHash)
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

func run(api api, args []string, appID, macHash string) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	var userVersion string
	var viewLicenses bool
	fs.StringVar(&userVersion, "version", "", "user-defined version string")
	fs.BoolVar(&viewLicenses, "licenses", false, "view OSS licenses")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}

	if viewLicenses {
		licenses()
		os.Exit(0)
	}

	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	log.Printf("Auklet Client version %s (%s)\n", version.Version, version.BuildDate)
	exec, err := app.NewExec(args[0], args[1:]...)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := broker.NewConfig(api, ".auklet/identification") // broker.API
	if err != nil {
		errorlog.Print(err)
	}

	producer, err := broker.NewMQTTProducer(cfg)
	if err != nil {
		errorlog.Print(err)
	}

	c := client{
		msgPath:     ".auklet/message",
		limPath:     ".auklet/datalimit.json",
		api:         api,
		exec:        exec,
		userVersion: userVersion,
		appID:       appID,
		macHash:     macHash,
		producer:    producer,
	}

	if err := c.run(); err != nil {
		log.Fatal(err)
	}
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

type api interface {
	Release(checksum string) error
	broker.API
	dataLimiter
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
	msgPath string
	limPath string
	api     interface {
		dataLimiter
		Release(string) error
	}
	exec        exec
	userVersion string
	username    string
	appID       string
	macHash     string
	producer    interface{ Serve(broker.MessageSource) }
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
	reqConfig, limConfig := pollConfig(c.api) // dataLimiter

	// main source of messages
	server := agent.NewServer(c.exec.AgentData(), c.exec.Decoder())

	c.producer.Serve(
		message.NewDataLimiter(
			message.FilePersistor{Path: c.limPath},
			limConfig,
			schema.NewConverter(
				schema.Config{
					Monitor:     device.NewMonitor(),
					Persistor:   broker.NewPersistor(c.msgPath),
					App:         c.exec, // schema.ExitSignalApp
					Username:    c.username,
					UserVersion: c.userVersion,
					AppID:       c.appID,
					MacHash:     c.macHash,
				},
				server,
				agent.NewLogger(c.exec.AppLogs()),
			),
			broker.NewMessageLoader(c.msgPath),
			agent.NewPeriodicRequester(
				c.exec.AgentData(),
				server.Done,
				reqConfig,
			),
		),
	)
}

type dataLimiter interface {
	DataLimit() (*backend.DataLimit, error)
}

// pollConfig periodically polls the backend for data-limiting parameters and
// sends them on its output channels.
func pollConfig(api dataLimiter) (<-chan int, <-chan backend.CellularConfig) {
	reqConfig := make(chan int, 1)
	limConfig := make(chan backend.CellularConfig, 1)

	poll := func() {
		dl, err := api.DataLimit()
		if err != nil {
			errorlog.Print(err)
			return
		}
		reqConfig <- dl.EmissionPeriod
		limConfig <- dl.Cellular
	}

	go func() {
		poll()
		for _ = range time.Tick(time.Hour) {
			poll()
		}
	}()

	return reqConfig, limConfig
}
