package agent

import (
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	tests := []struct {
		input   string
		problem bool
	}{
		{
			input: `{"type":"message","data":"hello, world"}`,
		},
		{
			input: `{"type":"event","data":"hello, world"}`,
		},
		{
			input:   `{"malformed`,
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
