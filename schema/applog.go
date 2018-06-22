package schema

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// NewAppLog converts msg into a custom log message.
func NewAppLog(msg []byte, app *app.App) (m broker.Message, err error) {
	var a AppLog
	a.Application = app.ID
	a.Checksum = app.CheckSum
	a.PublicIP = device.CurrentIP()
	a.Id = uuid.NewV4().String()
	a.Timestamp = time.Now().String()
	a.MacAddressHash = device.MacHash
	a.SystemMetrics = device.GetMetrics()
	a.Message = msg
	b, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return
	}
	return broker.StdPersistor.CreateMessage(b, broker.Log)
}
