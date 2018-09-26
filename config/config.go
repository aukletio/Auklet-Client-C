// Package config provides Auklet client configuration data.
package config

import (
	"os"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// Production defines the base URL for the production environment.
const Production = "https://api.auklet.io"

type Getenv func(string) string

var OS Getenv = os.Getenv

// StaticBaseURL is provided at compile-time; DO NOT MODIFY.
var StaticBaseURL = ""

// prefix is the prefix used by all Auklet environment variables.
const prefix = "AUKLET_"

func (getenv Getenv) envar(s string) string {
	k := getenv(prefix + s)
	if k == "" {
		errorlog.Print("warning: empty ", prefix+s)
	}
	return k
}

// APIKey returns the API granted to the customer upon onboarding.
// It is used in most API calls, such as requesting SSL certs and getting and
// posting a device.
func (getenv Getenv) APIKey() string {
	return getenv.envar("API_KEY")
}

// AppID identifies a customer's application as a whole, but not a particular
// release of it. It is used in API calls relating to devices and in profile
// data sent to broker.
func (getenv Getenv) AppID() string {
	return getenv.envar("APP_ID")
}

// BaseURL returns the base URL, as dependent on the version.
func (getenv Getenv) BaseURL(version string) string {
	if version == "local-build" {
		url := getenv.envar("BASE_URL")
		if url == "" {
			return Production
		}
		return url
	}
	return StaticBaseURL
}

func (getenv Getenv) LogErrors() bool {
	return getenv(prefix+"LOG_ERRORS") == "true"
}

func (getenv Getenv) LogInfo() bool {
	return getenv(prefix+"LOG_INFO") == "true"
}
