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
	"net/http"

	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/device"
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

func get(args, contenttype string) (*http.Response, error) {
	url := BaseURL + args
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "JWT "+config.APIKey())
	if contenttype != "" {
		req.Header.Add("content-type", contenttype)
	}

	return http.DefaultClient.Do(req)
}

type ErrNotReleased string

func (err ErrNotReleased) Error() string {
	return fmt.Sprintf("not released: %v", string(err))
}

// Release returns nil if checksum represents an app that has been released.
//
// There are two classes of errors:
//
// 1. The HTTP GET request failed.
// 2. The response's status code is not 200.
//
func Release(checksum string) error {
	resp, err := get(releasesEP+checksum, "")
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return ErrNotReleased(checksum)
	}

	return nil
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

type ErrStatus struct {
	resp *http.Response
}

func (err ErrStatus) Error() string {
	return fmt.Sprintf("unexpected status: %v from %v", err.resp.Status, err.resp.Request.URL)
}

// Certificates retrieves SSL certificates.
func Certificates() (*tls.Config, error) {
	resp, err := get(certificatesEP, "")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, ErrStatus{resp}
	}
	ca, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return tlsConfig(ca)
}

// CreateOrGetDevice associates machash and appid in the backend.
func CreateOrGetDevice() (func() (string, string), error) {
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
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "JWT "+config.APIKey())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, ErrStatus{resp}
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var v struct {
		Username string `json:"id"`
		Password string `json:"client_password"`
	}

	if err := json.Unmarshal(body, &v); err != nil {
		return nil, ErrEncoding{err, string(body)}
	}

	fmt.Printf("%#v\n", v)
	return func() (string, string) {
		return v.Username, v.Password
	}, nil
}

type ErrEncoding struct {
	Err  error
	What string
}

func (err ErrEncoding) Error() string {
	return fmt.Sprintf("encoding error: %v in %v", err.Err, err.What)
}

// GetBrokerAddr returns a broker from the config endpoint.
func GetBrokerAddr() (string, error) {
	resp, err := get(configEP, "application/json")
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", ErrStatus{resp}
	}

	var k struct {
		Broker string `json:"brokers"`
		Port   string `json:"port"`
	}

	d := json.NewDecoder(resp.Body)
	if err := d.Decode(&k); err != nil && err != io.EOF {
		b, _ := ioutil.ReadAll(d.Buffered())
		return "", ErrEncoding{
			Err:  err,
			What: string(b),
		}
	}

	return fmt.Sprintf("ssl://%s:%s", k.Broker, k.Port), nil
}
