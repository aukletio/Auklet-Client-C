// Package certs provides access to SSL certificates from the API.
package certs

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// FromURL fetches zipped SSL certs against baseurl using apikey and returns
// them as a *tls.Config.
func FromURL(baseurl, apikey string) (c *tls.Config) {
	url := baseurl + "/certificates/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print(err)
		return
	}
	req.Header.Add("Authorization", "JWT"+apikey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("request to %v with API key %v... gave unexpected status %v",
			url, apikey[:10], resp.Status)
		return
	}
	cts, err := unpack(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}
	return cts.tlsconfig()
}

// certs represents SSL certificates.
type certs struct {
	ca         []byte
	cert       []byte
	privatekey []byte
}

// tlsconfig converts c into a *tls.Config.
func (c *certs) tlsconfig() (tc *tls.Config) {
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(c.ca)
	cert, err := tls.X509KeyPair(c.cert, c.privatekey)
	if err != nil {
		log.Print(err)
		return
	}
	tc = &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}
	return
}

// unpack reads certs from a zip-formatted stream and puts them into
// a map. The process is not simple:
//
// ioutil.ReadAll  : io.Reader   -> []byte
// bytes.NewReader : []byte      -> bytes.Reader (implements io.ReaderAt)
// zip.NewReader   : io.ReaderAt -> zip.Reader (array of zip.File)
// zip.Open        : zip.File    -> io.ReadCloser (implements io.Reader)
// ioutil.ReadAll  : io.Reader   -> []byte
func unpack(r io.Reader) (c *certs, err error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return
	}
	m := make(map[string][]byte)
	for _, f := range z.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		cert, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		m[f.Name] = cert
	}
	err = verify(m)
	if err != nil {
		return
	}
	c = &certs{
		ca:         m["ck_ca"],
		cert:       m["ck_cert"],
		privatekey: m["ck_private_key"],
	}
	return
}

// verify checks that the map produced by unpack has the right
// number of files and the correct file names.
func verify(m map[string][]byte) (err error) {
	filenames := []string{"ck_ca", "ck_cert", "ck_private_key"}
	errs := []string{}
	if len(m) != len(filenames) {
		format := "got %v cert files, expected %v"
		errs = append(errs, fmt.Sprintf(format, len(m), len(filenames)))
	}
	for _, name := range filenames {
		if _, ok := m[name]; !ok {
			errs = append(errs, "could not find cert file named "+name)
		}
	}
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, "\n"))
	}
	return
}
