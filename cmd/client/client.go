// Command client is an Auklet client for ELF executables instrumented with
// libauklet.
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	auklet "github.com/ESG-USA/Auklet-Client/api"
	application "github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/config"
	"github.com/ESG-USA/Auklet-Client/device"
	"github.com/ESG-USA/Auklet-Client/producer"
	"github.com/ESG-USA/Auklet-Client/proxy"
	"github.com/ESG-USA/Auklet-Client/schema"
)

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
}

var (
	server    proxy.Proxy
	sock      net.Listener
	app       *application.App
	api       auklet.API
	cfg       config.Config
	prod      *producer.Producer
	kp        auklet.KafkaParams
	errsigged = false
)

func init() {
	log.SetFlags(log.Lmicroseconds)
	log.Printf("Auklet Client version %s (%s)\n", Version, BuildDate)
}

func checkArgs() (args []string) {
	args = os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	return
}

func checkRelease() {
	if !api.Release(app.CheckSum) {
		if err := app.Start(); err == nil {
			app.Wait()
		}
		os.Exit(0)
	}
}

func setLogOutput() {
	var w io.Writer
	if cfg.Dump {
		w = io.MultiWriter(os.Stdout, prod)
	} else {
		w = prod
	}
	log.SetOutput(w)
}

func setupProducer() {
	kp = api.KafkaParams()
	prod = producer.New(kp.Brokers, api.Certificates())
	if prod != nil {
		prod.LogTopic = kp.LogTopic
	}
}

func openSocket() {
	var err error
	sock, err = net.Listen("unix", "/tmp/auklet-"+strconv.Itoa(os.Getpid()))
	if err != nil {
		log.Print(err)
	}
}

func serveApp() {
	err := app.Start()
	if err != nil {
		os.Exit(1)
	}
	server.Serve()
	if !errsigged {
		app.Wait()
		err = prod.Send(schema.NewExit(app, kp.EventTopic))
		if err != nil {
			log.Print(err)
		}
	}
}

func getConfig() {
	if Version == "local-build" {
		cfg = config.LocalBuild()
	} else {
		cfg = config.ReleaseBuild()
	}
}

func main() {
	args := checkArgs()
	getConfig()
	api = auklet.New(cfg.BaseURL, cfg.APIKey)
	app = application.New(args, cfg.AppID)

	checkRelease()
	setupProducer()
	setLogOutput()
	go api.CreateOrGetDevice(device.MacHash, cfg.AppID)

	openSocket()
	defer sock.Close()

	server = proxy.Proxy{
		Listener: sock,
		Producer: prod,
		Handlers: customHandlers,
		Interval: time.Second,
	}

	serveApp()
}
