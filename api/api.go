// Package api implements an interface to the Auklet backend API.
package api

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// namespaces and endpoints for the API. All new endpoints should be entered
// here.
var (
	releasesEP     = "/private/releases/"
	certificatesEP = "/private/devices/certificates/"
	devicesEP      = "/private/devices/"
	configEP       = "/private/devices/config/"
	dataLimitEP    = "/private/devices/" + config.AppID() + "/app_config/"
)

// BaseURL is the subdomain against which requests will be performed. It
// should not assume any particular namespace.
var BaseURL string

type Call interface {
	Request() *http.Request
	Handle(*http.Response) error
}

func Do(c Call) error {
	req := c.Request()
	req.Header.Add("Authorization", "JWT "+config.APIKey())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return c.Handle(resp)
}

type errStatus struct {
	resp *http.Response
}

func (err errStatus) Error() string {
	return fmt.Sprintf("unexpected status: %v from %v", err.resp.Status, err.resp.Request.URL)
}

// Credentials represents credentials required for sending broker messages.
type Credentials struct {
	Username string `json:"id"`
	Password string `json:"client_password"`
	Org      string `json:"organization"`
	ClientID string `json:"client_id"`
}

// GetCredentials retrieves credentials from the filesystem or API, whichever is
// available.
func GetCredentials(path string) (*Credentials, error) {
	creds, err := credsFromFile(path)
	if err != nil {
		// file doesn't exist; ask the API for credentials
		return getAndSaveCredentials(path)
	}
	return creds, nil
}

func credsFromFile(path string) (*Credentials, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// decrypt here
	creds := new(Credentials)
	if err := json.Unmarshal(b, creds); err != nil {
		return nil, errEncoding{err, string(b), "credsFromFile"}
	}
	return creds, nil
}

// getAndSaveCredentials requests credentials from the API. If it receives them,
// it writes them to the given path.
func getAndSaveCredentials(path string) (*Credentials, error) {
	creds := new(Credentials)
	if err := Do(creds); err != nil {
		return nil, err
	}
	b, _ := json.Marshal(creds)
	// encrypt here
	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		return nil, err
	}
	return creds, nil
}

// decodeCredentials unmarshals data into Credentials. If the password is empty,
// it returns an error.
//
// The API returns an empty password if a device's credentials have been
// requested more than once.
func (creds *Credentials) decodeCredentials(data []byte) error {
	if err := json.Unmarshal(data, creds); err != nil {
		return errEncoding{err, string(data), "decodeCredentials"}
	}

	if creds.Password == "" {
		return errors.New("empty password")
	}

	return nil
}

type errEncoding struct {
	Err  error
	What string
	Op   string
}

func (Credentials) Request() *http.Request {
	b, _ := json.Marshal(struct {
		Mac   string `json:"mac_address_hash"`
		AppID string `json:"application"`
	}{
		// device info
		Mac:   device.MacHash,
		AppID: config.AppID(),
	})

	url := BaseURL + devicesEP
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Add("content-type", "application/json")
	return req
}

func (c *Credentials) Handle(resp *http.Response) error {
	if resp.StatusCode != 201 {
		return errStatus{resp}
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return c.decodeCredentials(body)
}

func (err errEncoding) Error() string {
	return fmt.Sprintf("encoding error during %v: %v in %v", err.Op, err.Err, err.What)
}

type Release struct {
	CheckSum string
}

func (r Release) Request() *http.Request {
	url := BaseURL + releasesEP + "?checksum=" + r.CheckSum
	req, _ := http.NewRequest("GET", url, nil)
	return req
}

func (r Release) Handle(resp *http.Response) error {
	if resp.StatusCode != 200 {
		return errNotReleased(r.CheckSum)
	}
	return nil
}

type errNotReleased string

func (err errNotReleased) Error() string {
	return fmt.Sprintf("not released: %v", string(err))
}

type Certificates struct {
	TLSConfig *tls.Config
}

func (Certificates) Request() *http.Request {
	req, _ := http.NewRequest("GET", BaseURL+certificatesEP, nil)
	return req
}

func (c *Certificates) Handle(resp *http.Response) (err error) {
	if resp.StatusCode != 200 {
		return errStatus{resp}
	}
	ca, _ := ioutil.ReadAll(resp.Body)
	c.TLSConfig, err = tlsConfig(ca)
	return
}

var errParseCA = errors.New("failed to parse CA")

var appendCertsFromPEM = func(certpool *x509.CertPool, ca []byte) bool {
	return certpool.AppendCertsFromPEM(ca)
}

// tlsConfig converts ca into a *tls.Config.
func tlsConfig(ca []byte) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	if !appendCertsFromPEM(certpool, ca) {
		return nil, errParseCA
	}
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: false,
	}, nil
}

type BrokerAddress struct {
	Address string
}

func (BrokerAddress) Request() *http.Request {
	req, _ := http.NewRequest("GET", BaseURL+configEP, nil)
	req.Header.Add("content-type", "application/json")
	return req
}

func (b *BrokerAddress) Handle(r *http.Response) error {
	if r.StatusCode != 200 {
		return errStatus{r}
	}

	var k struct {
		Broker string `json:"brokers"`
		Port   string `json:"port"`
	}

	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, &k); err != nil {
		return errEncoding{err, string(body), "GetBrokerAddr"}
	}

	b.Address = fmt.Sprintf("ssl://%s:%s", k.Broker, k.Port)
	return nil
}

type CellularConfig struct {
	Date int // day of the month in [1,28]

	Defined bool
	Limit   int
}

type DataLimit struct {
	EmissionPeriod int
	Cellular       CellularConfig
}

func (DataLimit) Request() *http.Request {
	req, _ := http.NewRequest("GET", BaseURL+dataLimitEP, nil)
	req.Header.Add("content-type", "application/json")
	return req
}

func (d *DataLimit) Handle(r *http.Response) (err error) {
	if r.StatusCode != 200 {
		return errStatus{r}
	}

	type storage struct {
		Limit *int `json:"storage_limit"`
	}

	type data struct {
		Limit *int `json:"cellular_data_limit"`
		Date  int  `json:"normalized_cell_plan_date"`
	}

	type config struct {
		EmissionPeriod int     `json:"emission_period"`
		Storage        storage `json:"storage"`
		Data           data    `json:"data"`
	}

	var l struct {
		Config config `json:"config"`
	}

	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, &l); err != nil {
		return errEncoding{err, string(body), "GetDataLimit"}
	}

	c := l.Config
	d.EmissionPeriod = c.EmissionPeriod
	d.Cellular.Date = c.Data.Date
	d.Cellular.Defined = c.Data.Limit != nil
	if d.Cellular.Defined {
		d.Cellular.Limit = *c.Data.Limit
	}

	return nil
}
