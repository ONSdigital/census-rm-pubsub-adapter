package config

import (
	"strings"
	"testing"
)

func TestGetErrorsWithoutValues(t *testing.T) {
	cfg = nil
	_, err := Get()
	if err == nil {
		t.Log("Config get did not return an error in prod mode when not provided any values")
		t.Fail()
		return
	}
	if !strings.Contains(err.Error(), "required key") || !strings.Contains(err.Error(), "missing value"){
		t.Log(err)
		t.Fail()
	}
}
