// Package api implements an interface to the Auklet backend API.
package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ESG-USA/Auklet-Client/certs"
)

// namespaces and endpoints for the API. All new endpoints should be entered
// here.
const (
	releases     = "/private/releases/?checksum="
	certificates = "/private/devices/certificates/"
	devices      = "/private/devices/"
)

// An API represents parameters common to all API requests.
type API struct {
	*http.Client

	// BaseURL is the subdomain against which requests will be performed. It
	// should not assume any particular namespace.
	BaseURL string

	// Key is the API key provided by package config.
	Key string
}

// New creates an API with the given parameters.
func New(baseurl, key string) API {
	return API{
		Client:  &http.Client{},
		BaseURL: baseurl,
		Key:     key,
	}
}

func (api API) get(args string) (resp *http.Response) {
	url := api.BaseURL + args
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print(err)
		return
	}
	req.Header.Add("Authorization", "JWT "+api.Key)
	resp, err = api.Do(req)
	if err != nil {
		log.Print(err)
	}
	if resp.StatusCode != 200 {
		log.Printf("api.get: %v from %v", resp.Status, url)
	}
	return
}

// Release returns true if checksum represents an app that has been released.
func (api API) Release(checksum string) (ok bool) {
	resp := api.get(releases + checksum)
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
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("api.Release: got unexpected status %v", resp.Status)
		log.Print(string(b))
	}
	return
}

// Certificates retrieves SSL certificates.
func (api API) Certificates() (c *tls.Config) {
	resp := api.get(certificates)
	if resp == nil {
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("api.Certificates: unexpected status %v", resp.Status)
		return
	}
	cts, err := certs.Unpack(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}
	return cts.TLSConfig()
}

type deviceJSON struct {
	Mac   string `json:"mac_address_hash"`
	AppID string `json:"application"`
}

// CreateOrGetDevice associates machash and appid in the backend.
func (api API) CreateOrGetDevice(machash, appid string) {
	b, _ := json.Marshal(deviceJSON{
		Mac:   machash,
		AppID: appid,
	})
	url := api.BaseURL + devices
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		log.Print(err)
		return
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "JWT "+api.Key)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("api.CreateOrGetDevice: got response status %v", resp.Status)
}