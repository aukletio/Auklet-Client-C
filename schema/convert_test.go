package schema

import (
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type persistor struct {
	err error
}

func (p persistor) CreateMessage(*broker.Message) error { return p.err }

type source chan agent.Message

func (s source) Output() <-chan agent.Message { return s }

type app struct{}

func (app) ID() string           { return "app id" }
func (app) CheckSum() string     { return "checksum" }
func (app) ExitStatus() int      { return 42 }
func (app) Signal() string       { return "something" }
func (app) AgentVersion() string { return "something" }

var cfg = Config{
	Persistor:   persistor{},
	App:         app{},
	Username:    "username",
	UserVersion: "userVersion",
	AppID:       "app id",
	MacHash:     "mac hash",
}

func TestConverter(t *testing.T) {
	type converterCase struct {
		input agent.Message
		ok    bool
	}
	cases := []converterCase{
		{input: agent.Message{Type: "event"}, ok: true},
		{input: agent.Message{Type: "profile"}, ok: true},
		{input: agent.Message{Type: "cleanExit"}, ok: true},
		{input: agent.Message{Type: "unknown"}, ok: false},
	}
	for i, c := range cases {
		s := make(source)
		converter := NewConverter(cfg, s)
		s <- c.input
		m := <-converter.Output()
		ok := m.Error == ""
		if ok != c.ok {
			t.Errorf("case %v: got %v, expected %v: %v", i, ok, c.ok, m.Error)
		}
		close(s)
	}
}
