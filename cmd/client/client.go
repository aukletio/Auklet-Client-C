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

	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/certs"
	"github.com/ESG-USA/Auklet-Client/config"
	"github.com/ESG-USA/Auklet-Client/device"
	"github.com/ESG-USA/Auklet-Client/producer"
	"github.com/ESG-USA/Auklet-Client/proxy"
	"github.com/ESG-USA/Auklet-Client/schema"
)

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
}

func main() {
	log.SetFlags(log.Lmicroseconds)
	log.Printf("Auklet Client version %s (%s)\n", Version, BuildDate)
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cfg := config.FromEnv()
	app := app.New(args, cfg.AppId)
	if !app.IsReleased(cfg.BaseUrl, cfg.APIKey) {
		if err := app.Start(); err == nil {
			app.Wait()
		}
		os.Exit(0)
	}

	certs := certs.FromUrl(cfg.BaseUrl, cfg.APIKey)
	prod := producer.New(cfg.Brokers, certs)
	prod.LogTopic = cfg.LogTopic
	if cfg.Dump {
		log.SetOutput(io.MultiWriter(os.Stdout, prod))
	} else {
		log.SetOutput(prod)
	}
	go device.CreateOrGet(cfg)
	sock, err := net.Listen("unixpacket", "/tmp/auklet-"+strconv.Itoa(os.Getpid()))
	if err != nil {
		log.Print(err)
	}
	defer sock.Close()

	profileHandler := func(data []byte) (m producer.Message, err error) {
		return schema.NewProfile(data, app, cfg.ProfileTopic)
	}

	errsigged := false
	errorsigHandler := func(data []byte) (m producer.Message, err error) {
		app.Cmd.Wait()
		errsigged = true
		log.Printf("app %v exited with error signal", app.Path)
		return schema.NewErrorSig(data, app, cfg.EventTopic)
	}

	logHandler := func(data []byte) (m producer.Message, err error) {
		m = schema.NewLog(data, cfg.LogTopic)
		return
	}

	server := proxy.Proxy{
		Listener: sock,
		Producer: prod,
		Handlers: map[string]proxy.Handler{
			"profile": profileHandler,
			"event":   errorsigHandler,
			"log":     logHandler,
		},
	}

	err = app.Start()
	if err != nil {
		os.Exit(1)
	}
	server.Serve()
	if !errsigged {
		app.Wait()
		err = prod.Send(schema.NewExit(app, cfg.EventTopic))
		if err != nil {
			log.Print(err)
		}
	}
}
