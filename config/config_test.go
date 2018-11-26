package config

import (
	"testing"
)

func TestBaseURL(t *testing.T) {
	baseURL := "http://example.com"
	fromCliBaseURL := "http://other.example.com"
	empty := func(string) string { return "" }
	nonempty := func(string) string { return baseURL }

	cases := []struct {
		baseURL string
		getenv  Getenv
		expect  string
	}{
		{
			baseURL: fromCliBaseURL,
			getenv:  nonempty,
			expect:  fromCliBaseURL,
		},
		{
			baseURL: "",
			getenv:  nonempty,
			expect:  baseURL,
		},
		{
			baseURL: "",
			getenv:  empty,
			expect:  Production,
		},
	}

	for i, c := range cases {
		got := c.getenv.BaseURL(c.baseURL)
		if got != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, got)
		}
	}
}
