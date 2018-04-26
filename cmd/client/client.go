// Command client is an Auklet client for ELF executables instrumented with
// libauklet.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/ESG-USA/Auklet-Client/agent"
	auklet "github.com/ESG-USA/Auklet-Client/api"
	application "github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/config"
	"github.com/ESG-USA/Auklet-Client/device"
	"github.com/ESG-USA/Auklet-Client/message"
	"github.com/ESG-USA/Auklet-Client/producer"
)

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
}

var (
	app  *application.App
	api  auklet.API
	cfg  config.Config
	prod *producer.Producer
	kp   auklet.KafkaParams
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

func exitIfNotReleased() {
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

	server := agent.NewServer("/tmp/auklet-"+strconv.Itoa(os.Getpid()), customHandlers)
	go server.Serve()

	watcher := message.NewExitWatcher(server, app, kp.EventTopic)
	go watcher.Serve()

	queue := message.NewQueue(watcher, "persist")
	go queue.Serve()

	prod = producer.New(queue, kp.Brokers, api.Certificates())
	if prod != nil {
		prod.LogTopic = kp.LogTopic
	}
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
	api = auklet.New(cfg.BaseURL, cfg.APIKey)
	app = application.New(args, cfg.AppID)

	exitIfNotReleased()
	setupProducer()
	//setLogOutput()
	go api.CreateOrGetDevice(device.MacHash, cfg.AppID)

	serveApp()
}
