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
	id := uuid.NewV4().String()
	p.Id = &id
	t := time.Now().UnixNano() / 1000000 // milliseconds
	p.Timestamp = &t
	p.Checksum = &app.CheckSum
	p.Application = &app.ID
	return broker.StdPersistor.CreateMessage(p, broker.Profile)
}
