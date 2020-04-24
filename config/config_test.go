package config

import (
	"reflect"
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
	if !strings.Contains(err.Error(), "required key") || !strings.Contains(err.Error(), "missing value") {
		t.Log(err)
		t.Fail()
	}
}

func TestBuildRabbitConfigurationString(t *testing.T) {
	cfg = &Configuration{
		RabbitHost:       "testHost",
		RabbitPort:       "123",
		RabbitUsername:   "testUser",
		RabbitPassword:   "testPass",
		RabbitVHost:      "/testVhost",
		EqReceiptProject: "testProject",
	}

	buildRabbitConnectionString(cfg)

	if !reflect.DeepEqual(cfg.RabbitConnectionString, "amqp://testUser:testPass@testHost:123/testVhost") {
		t.Logf("Built rabbit connection string did not match expected, found: %q", cfg.RabbitConnectionString)
		t.Fail()
	}
}
