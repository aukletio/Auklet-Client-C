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
	if logEmptyFields(c) {
		t.Fail()
	}
}
