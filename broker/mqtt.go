package broker

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/eclipse/paho.mqtt.golang"

	backend "github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// MQTTProducer wraps an MQTT client.
type MQTTProducer struct {
	c       client
	org, id string
}

type token interface {
	Wait() bool
	Error() error
}

// wait turns Paho's async API into a sync API.
var wait = func(t token) error {
	t.Wait()
	return t.Error()
}

type client interface {
	Connect() mqtt.Token
	Publish(string, byte, bool, interface{}) mqtt.Token
	Disconnect(uint)
}

// newClient allows us to mock the MQTT client in tests.
var newClient = func(o *mqtt.ClientOptions) client {
	return mqtt.NewClient(o)
}

// Config provides parameters for an MQTTProducer.
type Config struct {
	Address string
	Certs   *tls.Config
	Creds   *backend.Credentials
}

// API consists of the backend interface needed to generate a Config.
type API interface {
	backend.Credentialer
	BrokerAddress() (string, error)
	Certificates() (*tls.Config, error)
}

// NewConfig returns a Config from the given API.
// If there is a credentials file at idPath, it is loaded.
// If new credentials are obtained, they are stored to idPath.
func NewConfig(api API, idPath string) (Config, error) {
	errChan := make(chan error)
	var c Config
	go func() {
		creds, err := backend.GetCredentials(api, idPath)
		errChan <- err
		c.Creds = creds
	}()
	go func() {
		addr, err := api.BrokerAddress()
		errChan <- err
		c.Address = addr
	}()
	go func() {
		certs, err := api.Certificates()
		errChan <- err
		c.Certs = certs
	}()
	for _ = range [3]struct{}{} {
		if err := <-errChan; err != nil {
			return c, err
		}
	}
	return c, nil
}

// NewMQTTProducer returns a new producer for the given input.
func NewMQTTProducer(cfg Config) (*MQTTProducer, error) {
	opt := mqtt.NewClientOptions()
	opt.AddBroker(cfg.Address)
	opt.SetTLSConfig(cfg.Certs)
	opt.SetClientID(cfg.Creds.ClientID)
	opt.SetCredentialsProvider(func() (string, string) {
		return cfg.Creds.Username, cfg.Creds.Password
	})
	c := newClient(opt)

	if err := wait(c.Connect()); err != nil {
		return nil, err
	}
	log.Print("producer: connected")

	return &MQTTProducer{
		c:   c,
		org: cfg.Creds.Org,
		id:  cfg.Creds.Username,
	}, nil
}

// Serve launches p, enabling it to send and receive messages.
func (p MQTTProducer) Serve(in MessageSource) {
	defer func() {
		p.c.Disconnect(0)
		log.Print("producer: disconnected")
	}()

	topic := map[Topic]string{
		Profile: "profiler",
		Event:   "events",
		Log:     "logs",
	}
	for k, v := range topic {
		topic[k] = fmt.Sprintf("c/%v/%v/%v", v, p.org, p.id)
	}

	for msg := range in.Output() {
		if err := wait(p.c.Publish(topic[msg.Topic], 1, false, []byte(msg.Bytes))); err != nil {
			errorlog.Print("producer:", err)
			continue
		}
		log.Printf("producer: sent %+q", msg.Bytes)
		msg.Remove()
	}
}
