package config

import (
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	RabbitHost                 string `envconfig:"RABBIT_HOST" required:"true"`
	RabbitPort                 string `envconfig:"RABBIT_PORT" required:"true"`
	RabbitUsername             string `envconfig:"RABBIT_USERNAME" required:"true"`
	RabbitPassword             string `envconfig:"RABBIT_PASSWORD"  required:"true"  json:"-"`
	RabbitVHost                string `envconfig:"RABBIT_VHOST"  default:"/"`
	RabbitConnectionString     string `json:"-"`
	EventsExchange             string `envconfig:"RABBIT_EXCHANGE"  default:"events"`
	ReceiptRoutingKey          string `envconfig:"RECEIPT_ROUTING_KEY"  default:"event.response.receipt"`
	UndeliveredMailRoutingKey  string `envconfig:"UNDELIVERED_MAIL_ROUTING_KEY"  default:"event.fulfilment.undelivered"`
	EqReceiptProject           string `envconfig:"EQ_RECEIPT_PROJECT" required:"true"`
	EqReceiptSubscription      string `envconfig:"EQ_RECEIPT_SUBSCRIPTION" default:"rm-receipt-subscription"`
	EqReceiptTopic             string `envconfig:"EQ_RECEIPT_TOPIC" default:"eq-submission-topic"`
	OfflineReceiptProject      string `envconfig:"OFFLINE_RECEIPT_PROJECT" required:"true"`
	OfflineReceiptSubscription string `envconfig:"OFFLINE_RECEIPT_SUBSCRIPTION" default:"rm-offline-receipt-subscription"`
	OfflineReceiptTopic        string `envconfig:"OFFLINE_RECEIPT_TOPIC" default:"offline-receipt-topic"`
}

var cfg *Configuration

func GetConfig() (*Configuration, error) {
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
