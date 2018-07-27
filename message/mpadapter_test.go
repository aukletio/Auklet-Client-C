package message

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/schema"
)

// TestAdapt tests that the adapt function can process all of the schemas we
// expect it to receive.
func TestAdapt(t *testing.T) {
	bogusAddr := "" // disable persistence
	p := broker.NewPersistor(bogusAddr)
	for _, v := range []interface{}{
		[]byte("hello, world"),
		schema.AppLog{},
		schema.ErrorSig{},
		schema.Exit{},
		schema.Profile{},
		json.RawMessage(`{"hello":"world"}`),
	} {
		m, err := p.CreateMessage(v, 0)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(m.Bytes))
		if err = adapt(&m); err != nil {
			t.Error(err)
		}
	}
}
