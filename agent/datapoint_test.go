package agent

import (
	"strings"
	"testing"
)

func TestDataPointServer(t *testing.T) {
	tests := []struct {
		data    string
		problem bool
	}{
		{
			data: `{
				"type": "",
				"data": {
					"arbitrary": "json"
				}
			}`,
		},
		{
			data: `{
				"type": "generic",
				"data": {
					"arbitrary": "json"
				}
			}`,
		},
		{
			data: `{
				"type": "location",
				"data": {
					"speed":      1.0,
					"longitude":  1.0,
					"latitude":   1.0,
					"altitude":   1.0,
					"course":     1.0,
					"timestamp":   10,
					"precision":  0.1
				}
			}`,
		},
		{
			data: `{
				"type": "motion",
				"data": {
					"x_axis": 1.0,
					"y_axis": 1.0,
					"z_axis": 1.0
				}
			}`,
		},
		{
			data:    `}`,
			problem: true,
		},
	}
	for _, test := range tests {
		server := newDataPointServer(strings.NewReader(test.data))
		go server.serve()
		for msg := range server.Output() {
			problem := msg.Error != ""
			if problem != test.problem {
				t.Errorf("case %+v: problem = %v, error = %v", test, problem, msg.Error)
			}
		}
	}
}
