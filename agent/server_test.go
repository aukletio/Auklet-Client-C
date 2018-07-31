package agent

import (
	"bytes"
	"testing"
)

func TestServer(t *testing.T) {
	conn := bytes.NewBuffer([]byte(`{"type":"message","data":"hello, world"}`))
	s := NewServer(conn)
	m := <-s.Output()
	switch m.Type {
	case "message":
		d := string(m.Data)
		exp := `"hello, world"`
		if d != exp {
			t.Errorf("expected %v, got %v", exp, d)
		}
	default:
		t.Errorf(`expected "message", got %q`, m.Type)
	}
	// output channel ought to close now
	select {
	case _, open := <-s.Output():
		if open {
			t.Errorf("expected channel to close, but it's open and still has messages")
		} else {
			// success
		}
	default:
		t.Errorf("expected channel to close immediately, but it's open and blocking")
	}
}
