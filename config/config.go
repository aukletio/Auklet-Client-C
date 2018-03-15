// Package config provides Auklet client configuration data.
package config

import (
	"log"
	"os"

	"github.com/ESG-USA/Auklet-Client/api"
)

// A Config represents the parameters of an Auklet client invocation. A
// Config can be defined programmatically or from the environment.
type Config struct {
	// A BaseURL is a URL of the API we would be working against; typically
	// either staging, QA, or production.
	BaseURL string

	// An App ID is a long string provided to the customer upon onboarding.
	// It identifies their application as a whole, but not a particular
	// release of it. It is used in API calls relating to devices and in
	// profile data sent to Kafka.
	AppID string

	// An API key is a long string provided to the customer upon onboarding
	// that grants them API access. It is used in most API calls, such as
	// requesting SSL certs and getting and posting a device.
	APIKey string

	// Dump enables console logs. In production it is false; in development
	// it is usually true.
	Dump bool
}

// Prefix is the prefix used by all Auklet environment variables.
const Prefix = "AUKLET_"

// LocalBuild creates a Config entirely from environment variables.
func LocalBuild() (c Config) {
	c = Config{
		BaseURL: os.Getenv(Prefix + "BASE_URL"),
		AppID:   os.Getenv(Prefix + "APP_ID"),
		APIKey:  os.Getenv(Prefix + "API_KEY"),
		Dump:    os.Getenv(Prefix+"DUMP") == "true",
	}
	if c.BaseURL == "" {
		c.BaseURL = api.Production
	}
	logEmptyFields(c)
	return
}

// ReleaseBuild creates a Config as would be required in a production
// environment. The base URL is hardcoded in this configuration and cannot be
// overridden by the end user.
func ReleaseBuild() (c Config) {
	c = Config{
		BaseURL: api.StaticBaseURL,
		AppID:   os.Getenv(Prefix + "APP_ID"),
		APIKey:  os.Getenv(Prefix + "API_KEY"),
		Dump:    false,
	}
	logEmptyFields(c)
	return
}

// logEmptyFields logs a warning for each empty field in c.
func logEmptyFields(c Config) (bad bool) {
	if c.BaseURL == "" {
		log.Print("warning: empty BASE_URL")
		bad = true
	}
	if c.AppID == "" {
		log.Print("warning: empty APP_ID")
		bad = true
	}
	if c.APIKey == "" {
		log.Print("warning: empty API_KEY")
		bad = true
	}
	return
}
