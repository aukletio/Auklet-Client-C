package agent

import (
	"strings"
	"testing"
)

func TestDataPointServer(t *testing.T) {
	tests := []struct {
		data string
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
	}
	for _, test := range tests {
		server := NewDataPointServer(strings.NewReader(test.data))
		for msg := range server.Output() {
			problem := msg.Error != ""
			if problem != test.problem {
				t.Error(msg.Error)
			}
		}
	}
}
