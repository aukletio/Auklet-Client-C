package broker

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/eclipse/paho.mqtt.golang"

	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// MQTTProducer wraps an MQTT client.
type MQTTProducer struct {
	c       client
	org, id string
}

// wait turns Paho's async API into a sync API.
var wait = func(t mqtt.Token) error {
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

// NewMQTTProducer returns a new producer for the given input.
func NewMQTTProducer(addr string, t *tls.Config, creds *api.Credentials) (*MQTTProducer, error) {
	opt := mqtt.NewClientOptions()
	opt.AddBroker(addr)
	opt.SetTLSConfig(t)
	opt.SetClientID(creds.Username)
	opt.SetCredentialsProvider(func() (string, string) {
		return creds.Username, creds.Password
	})
	c := newClient(opt)

	if err := wait(c.Connect()); err != nil {
		return nil, err
	}
	log.Print("producer: connected")

	return &MQTTProducer{
		c:   c,
		org: creds.Org,
		id:  creds.Username,
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
