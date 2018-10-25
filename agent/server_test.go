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
		},
		{
			input: []byte(`{"type":"event","data":"hello, world"}`),
			expect: Message{
				Type:  "event",
				Data:  []byte(`"hello, world"`),
				Error: "",
			},
		},
		{
			input: []byte(`{"malformed`),
			expect: Message{
				Type:  "log",
				Data:  []byte{},
				Error: `unexpected EOF in {"malformed`,
			},
		},
	}
	for _, c := range cases {
		s := NewServer(bytes.NewBuffer(c.input), nil)
		got := <-s.Output()
		if !compare(got, c.expect) {
			t.Errorf("expected %v, got %v", c.expect, got)
		}
	}
}
