// Package config provides Auklet client configuration data.
package config

import (
	"log"
	"os"
	"strings"
)

// Type Config represents the parameters of an Auklet client invocation. A
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

	// Brokers is a list of broker addresses used by Kafka.
	Brokers []string

	// LogTopic, ProfileTopic, and EventTopic are topics to which we produce
	// Kafka messages.
	LogTopic, ProfileTopic, EventTopic string

	// Dump enables console logs. In production it is false; in development
	// it is usually true.
	Dump bool
}

// Prefix is the prefix used by all Auklet environment variables.
const Prefix = "AUKLET_"

// FromEnv creates a Config entirely from environment variables.
func FromEnv() (c Config) {
	c = Config{
		BaseURL:      os.Getenv(Prefix + "BASE_URL"),
		AppID:        os.Getenv(Prefix + "APP_ID"),
		APIKey:       os.Getenv(Prefix + "API_KEY"),
		Brokers:      strings.Split(os.Getenv(Prefix+"BROKERS"), ","),
		LogTopic:     os.Getenv(Prefix + "LOG_TOPIC"),
		ProfileTopic: os.Getenv(Prefix + "PROF_TOPIC"),
		EventTopic:   os.Getenv(Prefix + "EVENT_TOPIC"),
		Dump:         os.Getenv(Prefix+"DUMP") == "true",
	}
	logEmptyFields(c)
	return
}

// Production creates a Config as would be required in a production environment.
func Production() (c Config) {
	c = Config{
		BaseURL:      "https://api.auklet.io/private",
		AppID:        os.Getenv(Prefix + "APP_ID"),
		APIKey:       os.Getenv(Prefix + "API_KEY"),
		Brokers:      strings.Split(os.Getenv(Prefix+"BROKERS"), ","),
		LogTopic:     os.Getenv(Prefix + "LOG_TOPIC"),
		ProfileTopic: os.Getenv(Prefix + "PROF_TOPIC"),
		EventTopic:   os.Getenv(Prefix + "EVENT_TOPIC"),
		Dump:         false,
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
	if c.LogTopic == "" {
		log.Print("warning: empty LOG_TOPIC")
		bad = true
	}
	if c.ProfileTopic == "" {
		log.Print("warning: empty PROF_TOPIC")
		bad = true
	}
	if c.EventTopic == "" {
		log.Print("warning: empty EVENT_TOPIC")
		bad = true
	}
	if strings.Join(c.Brokers, "") == "" {
		log.Print("warning: empty BROKERS")
		bad = true
	}
	return
}
