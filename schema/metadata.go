package schema

import (
	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/device"
	"github.com/ESG-USA/Auklet-Client-C/version"
)

type metadata struct {
	Version      string `json:"clientVersion"`
	AgentVersion string `json:"agentVersion"`
	AppID        string `json:"application"`
	CheckSum     string `json:"checksum"`  // SHA512/224 hash of the executable
	IP           string `json:"publicIP"`  // current public IP address
	UUID         string `json:"id"`        // identifier for this message
	Time         int64  `json:"timestamp"` // Unix milliseconds
	Error        string `json:"error,omitempty"`
}

// App is anything that can return a checksum and agent version.
type App interface {
	AgentVersion() string
	CheckSum() string
}

func newMetadata(app App) metadata {
	return metadata{
		Version:      version.Version,
		AgentVersion: app.AgentVersion(),
		AppID:        config.AppID(),
		CheckSum:     app.CheckSum(),
		IP:           device.CurrentIP(),
		UUID:         uuid.NewV4().String(),
		Time:         nowMilli(),
	}
}
