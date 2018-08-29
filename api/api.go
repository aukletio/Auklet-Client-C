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
	releasesEP     = "/private/releases/?checksum="
	certificatesEP = "/private/devices/certificates/"
	devicesEP      = "/private/devices/"
	configEP       = "/private/devices/config/"
	dataLimitEP    = "/private/devices/"+config.AppID()+"/app_config/"
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

	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		return nil, errStatus{resp}
	}

	return resp, nil
}

type errNotReleased string

func (err errNotReleased) Error() string {
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
	_, err := get(releasesEP+checksum, "")
	return err
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

type errStatus struct {
	resp *http.Response
}

func (err errStatus) Error() string {
	return fmt.Sprintf("unexpected status: %v from %v", err.resp.Status, err.resp.Request.URL)
}

// Certificates retrieves SSL certificates.
func Certificates() (*tls.Config, error) {
	resp, err := get(certificatesEP, "")
	if err != nil {
		return nil, err
	}

	ca, _ := ioutil.ReadAll(resp.Body)
	return tlsConfig(ca)
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
func GetCredentials() (*Credentials, error) {
	path := ".auklet/identification"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		// file doesn't exist; ask the API for credentials
		return getAndSaveCredentials(path)
	}
	// decrypt here
	var creds Credentials
	if err := json.Unmarshal(b, &creds); err != nil {
		return nil, errEncoding{err, string(b), "GetCredentials"}
	}
	return &creds, nil
}

// getAndSaveCredentials requests credentials from the API. If it receives them,
// it writes them to the given path.
func getAndSaveCredentials(path string) (*Credentials, error) {
	creds, err := createOrGetDevice()
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(creds)
	// encrypt here
	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		return nil, err
	}
	return creds, nil
}

// createOrGetDevice requests credentials for this device from the API.
func createOrGetDevice() (*Credentials, error) {
	b, _ := json.Marshal(struct {
		Mac   string `json:"mac_address_hash"`
		AppID string `json:"application"`
	}{
		// device info
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
		return nil, errStatus{resp}
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return decodeCredentials(body)
}

// decodeCredentials unmarshals data into Credentials. If the password is empty,
// it returns an error.
//
// The API returns an empty password if a device's credentials have been
// requested more than once.
func decodeCredentials(data []byte) (*Credentials, error) {
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, errEncoding{err, string(data), "decodeCredentials"}
	}

	if creds.Password == "" {
		return nil, errors.New("empty password")
	}

	return &creds, nil
}

type errEncoding struct {
	Err  error
	What string
	Op   string
}

func (err errEncoding) Error() string {
	return fmt.Sprintf("encoding error during %v: %v in %v", err.Op, err.Err, err.What)
}

// GetBrokerAddr returns a broker from the config endpoint.
func GetBrokerAddr() (string, error) {
	resp, err := get(configEP, "application/json")
	if err != nil {
		return "", err
	}

	var k struct {
		Broker string `json:"brokers"`
		Port   string `json:"port"`
	}

	b, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &k); err != nil {
		return "", errEncoding{err, string(b), "GetBrokerAddr"}
	}

	return fmt.Sprintf("ssl://%s:%s", k.Broker, k.Port), nil
}

// CellularConfig defines a limit and date for devices that use a cellular
// connection.
type CellularConfig struct {
	// LimitPtr is a pointer to the maximum number of application layer
	// megabytes/period that the client may send over a cellular connection. If
	// nil, there is no limit. This field is provided only for serialization.
	// Clients should use Limit and LimitDefined instead.
	LimitPtr *int `json:"cellular_data_limit"`

	Limit        int
	LimitDefined bool

	// Date is the day of the month that delimits a cellular
	// data plan period. Valid values are within [1, 28].
	Date int `json:"normalized_cell_plan_date"`
}

// DataLimit represents parameters that control the client's use of data.
type DataLimit struct {
	// EmissionPeriod is the time in seconds the client is to wait
	// between emission requests to the agent.
	EmissionPeriod int `json:"emission_period"`
	Storage        struct {
		// Limit is the maximum number of megabytes the client
		// may use to store unsent messages. If nil, there is no
		// storage limit.
		Limit *int `json:"storage_limit"`
	} `json:"storage"`
	Cellular CellularConfig `json:"data"`
}

// GetDataLimit returns a DataLimit from the dataLimit endpoint.
func GetDataLimit() (*DataLimit, error) {
	resp, err := get(dataLimitEP, "application/json")
	if err != nil {
		return nil, err
	}

	var l struct {
		DataLimit `json:"config"`
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &l); err != nil {
		return nil, errEncoding{err, string(body), "GetDataLimit"}
	}

	depointerize(&l.DataLimit.Cellular)
	return &l.DataLimit, nil
}

func depointerize(c *CellularConfig) {
	if c.LimitPtr == nil {
		return
	}
	c.LimitDefined = true
	c.Limit = *c.LimitPtr
}
