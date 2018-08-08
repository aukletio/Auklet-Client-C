package schema

import (
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type persistor struct{}

func (persistor) CreateMessage(broker.Message) error { return nil }

type source chan agent.Message

func (s source) Output() <-chan agent.Message { return s }

type app struct{}

func (app) ID() string                  { return "app id" }
func (app) CheckSum() string            { return "checksum" }
func (app) Wait()                       {}
func (app) ExitStatus() int             { return 42 }
func (app) AgentVersion() <-chan string { return nil }

func TestConverter(t *testing.T) {
	type converterCase struct {
		input agent.Message
		err   string
	}
	cases := []converterCase{
		{
			input: agent.Message{Type: "event"},
			err:   "",
		},
	}
	s := make(source)
	for i, c := range cases {
		converter := NewConverter(s, persistor{}, app{})
		s <- c.input
		m := <-converter.Output()
		if m.Error != c.err {
			t.Errorf("case %v: got %v, expected %v", i, m.Error, c.err)
		}
		close(s)
	}
}
