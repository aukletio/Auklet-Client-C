package agent

import (
	"bytes"
	"fmt"
	"testing"
)

type testServerCase struct {
	input  []byte
	expect Message
}

func compare(a, b Message) bool {
	return a.Type == b.Type && bytes.Compare(a.Data, b.Data) == 0 && a.Error == b.Error
}

func (m Message) String() string {
	return fmt.Sprint(struct {
		Type, Data, Error string
	}{
		Type:  m.Type,
		Data:  string(m.Data),
		Error: m.Error,
	})
}

func TestServer(t *testing.T) {
	cases := []testServerCase{
		{
			input: []byte(`{"type":"message","data":"hello, world"}`),
			expect: Message{
				Type:  "message",
				Data:  []byte(`"hello, world"`),
				Error: "",
			},
		}, {
			input: []byte(`{"malformed`),
			expect: Message{
				Type:  "",
				Data:  []byte{},
				Error: `unexpected EOF in {"malformed`,
			},
		},
	}
	for _, c := range cases {
		s := NewServer(bytes.NewBuffer(c.input))
		got := <-s.Output()
		if !compare(got, c.expect) {
			t.Errorf("expected %v, got %v", c.expect, got)
		}
	}
}

func TestServerEOF(t *testing.T) {
	s := NewServer(bytes.NewBuffer([]byte("{}")))
	<-s.Output()
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
