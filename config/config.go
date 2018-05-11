// Package config provides Auklet client configuration data.
package config

import (
	"log"
	"os"
)

// A Config represents the parameters of an Auklet client invocation.
type Config struct {
	// A BaseURL is a URL of the API we would be working against;
	// typically either staging, QA, or production.
	BaseURL string

	// Dump enables console logs. In production it is false; in
	// development it is usually true.
	Dump bool
}

// Production defines the base URL for the production environment.
const Production = "https://api.auklet.io"

// StaticBaseURL is provided at compile-time; DO NOT MODIFY.
var StaticBaseURL = ""

// prefix is the prefix used by all Auklet environment variables.
const prefix = "AUKLET_"

// LocalBuild creates a Config entirely from environment variables.
func LocalBuild() (c Config) {
	c = Config{
		BaseURL: envar("BASE_URL"),
		Dump:    envar("DUMP") == "true",
	}
	if c.BaseURL == "" {
		c.BaseURL = Production
	}
	return
}

// ReleaseBuild creates a Config as would be required in a production
// environment. The base URL is hardcoded in this configuration and cannot be
// overridden by the end user.
func ReleaseBuild() Config {
	return Config{
		BaseURL: StaticBaseURL,
		Dump:    false,
	}
}

func envar(s string) string {
	k := os.Getenv(prefix + s)
	if k == "" {
		log.Print("warning: empty", s)
	}
	return k
}

// APIKey returns the API granted to the customer upon onboarding.
// It is used in most API calls, such as requesting SSL certs and getting and
// posting a device.
func APIKey() string {
	return envar("API_KEY")
}

// AppID identifies a customer's application as a whole, but not a particular
// release of it. It is used in API calls relating to devices and in profile
// data sent to Kafka.
func AppID() string {
	return envar("APP_ID")
}
