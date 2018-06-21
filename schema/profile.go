package schema

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// NewProfile creates a Profile for app out of raw message data.
func NewProfile(data []byte, app *app.App) (m broker.Message, err error) {
	var p Profile
	err = json.Unmarshal(data, &p)
	if err != nil {
		// There was a problem decoding the raw message.
		return
	}
	p.PublicIP = device.CurrentIP()
	p.Id = uuid.NewV4().String()
	p.Timestamp = time.Now().UnixNano() / 1000000 // milliseconds
	p.Checksum = app.CheckSum
	p.Application = app.ID
	b, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return
	}
	return broker.StdPersistor.CreateMessage(b, broker.Profile)
}
