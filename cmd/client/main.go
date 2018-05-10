package main

import (
	"github.com/ESG-USA/Auklet-Client/api"
	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/config"
)

type client struct {
	cfg   config.Config
	app   *app.App
	api   api.API
	kafka api.KafkaParams
	
	prod  *producer.Producer
}

func newclient(cfg config.Config, args []string) client {
	c := client {
		cfg: cfg,
		app: app.New(args),
		api: api.New(cfg.BaseURL, cfg.APIKey),
	}

	kafka := backend.KafkaParams()
	certs := backend.Certificates()
}

func (c client) createPipeline() {
	addr := "/tmp/auklet-"+strconv.Itoa(os.Getpid())
	server := agent.NewServer(addr, customHandlers)
	watcher := message.NewExitWatcher(server, c.app, c.kafka.EventTopic)
	limiter := message.NewDataLimiter(watcher, ".auklet/limit.json")
	queue := message.NewQueue(limiter, ".auklet")
	c.prod = producer.New(queue, c.kafka.Brokers, c.certs)

	go server.Serve()
	go watcher.Serve()
	go limiter.Serve()
	go queue.Serve()

	return c
}

func (c client) run() {
	if !c.backend.Release(c.app.CheckSum) {
		// not released. Start the app, but don't serve it.
		if err := c.app.Start(), err == nil {
			app.Wait()
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

func init() {
	log.SetFlags(log.Lmicroseconds)
	log.Printf("Auklet Client version %s (%s)\n", Version, BuildDate)
}

func main() {
	args := checkArgs()
	cfg := getConfig()
	if !cfg.Dump {
		log.SetOutput(ioutil.Discard)
	}
	c := newclient(cfg, args)
}
