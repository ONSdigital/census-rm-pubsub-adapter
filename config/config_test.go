package config

import (
	"reflect"
	"testing"
)

func TestGetReturnsDefaultValues(t *testing.T) {
	testConfig, err := Get()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(testConfig, &Configuration{
		RabbitHost:            "localhost",
		RabbitPort:            "6672",
		RabbitUsername:        "guest",
		RabbitPassword:        "guest",
		RabbitVHost:           "/",
		EqReceiptProject:      "project",
		EqReceiptSubscription: "rm-receipt-subscription",
	}) {
		t.Logf("Default config did not match the expected values, found config: %q", testConfig)
		t.Fail()
	}
}
