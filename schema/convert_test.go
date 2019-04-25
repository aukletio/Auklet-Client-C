package schema

import (
	"encoding/json"
	"math"
	"math/big"
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

var converterTests = []struct {
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

func TestConverter(t *testing.T) {
	for i, test := range converterTests {
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

var dataPointTests = []struct {
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

func TestDataPoint(t *testing.T) {
	c := newConverter(cfg)
	for _, test := range dataPointTests {
		dp := c.dataPoint([]byte(test.input))
		problem := dp.Error != ""
		if problem != test.problem {
			t.Errorf("case %+v: problem = %v", test, problem)
			t.Errorf("case %+v: error = %v", test, dp.Error)
		}
	}
}

type numberTest struct {
	number string // JSON number literal
	want   string // msgpack ext encoding
}

func exp(base, pow int64) *big.Int {
	return big.NewInt(0).
		Exp(
			big.NewInt(base),
			big.NewInt(pow),
			nil, // not modular exponentiation
		)
}

func header(length int) string {
	cases := []struct {
		max uint64
		tag string
	}{
		{max: math.MaxUint8, tag: "\xc7"},
		{max: math.MaxUint16, tag: "\xc8"},
		{max: math.MaxUint32, tag: "\xc9"},
	}

	for _, cas := range cases {
		if uint64(length) < cas.max {
			return cas.tag
		}
	}
	panic("too big to encode")
}

func from(data string) numberTest {
	n := len(data)
	return numberTest{
		number: data,
		want: header(n) +
			string(n) +
			"\x00" + // type
			data,
	}
}

var numberTests = []numberTest{
	{
		number: "1",
		want: "\xd4" + // fixext1 header
			"\x00" + // type field = 0
			"1", // data
	},
	{
		number: "10",
		want: "\xd5" + // fixext2 header
			"\x00" + // type field = 0
			"10", // data
	},
	{
		number: "210",
		want: "\xc7" + // ext8 header
			"\x03" + // length
			"\x00" + // type field = 0
			"210", // data
	},
	{
		number: "3210",
		want: "\xd6" + // fixext4 header
			"\x00" + // type field = 0
			"3210", // data
	},
	{
		number: "43210",
		want: "\xc7" + // ext8 header
			"\x05" + // length
			"\x00" + // type field = 0
			"43210", // data
	},
	{
		number: "543210",
		want: "\xc7" + // ext8 header
			"\x06" + // length
			"\x00" + // type field = 0
			"543210", // data
	},
	from(exp(2, 100).String()),
	from(exp(2, 200).String()),
	from(exp(2, 300).String()),
	from(exp(2, 400).String()),
}

func TestNumber(t *testing.T) {
	for _, test := range numberTests {
		var v interface{}
		if err := unmarshalStrict([]byte(test.number), &v); err != nil {
			t.Errorf("case %+v: %v", test, err)
		}
		b, err := msgpackMarshal(v)
		if err != nil {
			t.Errorf("case %+v: %v", test, err)
		}
		got := string(b)
		if got != test.want {
			t.Errorf("case %+v: got %v", test, got)
		}
	}
}
