// Command client is an Auklet client for ELF executables instrumented with
// libauklet.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ESG-USA/Auklet-Client/api"
	application "github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/config"
	"github.com/ESG-USA/Auklet-Client/device"
	"github.com/ESG-USA/Auklet-Client/kafka"
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
	cfg       config.Config
	prod      *kafka.Producer
	kp        api.KafkaParams
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

func setLogOutput() {
	if !cfg.Dump {
		log.SetOutput(ioutil.Discard)
	}
}

func setupProducer() {
	prod = kafka.NewProducer()
}

func openSocket() {
	var err error
	sock, err = net.Listen("unixpacket", "/tmp/auklet-"+strconv.Itoa(os.Getpid()))
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
		err = prod.Send(schema.NewExit(app))
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
	api.BaseURL = cfg.BaseURL
}

func main() {
	args := checkArgs()
	getConfig()
	app = application.New(args)
	if !app.IsReleased {
		if err := app.Start(); err == nil {
			app.Wait()
		}
		os.Exit(0)
	}

	setupProducer()
	setLogOutput()
	go api.CreateOrGetDevice(device.MacHash, app.ID)

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
