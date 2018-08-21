package config

import (
	"os"
	"testing"
)

func TestLocalBuild(t *testing.T) {
	cases := []struct{
		getenv func(string) string
		expect Config
	} {
		{
			getenv: func(key string) string { return "" },
			expect: Config{
				BaseURL:   Production,
				LogErrors: false,
				LogInfo:   false,
			},
		},
		{
			getenv: func(key string) string {
				if key == "AUKLET_BASE_URL" {
					return "something"
				}
				return ""
			},
			expect: Config{
				BaseURL:   "something",
				LogErrors: false,
				LogInfo:   false,
			},
		},
		{
			getenv: func(key string) string {
				if key == "AUKLET_LOG_ERRORS" {
					return "true"
				}
				return ""
			},
			expect: Config{
				BaseURL:   Production,
				LogErrors: true,
				LogInfo:   false,
			},
		},
		{
			getenv: func(key string) string {
				if key == "AUKLET_LOG_INFO" {
					return "true"
				}
				return ""
			},
			expect: Config{
				BaseURL:   Production,
				LogErrors: false,
				LogInfo:   true,
			},
		},
	}

	for i, c := range cases {
		getenv = c.getenv
		if got := LocalBuild(); got != c.expect {
			t.Errorf("case %v: got %v, expected %v", i, got, c.expect)
		}
		getenv = os.Getenv
	}
}

func TestReleaseBuild(t *testing.T) {
	cases := []struct {
		getenv func(string) string
		expect Config
	} {
		{
			getenv: func(key string) string { return "" },
			expect: Config{
				BaseURL: StaticBaseURL,
				LogErrors: false,
				LogInfo: false,
			},
		},
	}
	
	for i, c := range cases {
		getenv = c.getenv
		if got := ReleaseBuild(); got != c.expect {
			t.Errorf("case %v: got %v, expected %v", i, got, c.expect)
		}
		getenv = os.Getenv
	}
}
