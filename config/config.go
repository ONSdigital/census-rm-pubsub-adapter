package config

import (
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	ReadinessFilePath      string `envconfig:"READINESS_FILE_PATH" default:"/tmp/pubsub-adapter-ready"`
	LogLevel               string `envconfig:"LOG_LEVEL" default:"ERROR"`
	QuarantineMessageUrl   string `envconfig:"QUARANTINE_MESSAGE_URL"  required:"true"`
	PublishersPerProcessor int    `envconfig:"PUBLISHERS_PER_PROCESSOR" default:"20"`
	RestartTimeout         int    `envconfig:"RESTART_TIMEOUT" default:"120"`

	// Rabbit
	RabbitHost                       string `envconfig:"RABBIT_HOST" required:"true"`
	RabbitPort                       string `envconfig:"RABBIT_PORT" required:"true"`
	RabbitUsername                   string `envconfig:"RABBIT_USERNAME" required:"true"`
	RabbitPassword                   string `envconfig:"RABBIT_PASSWORD"  required:"true"  json:"-"`
	RabbitVHost                      string `envconfig:"RABBIT_VHOST"  default:"/"`
	RabbitConnectionString           string `json:"-"`
	EventsExchange                   string `envconfig:"RABBIT_EXCHANGE"  default:"events"`
	ReceiptRoutingKey                string `envconfig:"RECEIPT_ROUTING_KEY"  default:"event.response.receipt"`
	UndeliveredRoutingKey            string `envconfig:"UNDELIVERED_ROUTING_KEY"  default:"event.fulfilment.undelivered"`
	FulfilmentConfirmationRoutingKey string `envconfig:"FULFILMENT_CONFIRMATION_ROUTING_KEY"  default:"event.fulfilment.confirmation"`
	FulfilmentRequestRoutingKey      string `envconfig:"FULFILMENT_REQUEST_ROUTING_KEY"  default:"event.fulfilment.request"`

	// PubSub
	EqReceiptProject                string `envconfig:"EQ_RECEIPT_PROJECT" required:"true"`
	EqReceiptSubscription           string `envconfig:"EQ_RECEIPT_SUBSCRIPTION" default:"rm-receipt-subscription"`
	EqReceiptTopic                  string `envconfig:"EQ_RECEIPT_TOPIC" default:"eq-submission-topic"`
	OfflineReceiptProject           string `envconfig:"OFFLINE_RECEIPT_PROJECT" required:"true"`
	OfflineReceiptSubscription      string `envconfig:"OFFLINE_RECEIPT_SUBSCRIPTION" default:"rm-offline-receipt-subscription"`
	OfflineReceiptTopic             string `envconfig:"OFFLINE_RECEIPT_TOPIC" default:"offline-receipt-topic"`
	PpoUndeliveredProject           string `envconfig:"PPO_UNDELIVERED_SUBSCRIPTION_PROJECT" required:"true"`
	PpoUndeliveredSubscription      string `envconfig:"PPO_UNDELIVERED_SUBSCRIPTION" default:"rm-ppo-undelivered-subscription"`
	PpoUndeliveredTopic             string `envconfig:"PPO_UNDELIVERED_TOPIC" default:"ppo-undelivered-topic"`
	QmUndeliveredProject            string `envconfig:"QM_UNDELIVERED_SUBSCRIPTION_PROJECT" required:"true"`
	QmUndeliveredSubscription       string `envconfig:"QM_UNDELIVERED_SUBSCRIPTION" default:"rm-qm-undelivered-subscription"`
	QmUndeliveredTopic              string `envconfig:"QM_UNDELIVERED_TOPIC" default:"qm-undelivered-topic"`
	FulfilmentConfirmedProject      string `envconfig:"FULFILMENT_CONFIRMED_PROJECT" required:"true"`
	FulfilmentConfirmedSubscription string `envconfig:"FULFILMENT_CONFIRMED_SUBSCRIPTION" default:"fulfilment-confirmed-subscription"`
	FulfilmentConfirmedTopic        string `envconfig:"FULFILMENT_CONFIRMED_TOPIC" default:"fulfilment-confirmed-topic"`
	EqFulfilmentProject             string `envconfig:"EQ_FULFILMENT_PROJECT" required:"true"`
	EqFulfilmentSubscription        string `envconfig:"EQ_FULFILMENT_SUBSCRIPTION" default:"eq-fulfilment-subscription"`
	EqFulfilmentTopic               string `envconfig:"EQ_FULFILMENT_TOPIC" default:"eq-fulfilment-topic"`
}

var cfg *Configuration
var TestConfig = &Configuration{
	PublishersPerProcessor:           1,
	RestartTimeout:                  5,
	ReadinessFilePath:               "/tmp/pubsub-adapter-ready",
	RabbitConnectionString:           "amqp://guest:guest@localhost:7672/",
	ReceiptRoutingKey:                "goTestReceiptQueue",
	UndeliveredRoutingKey:            "goTestUndeliveredQueue",
	FulfilmentRequestRoutingKey:      "goTestFulfilmentRequestQueue",
	FulfilmentConfirmationRoutingKey: "goTestFulfilmentConfirmedQueue",
	EqReceiptProject:                 "project",
	EqReceiptSubscription:            "rm-receipt-subscription",
	EqReceiptTopic:                   "eq-submission-topic",
	OfflineReceiptProject:            "offline-project",
	OfflineReceiptSubscription:       "rm-offline-receipt-subscription",
	OfflineReceiptTopic:              "offline-receipt-topic",
	PpoUndeliveredProject:            "ppo-undelivered-project",
	PpoUndeliveredTopic:              "ppo-undelivered-mail-topic",
	PpoUndeliveredSubscription:       "rm-ppo-undelivered-subscription",
	QmUndeliveredProject:             "qm-undelivered-project",
	QmUndeliveredTopic:               "qm-undelivered-mail-topic",
	QmUndeliveredSubscription:        "rm-qm-undelivered-subscription",
	FulfilmentConfirmedProject:       "fulfilment-confirmed-project",
	FulfilmentConfirmedSubscription:  "fulfilment-confirmed-subscription",
	FulfilmentConfirmedTopic:         "fulfilment-confirmed-topic",
	EqFulfilmentProject:              "eq-fulfilment-project",
	EqFulfilmentSubscription:         "eq-fulfilment-subscription",
	EqFulfilmentTopic:                "eq-fulfilment-topic",
}

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
