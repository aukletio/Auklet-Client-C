package broker

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/eclipse/paho.mqtt.golang"

	backend "github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// MQTTProducer wraps an MQTT Client.
type MQTTProducer struct {
	c       Client
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

// Client provides an MQTT client interface.
type Client interface {
	Connect() mqtt.Token
	Publish(string, byte, bool, interface{}) mqtt.Token
	Disconnect(uint)
}

// Config provides parameters for an MQTTProducer.
type Config struct {
	Creds  *backend.Credentials
	Client Client
}

// API consists of the backend interface needed to generate a Config.
type API interface {
	backend.Credentialer
	BrokerAddress() (string, error)
	Certificates() (*tls.Config, error)
}

// NewConfig returns a Config from the given API.
func NewConfig(api API) (Config, error) {
	creds, err := api.Credentials()
	if err != nil {
		return Config{}, err
	}

	addr, err := api.BrokerAddress()
	if err != nil {
		return Config{}, err
	}
	log.Printf("broker address: %v", addr)

	certs, err := api.Certificates()
	if err != nil {
		return Config{}, err
	}

	opt := mqtt.NewClientOptions()
	opt.AddBroker(addr)
	opt.SetTLSConfig(certs)
	opt.SetClientID(creds.ClientID)
	opt.SetCredentialsProvider(func() (string, string) {
		return creds.Username, creds.Password
	})

	return Config{
		Creds:  creds,
		Client: mqtt.NewClient(opt),
	}, nil
}

// NewMQTTProducer returns a new producer for the given input.
func NewMQTTProducer(cfg Config) (*MQTTProducer, error) {
	c := cfg.Client
	if err := wait(c.Connect()); err != nil {
		return nil, fmt.Errorf("connecting to broker: %v", err)
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

	for msg := range in.Output() {
		topic := fmt.Sprintf("c/%v/%v/%v", msg.Topic, p.org, p.id)
		err := wait(p.c.Publish(topic, 1, false, msg.Bytes))
		if err != nil {
			errorlog.Print("publishing to broker:", err)
			continue
		}
		log.Printf("producer: sent %+q", msg.Bytes)
		msg.Remove()
	}
}
