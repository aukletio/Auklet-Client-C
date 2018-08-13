// Package api implements an interface to the Auklet backend API.
package api

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/device"
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

var errParseCA = errors.New("failed to parse CA")

// tlsConfig converts ca into a *tls.Config.
func tlsConfig(ca []byte) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(ca) {
		return nil, errParseCA
	}
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: false,
	}, nil
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
	ca, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorlog.Print(err)
		return
	}
	c, err = tlsConfig(ca)
	if err != nil {
		errorlog.Print(err)
	}
	return
}

// CreateOrGetDevice associates machash and appid in the backend.
func CreateOrGetDevice() {
	b, _ := json.Marshal(struct {
		Mac   string `json:"mac_address_hash"`
		AppID string `json:"application"`
	}{
		Mac:   device.MacHash,
		AppID: config.AppID(),
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

// GetBrokerAddr returns a broker from the config endpoint.
func GetBrokerAddr() string {
	resp := get(configEP, "application/json")
	if resp == nil {
		return ""
	}
	if resp.StatusCode != 200 {
		errorlog.Printf("api.Config: unexpected status %v", resp.Status)
		return ""
	}

	var k struct {
		Broker string `json:"brokers"`
		Port   string `json:"port"`
	}

	d := json.NewDecoder(resp.Body)
	if err := d.Decode(&k); err != nil && err != io.EOF {
		errorlog.Print(err)
	}

	return fmt.Sprintf("ssl://%s:%s", k.Broker, k.Port)
}
