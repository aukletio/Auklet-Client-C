package agent

import (
	"testing"
	"strings"
)

func TestDataPointServer(t *testing.T) {
	data := `
	{ "type":        "", "data": { "arbitrary": "json" } }
	{ "type": "generic", "data": { "arbitrary": "json" } }
	{
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
	}
	{
		"type": "motion",
		"data": {
			"x_axis": 1.0,
			"y_axis": 1.0,
			"z_axis": 1.0
		}
	}`
	server := NewDataPointServer(strings.NewReader(data))
	for msg := range server.Output() {
		if msg.Error != "" {
			t.Error(msg.Error)
		}
	}
}
