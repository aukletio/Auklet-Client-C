package config

import (
	"os"
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/version"
)

func empty(key string) string { return "" }

func baseDefined(key string) string {
	if key == "AUKLET_BASE_URL" {
		return "something"
	}
	return ""
}

func errorsDefined(key string) string {
	if key == "AUKLET_LOG_ERRORS" {
		return "true"
	}
	return ""
}

func infoDefined(key string) string {
	if key == "AUKLET_LOG_INFO" {
		return "true"
	}
	return ""
}

func TestLocalBuild(t *testing.T) {
	cases := []struct {
		getenv func(string) string
		expect Config
	}{
		{
			getenv: empty,
			expect: Config{BaseURL: Production},
		},
		{
			getenv: baseDefined,
			expect: Config{BaseURL: "something"},
		},
		{
			getenv: errorsDefined,
			expect: Config{BaseURL: Production, LogErrors: true},
		},
		{
			getenv: infoDefined,
			expect: Config{BaseURL: Production, LogInfo: true},
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
	}{
		{
			getenv: empty,
			expect: Config{
				BaseURL: StaticBaseURL,
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

func TestAPIKey(t *testing.T) {
	getenv = func(string) string {
		return "api key"
	}

	if got := APIKey(); got != "api key" {
		t.Fail()
	}
}

func TestAppID(t *testing.T) {
	getenv = func(string) string {
		return "app ID"
	}

	if got := AppID(); got != "app ID" {
		t.Fail()
	}
}

func TestGet(t *testing.T) {
	if Get() != LocalBuild() {
		t.Fail()
	}
	version.Version = "not local build"
	if Get() != ReleaseBuild() {
		t.Fail()
	}
}
