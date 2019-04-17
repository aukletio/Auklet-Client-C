package schema

import (
	"encoding/json"
	"testing"

	"github.com/aukletio/Auklet-Client-C/agent"
	"github.com/aukletio/Auklet-Client-C/broker"
	"github.com/aukletio/Auklet-Client-C/device"
)

type persistor struct{}

func (p persistor) CreateMessage(*broker.Message) error { return nil }

type source chan agent.Message

func (s source) Output() <-chan agent.Message { return s }

type app struct{}

func (app) CheckSum() string     { return "checksum" }
func (app) ExitStatus() int      { return 42 }
func (app) Signal() string       { return "something" }
func (app) AgentVersion() string { return "something" }

type monitor struct{}

func (monitor) GetMetrics() device.Metrics { return device.Metrics{} }
func (monitor) Close()                     {}

var cfg = Config{
	Monitor:     monitor{},
	Persistor:   persistor{},
	App:         app{},
	Username:    "username",
	UserVersion: "userVersion",
	AppID:       "app id",
	MacHash:     "mac hash",
}

func TestConverter(t *testing.T) {
	tests := []struct {
		input agent.Message
		ok    bool
	}{
		{input: agent.Message{Type: "event"}, ok: true},
		{input: agent.Message{Type: "profile"}, ok: true},
		{input: agent.Message{Type: "cleanExit"}, ok: true},
		{
			input: agent.Message{
				Type: "datapoint",
				Data: json.RawMessage(`{
					"type": "",
					"payload": {}
				}`),
			},
			ok: true,
		},
		{
			input: agent.Message{
				Type: "datapoint",
				Data: json.RawMessage(`{
					"type": "generic",
					"payload": {}
				}`),
			},
			ok: true,
		},
		{
			input: agent.Message{
				Type: "datapoint",
				Data: json.RawMessage(`{
					"type": "location",
					"payload": {
						"speed": 1.0,
						"longitude": 1.0,
						"latitude": 1.0,
						"altitude": 1.0,
						"course": 1.0,
						"timestamp": 10,
						"precision": 0.1
					}
				}`),
			},
			ok: true,
		},
		{
			input: agent.Message{
				Type: "datapoint",
				Data: json.RawMessage(`{
					"type": "motion",
					"payload": {
						"x_axis": 1.0,
						"y_axis": 1.0,
						"z_axis": 1.0
					}
				}`),
			},
			ok: true,
		},
		{input: agent.Message{Type: "unknown"}, ok: false},
	}
	for i, test := range tests {
		s := make(source)
		converter := NewConverter(cfg, s)
		s <- test.input
		m := <-converter.Output()
		ok := m.Error == ""
		if ok != test.ok {
			t.Errorf("case %v: got %v, expected %v: %v", i, ok, test.ok, m.Error)
		}
		close(s)
	}
}

func TestUnmarshalStrict(t *testing.T) {
	tests := []struct {
		input   string
		problem bool
	}{
		{
			input:   `{"bogus":0}`,
			problem: true,
		},
		{
			input: `{"number":0}`,
		},
	}
	for _, test := range tests {
		var v struct {
			Number int `json:"number"`
		}
		err := unmarshalStrict([]byte(test.input), &v)
		problem := err != nil
		if problem != test.problem {
			t.Errorf("case %+v: problem = %v", test, problem)
		}
	}
}

func TestDataPoint(t *testing.T) {
	c := newConverter(cfg)
	tests := []struct {
		input   string
		problem bool
	}{
		{
			input: `{
				"type": "",
				"payload": {}
			}`,
		},
		{
			input: `{
				"type": "",
				"bogus": {}
			}`,
			problem: true,
		},
		{
			input: `{
				"type": "bogus",
				"payload": {}
			}`,
			problem: true,
		},
		{
			input: `{
				"type": "location",
				"payload": {}
			}`,
		},
		{
			input: `{
				"type": "location",
				"payload": {"bogus":null}
			}`,
			problem: true,
		},
		{
			input: `{
				"type": "motion",
				"payload": {}
			}`,
		},
		{
			input: `{
				"type": "motion",
				"payload": {"bogus":null}
			}`,
			problem: true,
		},
	}
	for _, test := range tests {
		dp := c.dataPoint([]byte(test.input))
		problem := dp.Error != ""
		if problem != test.problem {
			t.Errorf("case %+v: problem = %v", test, problem)
			t.Errorf("case %+v: error = %v", test, dp.Error)
		}
	}
}
