package schema

import (
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type persistor struct{}

func (persistor) CreateMessage(*broker.Message) error { return nil }

type source chan agent.Message

func (s source) Output() <-chan agent.Message { return s }

type app struct{}

func (app) ID() string           { return "app id" }
func (app) CheckSum() string     { return "checksum" }
func (app) Wait()                {}
func (app) ExitStatus() int      { return 42 }
func (app) Signal() string       { return "something" }
func (app) AgentVersion() string { return "something" }

func TestConverter(t *testing.T) {
	type converterCase struct {
		input agent.Message
		err   string
	}
	cases := []converterCase{
		{
			input: agent.Message{Type: "event"},
			err:   "",
		}, {
			input: agent.Message{Type: "profile"},
			err:   "",
		}, {
			input: agent.Message{Type: "cleanExit"},
			err:   "",
		}, {
			input: agent.Message{Type: "unknown"},
			err:   `message of type "unknown" not handled`,
		},
	}
	for i, c := range cases {
		s := make(source)
		converter := NewConverter(s, persistor{}, app{}, "username")
		s <- c.input
		m := <-converter.Output()
		if m.Error != c.err {
			t.Errorf("case %v: got %v, expected %v", i, m.Error, c.err)
		}
		close(s)
	}
}

func TestDrop(t *testing.T) {
	s := make(source)
	NewConverter(s, persistor{}, app{}, "username")
	s <- agent.Message{Type: "log"}
	close(s)
}

func TestConvert(t *testing.T) {
	s := make(source)
	close(s)
	c := NewConverter(s, persistor{}, app{}, "username")
	c.convert(agent.Message{Type: "log"})
	c.convert(agent.Message{Type: "applog"})
}

func TestMarshal(t *testing.T) {
	m := marshal(func() {}, 0)
	if m.Error == "" {
		t.Fail()
	}
}
