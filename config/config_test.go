package config

import (
	"testing"
)

func TestBaseURL(t *testing.T) {
	baseURL := "http://example.com"
	empty := func(string) string { return "" }
	nonempty := func(string) string { return baseURL }

	cases := []struct {
		version string
		getenv  Getenv
		expect  string
	}{
		{
			version: "local-build",
			getenv:  empty,
			expect:  Production,
		},
		{
			version: "local-build",
			getenv:  nonempty,
			expect:  baseURL,
		},
		{
			version: "not local-build",
			getenv:  nil,
			expect:  StaticBaseURL,
		},
	}

	for i, c := range cases {
		got := c.getenv.BaseURL(c.version)
		if got != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, got)
		}
	}
}
