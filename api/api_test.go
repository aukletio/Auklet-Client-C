package api

import (
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	switch r.URL.Path {
	case releasesEP:
		if len(r.URL.Query()["checksum"]) < 1 {
			http.Error(w, "404", http.StatusNotFound)
		}
	case certificatesEP:
	case devicesEP:
		w.WriteHeader(201)
		w.Write([]byte(`{"client_password":"nonempty"}`))
	case configEP, dataLimitEP:
	default:
		http.Error(w, "404", http.StatusNotFound)
	}
}

type mockCall struct {
	url string
}

func (m mockCall) Request() *http.Request {
	req, err := http.NewRequest("GET", m.url, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func (mockCall) Handle(*http.Response) error { return nil }

func TestDo(t *testing.T) {
	s := httptest.NewServer(handler)
	BaseURL = s.URL
	defer s.Close()

	cases := []struct {
		call Call
		ok   bool
	}{
		{call: mockCall{""}, ok: false},
		{call: mockCall{s.URL}, ok: true},
	}

	for i, c := range cases {
		err := Do(c.call)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: got %v, expected %v: %v", i, ok, c.ok, err)
		}
	}
}

func TestRequest(t *testing.T) {
	cases := []interface{ Request() *http.Request }{
		Credentials{},
		Release{},
		Certificates{},
		BrokerAddress{},
		DataLimit{},
	}

	for i, c := range cases {
		if c.Request() == nil {
			t.Errorf("case %v: nil request", i)
		}
	}
}

func body(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

func TestHandle(t *testing.T) {
	cases := []struct {
		handler interface{ Handle(*http.Response) error }
		resp    http.Response
		ok      bool
	}{
		{
			handler: new(Credentials),
			resp:    http.Response{StatusCode: 404},
			ok:      false, // non-201 status
		},
		{
			handler: new(Credentials),
			resp:    http.Response{StatusCode: 201, Body: body(`}`)},
			ok:      false,
		},
		{
			handler: new(Credentials),
			resp: http.Response{
				StatusCode: 201,
				Body:       body(`{"client_password":""}`),
			},
			ok: false,
		},
		{
			handler: new(Credentials),
			resp: http.Response{
				StatusCode: 201,
				Body:       body(`{"client_password":"nonempty"}`),
			},
			ok: true,
		},
		{
			handler: Release{},
			resp:    http.Response{StatusCode: 404},
			ok:      false,
		},
		{
			handler: Release{},
			resp:    http.Response{StatusCode: 200},
			ok:      true,
		},
		{
			handler: new(Certificates),
			resp:    http.Response{StatusCode: 404},
			ok:      false, // non-200 status
		},
		{
			handler: new(Certificates),
			resp:    http.Response{StatusCode: 200, Body: body("")},
			ok:      false,
		},
		{
			handler: new(BrokerAddress),
			resp:    http.Response{StatusCode: 404},
			ok:      false,
		},
		{
			handler: new(BrokerAddress),
			resp:    http.Response{StatusCode: 200, Body: body("")},
			ok:      false,
		},
		{
			handler: new(BrokerAddress),
			resp: http.Response{
				StatusCode: 200,
				Body:       body(`{"brokers":null,"port":null}`),
			},
			ok: true,
		},
		{
			handler: new(DataLimit),
			resp:    http.Response{StatusCode: 404},
			ok:      false,
		},
		{
			handler: new(DataLimit),
			resp:    http.Response{StatusCode: 200, Body: body(``)},
			ok:      false,
		},
		{
			handler: new(DataLimit),
			resp:    http.Response{StatusCode: 200, Body: body(`{}`)},
			ok:      true,
		},
		{
			handler: new(DataLimit),
			resp: http.Response{
				StatusCode: 200,
				Body:       body(`{"config":{"data":{"cellular_data_limit":1}}}`),
			},
			ok: true,
		},
	}

	for i, c := range cases {
		err := c.handler.Handle(&c.resp)
		ok := err == nil
		if c.ok != ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func TestCredsFromFile(t *testing.T) {
	cases := []struct {
		path string
		ok   bool
	}{
		{path: "testdata/noexist", ok: false},
		{path: "testdata/invalid.json", ok: false},
		{path: "testdata/valid.json", ok: true},
	}

	for i, c := range cases {
		_, err := credsFromFile(c.path)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func TestGetAndSave(t *testing.T) {
	s := httptest.NewServer(handler)
	defer s.Close()

	cases := []struct {
		url, path string
		ok        bool
	}{
		{
			url: s.URL + "bogus",
			ok:  false,
		},
		{
			url:  s.URL,
			path: "testdata/noexist/file",
			ok:   false,
		},
		{
			url:  s.URL,
			path: "testdata/file",
			ok:   true,
		},
	}

	for i, c := range cases {
		BaseURL = c.url
		_, err := getAndSaveCredentials(c.path)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func TestTLSConfig(t *testing.T) {
	appendCertsFromPEM = func(*x509.CertPool, []byte) bool {
		return true
	}
	if _, err := tlsConfig([]byte{}); err == nil {
		t.Fail()
	}
}
