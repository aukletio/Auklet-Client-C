// Command client is an Auklet client for ELF executables instrumented with
// libauklet.
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
	application "github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/config"
	"github.com/ESG-USA/Auklet-Client/device"
	"github.com/ESG-USA/Auklet-Client/kafka"
	"github.com/ESG-USA/Auklet-Client/message"
)

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
}

var (
	app  *application.App
	cfg  config.Config
	prod *kafka.Producer
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

func getConfig() config.Config {
	if Version == "local-build" {
		return config.LocalBuild()
	}
	return config.ReleaseBuild()
}

func setLogOutput() {
	if !cfg.Dump {
		log.SetOutput(ioutil.Discard)
	}
}

func setupProducer() {
	server := agent.NewServer("/tmp/auklet-"+strconv.Itoa(os.Getpid()), time.Minute, customHandlers)
	go server.Serve()

	watcher := message.NewExitWatcher(server, app)
	go watcher.Serve()

	queue := message.NewQueue(watcher, "persist")
	go queue.Serve()

	prod = kafka.NewProducer(queue)
}

func serveApp() {
	err := app.Start()
	if err != nil {
		os.Exit(1)
	}
	prod.Serve()
}

func main() {
	args := checkArgs()
	cfg = getConfig()
	api.BaseURL = cfg.BaseURL
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

	serveApp()
}
