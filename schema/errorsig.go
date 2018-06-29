// +build linux

package schema

import (
	"encoding/json"
	"syscall"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// NewErrorSig creates an Exit for app out of raw message data. It assumes
// that app.Wait() has returned.
func NewErrorSig(data []byte, app *app.App) (m broker.Message, err error) {
	var e Exit
	err = json.Unmarshal(data, &e)
	if err != nil {
		return
	}
	e.Application = &app.ID
	e.Checksum = &app.CheckSum
	e.PublicIP = device.CurrentIP()
	id := uuid.NewV4().String()
	e.Id = &id
	t := time.Now().String()
	e.Timestamp = &t
	es := int32(app.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
	e.ExitStatus = &es
	e.MacAddressHash = &device.MacHash
	e.SystemMetrics = device.GetMetrics()
	return broker.StdPersistor.CreateMessage(e, broker.Event)
}
