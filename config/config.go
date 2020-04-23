package config

import (
	"encoding/json"
	"github.com/kelseyhightower/envconfig"
)

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	RabbitHost            string `envconfig:"RABBIT_HOST"`
	RabbitPort            string `envconfig:"RABBIT_PORT"`
	RabbitUsername        string `envconfig:"RABBIT_USERNAME"`
	RabbitPassword        string `envconfig:"RABBIT_PASSOWRD"  json:"-"`
	RabbitVHost           string `envconfig:"RABBIT_VHOST"`
	EqReceiptProject      string `envconfig:"EQ_RECEIPT_PROJECT"`
	EqReceiptSubscription string `envconfig:"EQ_RECEIPT_SUBSCRIPTION"`
}

var cfg *Configuration

// Get the application and returns the configuration structure
func Get() (*Configuration, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Configuration{
		RabbitHost:            "localhost",
		RabbitPort:            "6672",
		RabbitUsername:        "guest",
		RabbitPassword:        "guest",
		RabbitVHost:           "/",
		EqReceiptProject:      "project",
		EqReceiptSubscription: "rm-receipt-subscription",
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Configuration) String() string {
	jsonConfig, _ := json.Marshal(config)
	return string(jsonConfig)
}
