package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/vmihailenco/msgpack"

	backend "github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/message"
)

type mockExec struct {
	checksum     string
	agentVersion string
	appLogs      io.Reader
	agentData    io.ReadWriter
	decoder      *json.Decoder
}

func newMockExec() *mockExec {
	return &mockExec{
		agentVersion: "agentVersion",
		checksum:     "checksum",
		appLogs:      strings.NewReader("appLogs\n"),
		agentData:    bytes.NewBufferString(`{"type":"profile","data":{}}`),
		decoder:      nil, // dynamically initialized
	}
}

func (m mockExec) CheckSum() string         { return m.checksum }
func (mockExec) Run() error                 { return nil }
func (mockExec) Connect() error             { return nil }
func (m mockExec) AgentData() io.ReadWriter { return m.agentData }
func (m mockExec) AppLogs() io.Reader       { return m.appLogs }
func (m mockExec) AgentVersion() string     { return m.agentVersion }
func (m *mockExec) Decoder() *json.Decoder {
	if m.decoder == nil {
		m.decoder = json.NewDecoder(m.agentData)
	}
	return m.decoder
}
func (mockExec) ExitStatus() int { return 0 }
func (mockExec) Signal() string  { return "signal" }

type mockAPI struct {
	checksum  string
	dataLimit backend.DataLimit
}

func (m mockAPI) Release(s string) error {
	if s != m.checksum {
		return errors.New("not released")
	}
	return nil
}

func (m mockAPI) DataLimit() (*backend.DataLimit, error) {
	return &m.dataLimit, nil
}

type mockProducer struct{}

func dump(in []byte) {
	var v interface{}
	err := msgpack.Unmarshal(in, &v)
	if err != nil {
		panic(err)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func (p *mockProducer) Serve(src broker.MessageSource) {
	for m := range src.Output() {
		dump(m.Bytes)
		m.Remove()
	}
}

func TestClient(t *testing.T) {
	e := newMockExec()

	c := client{
		msgPath:      ".auklet/message",
		limPersistor: &message.MemPersistor{},
		api: mockAPI{
			checksum: "checksum",
			dataLimit: backend.DataLimit{
				EmissionPeriod: 1,
				Cellular: backend.CellularConfig{
					Date:    1,
					Defined: false,
					Limit:   0,
				},
			},
		},
		userVersion: "userVersion",
		username:    "username",
		appID:       "appID",
		macHash:     "macHash",
		producer:    &mockProducer{},
		fs:          afero.NewMemMapFs(),
	}

	if err := c.run(e); err != nil {
		t.Error(err)
	}
}

func TestDumper(t *testing.T) {
	e := newMockExec()

	var d dumper
	if err := d.run(e); err != nil {
		t.Error(err)
	}
}

func TestSerial(t *testing.T) {
	e := newMockExec()
	addr := "serial-device"
	s := serial{
		userVersion: "userVersion",
		appID:       "appID",
		macHash:     "macHash",
		addr:        addr,
		fs:          afero.NewMemMapFs(),
	}
	if err := s.run(e); err != nil {
		t.Error(err)
	}
	f, err := s.fs.Open(addr)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	line := bufio.NewScanner(f)
	for line.Scan() {
		var v interface{}
		if err := json.Unmarshal(line.Bytes(), &v); err != nil {
			t.Error()
		}
		m := v.(map[string]interface{})
		if m["topic"] == "" {
			t.Fail()
		}
		if m["payload"] == nil {
			t.Fail()
		}
	}
}
