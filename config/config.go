package config

import (
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	RabbitHost             string `envconfig:"RABBIT_HOST" required:"true"`
	RabbitPort             string `envconfig:"RABBIT_PORT" required:"true"`
	RabbitUsername         string `envconfig:"RABBIT_USERNAME" required:"true"`
	RabbitPassword         string `envconfig:"RABBIT_PASSWORD"  required:"true"  json:"-"`
	RabbitVHost            string `envconfig:"RABBIT_VHOST"  default:"/"`
	RabbitConnectionString string `json:"-"`
	EqReceiptProject       string `envconfig:"EQ_RECEIPT_PROJECT" required:"true"`
	EqReceiptSubscription  string `envconfig:"EQ_RECEIPT_SUBSCRIPTION" default:"rm-receipt-subscription"`
}

var cfg *Configuration

// Get the application and returns the configuration structure
func Get() (*Configuration, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Configuration{}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	buildRabbitConnectionString(cfg)

	return cfg, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Configuration) String() string {
	jsonConfig, _ := json.Marshal(config)
	return string(jsonConfig)
}

func buildRabbitConnectionString(cfg *Configuration) {
	if cfg.RabbitConnectionString == "" {
		cfg.RabbitConnectionString = fmt.Sprintf("amqp://%s:%s@%s:%s%s",
			cfg.RabbitUsername, cfg.RabbitPassword, cfg.RabbitHost, cfg.RabbitPort, cfg.RabbitVHost)
	}
}
