package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/vmihailenco/msgpack"

	backend "github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type mockExec struct {
	checksum     string
	agentVersion string
	appLogs      io.Reader
	agentData    io.ReadWriter
	decoder      *json.Decoder
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
	limPath := "testdata/w/datalimit.json"
	msgPath := "testdata/w/message"
	defer func() {
		os.Remove(limPath)
		os.Remove(msgPath)
	}()
	c := client{
		msgPath: msgPath,
		limPath: limPath,
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
		exec: &mockExec{
			agentVersion: "agentVersion",
			checksum:     "checksum",
			appLogs:      strings.NewReader("appLogs\n"),
			agentData:    bytes.NewBufferString(`{"type":"profile","data":{}}`),
			decoder:      nil, // dynamically initialized
		},
		userVersion: "userVersion",
		username:    "username",
		appID:       "appID",
		macHash:     "macHash",
		producer:    &mockProducer{},
	}
	if err := c.run(); err != nil {
		t.Error(err)
	}
}
