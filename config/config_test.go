package config

import (
	"testing"
)

func TestEmptyFields(t *testing.T) {
	c := Config{}
	if !logEmptyFields(c) {
		t.Fail()
	}
	c.BaseURL = "not empty"
	c.AppID = "not empty"
	c.APIKey = "not empty"
	c.Brokers = []string{"not empty"}
	c.LogTopic = "not empty"
	c.ProfileTopic = "not empty"
	c.EventTopic = "not empty"
	if logEmptyFields(c) {
		t.Fail()
	}
}
