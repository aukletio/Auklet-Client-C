package agent

import (
	"strings"
	"testing"
)

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
			problem := s.err != nil
			if problem != test.problem {
				t.Errorf("case %+v: problem = %v, error = %v", test, problem, s.err)
			}
		}
	}
}
