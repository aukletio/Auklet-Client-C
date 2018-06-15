// Package api implements an interface to the Auklet backend API.
package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ESG-USA/Auklet-Client-C/certs"
	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// namespaces and endpoints for the API. All new endpoints should be entered
// here.
const (
	releasesEP     = "/private/releases/?checksum="
	certificatesEP = "/private/devices/certificates/"
	devicesEP      = "/private/devices/"
	configEP       = "/private/devices/config/"
	dataLimitEP    = "/private/devices/%v/app_config/" // AppID
)

// BaseURL is the subdomain against which requests will be performed. It
// should not assume any particular namespace.
var BaseURL string

func get(args, contenttype string) (resp *http.Response) {
	url := BaseURL + args
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errorlog.Print(err)
		return
	}
	req.Header.Add("Authorization", "JWT "+config.APIKey())
	if contenttype != "" {
		req.Header.Add("content-type", contenttype)
	}
	c := &http.Client{}
	resp, err = c.Do(req)
	if err != nil {
		errorlog.Print(err)
		return
	}
	if resp.StatusCode != 200 {
		errorlog.Printf("api.get: got unexpected status %v from %v", resp.Status, url)
	}
	return
}

// Release returns true if checksum represents an app that has been released.
func Release(checksum string) (ok bool) {
	resp := get(releasesEP+checksum, "")
	if resp == nil {
		return
	}
	switch resp.StatusCode {
	case 200:
		ok = true
	case 404:
		log.Printf("not released: %v", checksum)
		ok = false
	default:
		errorlog.Printf("api.Release: got unexpected status %v", resp.Status)
	}
	return
}

// Certificates retrieves SSL certificates.
func Certificates() (c *tls.Config) {
	resp := get(certificatesEP, "")
	if resp == nil {
		return
	}
	if resp.StatusCode != 200 {
		errorlog.Printf("api.Certificates: unexpected status %v", resp.Status)
		return
	}
	cts, err := certs.Unpack(resp.Body)
	if err != nil {
		errorlog.Print(err)
		return
	}
	return cts.TLSConfig()
}

// CreateOrGetDevice associates machash and appid in the backend.
func CreateOrGetDevice(machash, appid string) {
	b, _ := json.Marshal(struct {
		Mac   string `json:"mac_address_hash"`
		AppID string `json:"application"`
	}{
		Mac:   machash,
		AppID: appid,
	})
	url := BaseURL + devicesEP
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		errorlog.Print(err)
		return
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "JWT "+config.APIKey())

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errorlog.Print(err)
		return
	}
	log.Printf("api.CreateOrGetDevice: got response status %v", resp.Status)
}

// BrokerParams represents parameters affecting broker communication.
type BrokerParams struct {
	// Brokers is a list of broker addresses.
	Brokers []string `json:"brokers"`

	// LogTopic, ProfileTopic, and EventTopic are topics to which we produce
	// broker messages.
	LogTopic     string `json:"log_topic"`
	ProfileTopic string `json:"prof_topic"`
	EventTopic   string `json:"event_topic"`
}

// GetBrokerParams returns broker parameters from the config endpoint.
func GetBrokerParams() (k BrokerParams) {
	resp := get(configEP, "application/json")
	if resp == nil {
		return
	}
	if resp.StatusCode != 200 {
		errorlog.Printf("api.Config: unexpected status %v", resp.Status)
		return
	}
	d := json.NewDecoder(resp.Body)
	err := d.Decode(&k)
	if err != nil && err != io.EOF {
		errorlog.Print(err)
	}
	return
}
