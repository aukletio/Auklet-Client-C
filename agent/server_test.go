package agent

import (
	"bytes"
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

func TestServer(t *testing.T) {
	conn := bytes.NewBuffer([]byte(`{"type":"message","data":"hello, world"}`))
	handlers := map[string]Handler{
		"message": func(data []byte) (_ broker.Message, _ error) {
			d := string(data)
			exp := `"hello, world"`
			if d != exp {
				t.Errorf("handler: expected %v, got %v", exp, d)
			}
			return
		},
	}
	for _ = range NewServer(conn, handlers).Output() {
	}
}
