package schema

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/device"
)

// Profile represents profile data as expected by Kafka consumers.
type Profile struct {
	// AppID is a long string uniquely associated with a particular app.
	AppID string `json:"app_id"`

	// CheckSum is the SHA512/224 hash of the executable, used to associate
	// tree data with a particular release.
	CheckSum string `json:"checksum"`

	// IP is the public IP address of the device on which we are running,
	// used to associate tree data with an estimated geographic location.
	IP string `json:"public_ip"`

	// UUID is a unique identifier for a particular tree.
	UUID string `json:"uuid"`

	// Time is the Unix epoch time (in milliseconds) at which a tree was
	// received.
	Time int64 `json:"timestamp"`

	// Tree represents the profile tree data generated by libauklet.
	Tree json.RawMessage `json:"tree"`

	kafkaTopic string
}

// NewProfile creates a Profile for app out of JSON data.
func NewProfile(data []byte, app *app.App, topic string) (p Profile, err error) {
	err = json.Unmarshal(data, &p)
	if err != nil {
		// There was a problem decoding the JSON.
		return
	}
	p.IP = device.CurrentIP()
	p.UUID = uuid.NewV4().String()
	p.Time = time.Now().UnixNano() / 1000000 // milliseconds
	p.CheckSum = app.CheckSum
	p.AppID = app.AppID
	p.kafkaTopic = topic
	return
}

// Topic returns the Kafka topic to which p should be sent.
func (p Profile) Topic() string {
	return p.kafkaTopic
}

// Bytes returns the Profile as a byte slice.
func (p Profile) Bytes() ([]byte, error) {
	return json.MarshalIndent(p, "", "\t")
}
