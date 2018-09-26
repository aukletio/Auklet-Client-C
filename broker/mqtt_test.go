package broker

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/eclipse/paho.mqtt.golang"

	"github.com/ESG-USA/Auklet-Client-C/api"
)

// The testing strategy is to mock mqtt.Client. This is done with the
// broker.client interface.
//
// We would have mocked mqtt.Token, but it's _impossible_ as of release 1.1.1.
// See https://github.com/eclipse/paho.mqtt.golang/issues/203
//
// There is talk that the token concept will be done away with in 2.0, and the
// entire API will be revamped.
//
// Until then, we need another strategy. Our client can return concrete token
// types from package mqtt. Thankfully, our wait function provides a thin
// wrapper around the tokens, which we can mock. Thus wait is declared as a
// variable.

func init() {
	newClient(mqtt.NewClientOptions())
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
	orig := wait
	defer func() { wait = orig }()

	errConn := errors.New("connect error")
	cases := []struct {
		wait   func(token) error
		expect error
	}{
		{
			wait:   func(token) error { return nil },
			expect: nil,
		}, {
			wait:   func(token) error { return errConn },
			expect: errConn,
		},
	}

	conf := new(tls.Config)
	creds := new(api.Credentials)
	for i, c := range cases {
		wait = c.wait
		if _, err := NewMQTTProducer(Config{
			Address: "",
			Certs:   conf,
			Creds:   creds,
		}); err != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, err)
		}
	}
}

type channel chan Message

func (c channel) Output() <-chan Message {
	return c
}

func TestPublish(t *testing.T) {
	orig := wait
	defer func() { wait = orig }()

	errPublish := errors.New("publish error")
	cases := []func(token) error{
		func(token) error { return nil },
		func(token) error { return errPublish },
	}

	for _, c := range cases {
		wait = c
		source := make(channel)
		go func() {
			defer close(source)
			source <- Message{}
		}()
		MQTTProducer{c: klient{}}.Serve(source)
	}
}

type tok struct{}

func (tok) Wait() bool   { return false }
func (tok) Error() error { return nil }

func TestWait(t *testing.T) {
	if wait(tok{}) != nil {
		t.Fail()
	}
}

type mockAPI struct {
	err error
}

func (a mockAPI) Credentials() (*api.Credentials, error) {
	return new(api.Credentials), a.err
}

func (a mockAPI) BrokerAddress() (string, error) {
	return "", a.err
}

func (a mockAPI) Certificates() (*tls.Config, error) {
	return new(tls.Config), a.err
}

func TestNewConfig(t *testing.T) {
	fail := mockAPI{errors.New("error")}
	pass := mockAPI{nil}
	_, err := NewConfig(fail, "testdata/w/identification")
	if err == nil {
		t.Error(err)
	}

	_, err = NewConfig(pass, "testdata/w/identification")
	if err != nil {
		t.Error(err)
	}
}
