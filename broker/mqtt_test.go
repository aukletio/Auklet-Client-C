package broker

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
)

// The testing strategy is to mock mqtt.Client. This is done with the
// broker.client interface.
//
// We would have mocked mqtt.Token, but it's _impossible_ as of release 1.1.1.
// See https://github.com/eclipse/paho.mqtt.golang/issues/203
//
// There is talk that the token concept will be done away with in 2.0, and the
// entire API will be revamped. After that, we hope to be liberated from silly,
// hapless, inane, tumultutous code.
//
// Until then, we need another strategy. Our client can return concrete token
// types from package mqtt. Thankfully, our wait function provides a thin
// wrapper around the tokens, which we can mock. Thus wait is declared as a
// variable.

func init() {
	newClient = func(*mqtt.ClientOptions) client {
		return klient{}
	}
}

type klient struct{}

func (k klient) Connect() mqtt.Token {
	return &mqtt.ConnectToken{}
}

func (k klient) Publish(string, byte, bool, interface{}) mqtt.Token {
	return &mqtt.PublishToken{}
}

func (k klient) Disconnect(uint) {}

func TestConnect(t *testing.T) {
	errConn := errors.New("connect error")
	cases := []struct {
		wait   func(mqtt.Token) error
		expect error
	}{
		{
			wait:   func(mqtt.Token) error { return nil },
			expect: nil,
		}, {
			wait:   func(mqtt.Token) error { return errConn },
			expect: errConn,
		},
	}

	conf := new(tls.Config)
	for i, c := range cases {
		wait = c.wait
		if _, err := NewMQTTProducer("", conf, "", "", ""); err != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, err)
		}
	}
}

type channel chan Message

func (c channel) Output() <-chan Message {
	return c
}

func TestPublish(t *testing.T) {
	errPublish := errors.New("publish error")
	cases := []struct {
		wait func(mqtt.Token) error
	}{
		{
			wait: func(mqtt.Token) error { return nil },
		}, {
			wait: func(mqtt.Token) error { return errPublish },
		},
	}

	for _, c := range cases {
		wait = c.wait
		source := make(channel)
		go func() {
			defer close(source)
			source <- Message{}
		}()
		MQTTProducer{c: klient{}}.Serve(source)
	}
}
