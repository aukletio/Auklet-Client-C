package agent

import (
	"bytes"
	"strings"
	"fmt"
	"testing"
)

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
	tests := []struct {
		input  string
		want Message
		problem bool
	}{
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
			problem: true,
		},
	}
	for _, test := range tests {
		s := newServer(strings.NewReader(test.input), nil)
		for s.scan() {
			got := s.msg
			problem := s.err != nil
			if problem != test.problem {
				t.Errorf("case %+v: problem = %v, error = %v", test, problem, s.err)
			}
			if !compare(got, test.want) {
				t.Errorf("expected %v, got %v", test.want, got)
			}
		}
	}
}
