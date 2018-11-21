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
	"os"

	"github.com/spf13/afero"

	"github.com/aukletio/Auklet-Client-C/fsutil"
)

// namespaces and endpoints for the API. All new endpoints should be entered
// here.
const (
	ReleasesEP     = "/private/releases/"
	CertificatesEP = "/private/devices/certificates/"
	DevicesEP      = "/private/devices/"
	ConfigEP       = "/private/devices/config/"
	DataLimitEP    = "/private/devices/%s/app_config/" // app id
)

// Fs provides file system functions.
type Fs interface {
	Open(string) (afero.File, error)
	OpenFile(string, int, os.FileMode) (afero.File, error)
}

// API provides an interface to the backend.
type API struct {
	// BaseURL is the subdomain against which requests will be performed. It
	// should not assume any particular namespace.
	BaseURL string
	Key     string
	AppID   string
	MacHash string

	// for credentials
	CredsPath string // where to save/load credentials
	Fs        Fs     // filesystem for saving/loading

	ReleasesEP     string
	CertificatesEP string
	DevicesEP      string
	ConfigEP       string
	DataLimitEP    string
}

// Release is an API call that checks whether
// the given checksum has been released.
func (a API) Release(checksum string) error {
	url := a.BaseURL + a.ReleasesEP + "?checksum=" + checksum
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "JWT "+a.Key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("checking release %v: %v", checksum, err)
	}
	if resp.StatusCode != 200 {
		return errNotReleased{checksum}
	}
	return nil
}

type errNotReleased struct {
	checksum string
}

func (err errNotReleased) Error() string {
	return fmt.Sprintf("not released: %v", err.checksum)
}

// Certificates retrieves CA certs.
func (a API) Certificates() (*tls.Config, error) {
	url := a.BaseURL + a.CertificatesEP
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "JWT "+a.Key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting certificates: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, errStatus{resp}
	}
	ca, _ := ioutil.ReadAll(resp.Body)
	return tlsConfig(ca)
}

// tlsConfig converts ca into a *tls.Config.
func tlsConfig(ca []byte) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(ca) {
		return nil, errParseCA
	}
	// We trust Go's PEM parser; no need to cover the successful case in tests.
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: false,
	}, nil
}

var errParseCA = errors.New("failed to parse CA")

// Credentials represents credentials required for sending broker messages.
type Credentials struct {
	Username string `json:"id"`
	Password string `json:"client_password"`
	Org      string `json:"organization"`
	ClientID string `json:"client_id"`
}

// credentials retrieves Credentials for sending broker messages.
func (a API) credentials() (*Credentials, error) {
	b, _ := json.Marshal(struct {
		Mac   string `json:"mac_address_hash"`
		AppID string `json:"application"`
	}{
		// device info
		Mac:   a.MacHash,
		AppID: a.AppID,
	})
	url := a.BaseURL + a.DevicesEP
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "JWT "+a.Key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting broker credentials: %v", err)
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
	c := new(Credentials)
	if err := json.Unmarshal(data, c); err != nil {
		return nil, errEncoding{err, string(data), "decodeCredentials"}
	}

	if c.Password == "" {
		return nil, errors.New("empty password")
	}

	return c, nil
}

// Credentialer provides a way to get Credentials.
type Credentialer interface {
	Credentials() (*Credentials, error)
}

// Credentials retrieves credentials from the filesystem,
// with a fallback to the API. If credentials are retrieved
// from the API, they are saved to the filesystem.
func (a API) Credentials() (*Credentials, error) {
	// Not covered in tests, as its callees are covered.

	c, err := credsFromFile(a.CredsPath, a.Fs.Open)
	if err != nil {
		// file doesn't exist; ask the API for credentials
		return a.getAndSaveCredentials()
	}
	return c, nil
}

type openFunc func(string) (afero.File, error)

func readFile(path string, open openFunc) ([]byte, error) {
	f, err := open(path)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func credsFromFile(path string, open openFunc) (*Credentials, error) {
	b, err := readFile(path, open)
	if err != nil {
		return nil, fmt.Errorf("could not read credentials file: %v", err)
	}
	// decrypt here
	c := new(Credentials)
	if err := json.Unmarshal(b, c); err != nil {
		return nil, errEncoding{err, string(b), "credsFromFile"}
	}
	return c, nil
}

// getAndSaveCredentials requests credentials from the API. If it receives them,
// it writes them to the given path.
func (a API) getAndSaveCredentials() (*Credentials, error) {
	c, err := a.credentials()
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(c)
	// encrypt here
	if err := fsutil.WriteFile(a.Fs.OpenFile, a.CredsPath, b); err != nil {
		return nil, fmt.Errorf("could not write credentials: %v", err)
	}
	return c, nil
}

// BrokerAddress returns an address to which we can send broker messages.
func (a API) BrokerAddress() (string, error) {
	req, _ := http.NewRequest("GET", a.BaseURL+a.ConfigEP, nil)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "JWT "+a.Key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("getting broker address: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", errStatus{resp}
	}
	var k struct {
		Broker string `json:"brokers"`
		Port   string `json:"port"`
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &k); err != nil {
		return "", errEncoding{err, string(body), "BrokerAddress"}
	}
	return fmt.Sprintf("ssl://%s:%s", k.Broker, k.Port), nil
}

// DataLimit holds parameters controlling Auklet's data usage.
type DataLimit struct {
	Storage        *int64
	EmissionPeriod int
	Cellular       CellularConfig
}

// CellularConfig holds parameters for a cellular plan.
type CellularConfig struct {
	Date int // day of the month in [1,28]

	Defined bool
	Limit   int
}

// DataLimit retrieves DataLimit parameters from the backend.
func (a API) DataLimit() (*DataLimit, error) {
	url := a.BaseURL + fmt.Sprintf(a.DataLimitEP, a.AppID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "JWT "+a.Key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting data limit configuration: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, errStatus{resp}
	}
	type storage struct {
		Limit *int64 `json:"storage_limit"`
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
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &l); err != nil {
		return nil, errEncoding{err, string(body), "DataLimit"}
	}
	c := l.Config
	return &DataLimit{
		Storage:        c.Storage.Limit,
		EmissionPeriod: c.EmissionPeriod,
		Cellular: CellularConfig{
			Date:    c.Data.Date,
			Defined: c.Data.Limit != nil,
			Limit: func() int {
				if c.Data.Limit != nil {
					return *c.Data.Limit
				}
				return 0
			}(),
		},
	}, nil
}

type errStatus struct {
	resp *http.Response
}

func (err errStatus) Error() string {
	return fmt.Sprintf("unexpected status: %v from %v", err.resp.Status, err.resp.Request.URL)
}

type errEncoding struct {
	Err  error
	What string
	Op   string
}

func (err errEncoding) Error() string {
	return fmt.Sprintf("encoding error during %v: %v in %v", err.Op, err.Err, err.What)
}
