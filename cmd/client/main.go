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

	"github.com/aukletio/Auklet-Client-C/agent"
	backend "github.com/aukletio/Auklet-Client-C/api"
	"github.com/aukletio/Auklet-Client-C/app"
	"github.com/aukletio/Auklet-Client-C/broker"
	"github.com/aukletio/Auklet-Client-C/config"
	"github.com/aukletio/Auklet-Client-C/device"
	"github.com/aukletio/Auklet-Client-C/errorlog"
	"github.com/aukletio/Auklet-Client-C/message"
	"github.com/aukletio/Auklet-Client-C/schema"
	"github.com/aukletio/Auklet-Client-C/version"
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
		baseURL            string
		userVersion        string
		viewLicenses       bool
		noNetwork          bool
		serialOut          string
		printClientVersion bool
	)
	flags.StringVar(&baseURL, "base-url", "", "Auklet API URL; do not change unless instructed by support")
	flags.StringVar(&userVersion, "appVersion", "", "version of your application")
	flags.StringVar(&serialOut, "serial-out", "", "address of serial device to write JSON")
	flags.BoolVar(&printClientVersion, "version", false, "print Auklet Client version")
	flags.BoolVar(&viewLicenses, "licenses", false, "view OSS licenses")
	flags.BoolVar(&noNetwork, "no-network", false, "disable network communication")

	err := flags.Parse(os.Args[1:])
	switch {
	case err != nil:
		log.Fatal(err)

	case printClientVersion:
		fmt.Printf("Auklet Client version %s (%s)\n", version.Version, version.BuildDate)
		os.Exit(0)

	case viewLicenses:
		licenses()
		os.Exit(0)

	case len(flags.Args()) == 0:
		flags.Usage()
		os.Exit(1)
	}

	pipeline := func() interface{ run(exec) error } {
		if serialOut != "" {
			return newserial(serialOut, userVersion)
		}
		if noNetwork {
			return dumper{}
		}
		p, err := newclient(userVersion, baseURL)
		if err != nil {
			log.Fatal(err)
		}
		return p
	}()

	e, err := app.NewExec(flags.Args()[0], flags.Args()[1:]...)
	if err != nil {
		log.Fatal(err)
	}

	if err := pipeline.run(e); err != nil {
		log.Fatal(err)
	}
}

func configureLogs(env config.Getenv) {
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
	DataPoints() io.Reader
}

type dumper struct{}

func (dumper) run(e exec) error {
	if err := e.Connect(); err != nil {
		return err
	}

	server := agent.NewServer(e.AgentData(), e.Decoder())
	logger := agent.NewDataPointServer(e.DataPoints())
	agent.NewPeriodicRequester(e.AgentData(), server.Done, nil)
	for m := range agent.Merge(server, logger).Output() {
		// dump the contents
		fmt.Printf(`type: %v
data: %v

`, m.Type, string(m.Data))
	}
	return nil
}

type serial struct {
	userVersion string
	appID       string
	macHash     string
	addr        string // address of serial device
	fs          afero.Fs
}

func newserial(addr, userVersion string) serial {
	return serial{
		userVersion: userVersion,
		appID:       config.OS.AppID(),
		macHash:     device.IfaceHash(),
		addr:        addr,
		fs:          afero.NewOsFs(),
	}
}

func (s serial) run(e exec) error {
	if err := e.Connect(); err != nil {
		return err
	}
	server := agent.NewServer(e.AgentData(), e.Decoder())
	converter := schema.NewConverter(
		schema.Config{
			Monitor:     device.NewMonitor(),
			Persistor:   nil,
			App:         e, // schema.ExitSignalApp
			Username:    "",
			UserVersion: s.userVersion,
			AppID:       s.appID,
			MacHash:     s.macHash,
			Encoding:    schema.JSON,
		},
		server,
		agent.NewDataPointServer(e.DataPoints()),
	)
	merger := message.Merge(
		converter,
		agent.NewPeriodicRequester(e.AgentData(), server.Done, nil),
	)

	tryWrite := func(msg broker.Message) {
		f, err := s.fs.OpenFile(s.addr, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Printf("could not open %v: %v", s.addr, err)
			return
		}
		defer f.Close()

		b, err := json.Marshal(struct {
			Topic   string          `json:"topic"`
			Payload json.RawMessage `json:"payload"`
		}{
			Topic:   fmt.Sprintf("c/%v/", msg.Topic),
			Payload: msg.Bytes,
		})

		if err != nil {
			log.Printf("could not serialize message: %v", err)
			return
		}

		if _, err = f.Write(append(b, []byte("\r\n")...)); err != nil {
			log.Printf("could not write to %v: %v", s.addr, err)
		}
	}

	for msg := range merger.Output() {
		tryWrite(msg)
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

func selectPrefix(fs afero.Fs, env config.Getenv) (string, error) {
	prefixes := []string{
		"./",              // pwd
		env("HOME") + "/", // $HOME
	}
	for _, prefix := range prefixes {
		if err := fs.MkdirAll(prefix+".auklet", 0777); err == nil {
			return prefix, nil
		}
	}
	return afero.TempDir(fs, "", "auklet-")
}

func newclient(userVersion string, baseURL string) (*client, error) {
	env := config.OS
	fs := afero.NewOsFs()

	prefix, err := selectPrefix(fs, env)
	if err != nil {
		errorlog.Print(err)
	} else {
		log.Printf("selected prefix %q", prefix)
	}

	appID := env.AppID()
	macHash := device.IfaceHash()

	api := backend.API{
		BaseURL: env.BaseURL(baseURL),
		Key:     env.APIKey(),
		AppID:   appID,
		MacHash: macHash,

		CredsPath: prefix + ".auklet/identification",
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

	configureLogs(env)
	return &client{
		msgPath:      prefix + ".auklet/message",
		limPersistor: message.FilePersistor{Path: prefix + ".auklet/datalimit.json"},
		api:          api,
		userVersion:  userVersion,
		username:     cfg.Creds.Username,
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
					Encoding:    schema.MsgPack,
				},
				server,
				agent.NewDataPointServer(exec.DataPoints()),
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
