// Package certs converts zipped SSL certificates into a usable format.
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
	"strings"

	"github.com/ESG-USA/Auklet-Client/errorlog"
)

// certs represents SSL certificates.
type certs struct {
	ca         []byte
	cert       []byte
	privatekey []byte
}

// TLSConfig converts c into a *tls.Config.
func (c *certs) TLSConfig() (tc *tls.Config) {
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(c.ca) {
		errorlog.Print("warning: failed to parse CA")
		return
	}
	tc = &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: false,
	}
	return
}

// Unpack reads certs from a zip-formatted stream and puts them into
// a map. The process is not simple:
//
// ioutil.ReadAll  : io.Reader   -> []byte
// bytes.NewReader : []byte      -> bytes.Reader (implements io.ReaderAt)
// zip.NewReader   : io.ReaderAt -> zip.Reader (array of zip.File)
// zip.Open        : zip.File    -> io.ReadCloser (implements io.Reader)
// ioutil.ReadAll  : io.Reader   -> []byte
func Unpack(r io.Reader) (c *certs, err error) {
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

// verify checks that the map produced by Unpack has the right
// number of files and the correct file names.
func verify(m map[string][]byte) (err error) {
	filenames := []string{"ck_ca"}
	errs := []string{}
	if len(m) < len(filenames) {
		format := "got %v cert files, expected at least %v"
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
