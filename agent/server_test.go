package agent

import (
	"bytes"
	"strings"
	"fmt"
	"testing"
)

type testServerCase struct {
	input  string
	want Message
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
	tests := []testServerCase{
		{
			input: `{"type":"message","data":"hello, world"}`,
			want: Message{
				Type:  "message",
				Data:  []byte(`"hello, world"`),
				Error: "",
			},
		},
		{
			input: `{"type":"event","data":"hello, world"}`,
			want: Message{
				Type:  "event",
				Data:  []byte(`"hello, world"`),
				Error: "",
			},
		},
		{
			input: `{"malformed`,
			want: Message{
				Type:  "log",
				Data:  []byte{},
				Error: `unexpected EOF in {"malformed`,
			},
		},
	}
	for _, test := range tests {
		s := newServer(strings.NewReader(test.input), nil)
		for s.scan() {
			got := s.msg
			if !compare(got, test.want) {
				t.Errorf("expected %v, got %v", test.want, got)
			}
		}
	}
}
