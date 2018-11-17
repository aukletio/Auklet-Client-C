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
		baseUrl string
		getenv  Getenv
		expect  string
	}{
		{
			baseUrl: fromCliBaseURL,
			getenv:  nonempty,
			expect:  fromCliBaseURL,
		},
		{
			baseUrl: "",
			getenv:  nonempty,
			expect:  baseURL,
		},
		{
			baseUrl: "",
			getenv:  empty,
			expect:  Production,
		},
	}

	for i, c := range cases {
		got := c.getenv.BaseURL(c.baseUrl)
		if got != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, got)
		}
	}
}
