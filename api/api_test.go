package api

import (
	"net/http/httptest"
	"testing"
	"net/http"
)

func TestGet(t *testing.T) {
	var handler http.HandlerFunc = func(http.ResponseWriter, *http.Request) {}
	s := httptest.NewServer(handler)
	defer s.Close()

	cases := []struct {
		args string
		ok   bool
	}{
		// unsupported protocol schemes
		{args: ":", ok: false},
		{args: "", ok: false},

		{args: s.URL, ok: true},
	}

	for i, c := range cases {
		_, err := get(c.args, "nonempty")
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: got %v, expected %v: %v", i, ok, c.ok, err)
		}
	}
}
