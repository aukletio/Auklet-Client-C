package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	switch r.URL.Path {
	case ReleasesEP:
		checksum := r.URL.Query()["checksum"][0]
		if len(checksum) < 1 {
			http.Error(w, "404", http.StatusNotFound)
		}
	case CertificatesEP:
	case DevicesEP:
		w.WriteHeader(201)
		w.Write([]byte(`{"client_password":"nonempty"}`))
	case ConfigEP:
	case fmt.Sprintf(DataLimitEP, "appid"):
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	default:
		http.Error(w, "404", http.StatusNotFound)
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
		api  API
		path string
		ok   bool
	}{
		{
			api: API{BaseURL: s.URL, DevicesEP: DevicesEP + "bogus"},
			ok:  false,
		},
		{
			api:  API{BaseURL: s.URL, DevicesEP: DevicesEP},
			path: "testdata/noexist/file",
			ok:   false,
		},
		{
			api:  API{BaseURL: s.URL, DevicesEP: DevicesEP},
			path: "testdata/file",
			ok:   true,
		},
	}

	for i, c := range cases {
		_, err := getAndSaveCredentials(c.api, c.path)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func TestRelease(t *testing.T) {
	s := httptest.NewServer(handler)
	defer s.Close()

	api := API{
		BaseURL:    s.URL,
		ReleasesEP: ReleasesEP,
	}

	if err := api.Release(""); err == nil {
		t.Fail()
	}

	if err := api.Release("valid"); err != nil {
		t.Error(err)
	}
}

func TestCertificates(t *testing.T) {
	s := httptest.NewServer(handler)
	defer s.Close()

	api := API{
		BaseURL:        s.URL,
		CertificatesEP: CertificatesEP,
	}

	_, err := api.Certificates()
	if err != errParseCA {
		t.Errorf("expected %v, got %v", errParseCA, err)
	}
}

func TestBrokerAddress(t *testing.T) {
	s := httptest.NewServer(handler)
	defer s.Close()

	api := API{
		BaseURL:  s.URL,
		ConfigEP: ConfigEP,
	}

	_, err := api.BrokerAddress()
	if _, is := err.(errEncoding); !is {
		t.Errorf("expected errEncoding, got %v", err)
	}
}

func TestDataLimit(t *testing.T) {
	s := httptest.NewServer(handler)
	defer s.Close()

	api := API{
		BaseURL:     s.URL,
		DataLimitEP: DataLimitEP,
		AppID:       "appid",
	}

	_, err := api.DataLimit()
	if err != nil {
		t.Error(err)
	}
}
