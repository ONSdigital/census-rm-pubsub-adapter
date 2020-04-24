package processor

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/streadway/amqp"
	"log"
)

type Processor struct {
	RabbitConn         *amqp.Connection
	RabbitChan         *amqp.Channel
	Config             *config.Configuration
	PubSubClient       *pubsub.Client
	PubSubSubscription *pubsub.Subscription
	MessageChan        chan pubsub.Message
}

func NewProcessor(ctx context.Context, appConfig *config.Configuration, pubSubProject string, pubSubSubscription string) *Processor {
	//set up rabbit connection
	var err error
	a := &Processor{}
	a.Config = appConfig
	a.RabbitConn, err = amqp.Dial(appConfig.RabbitConnectionString)
	failOnError(err, "Failed to connect to RabbitMQ")

	a.RabbitChan, err = a.RabbitConn.Channel()
	failOnError(err, "Failed to open a channel")

	//setup pubsub connection
	a.PubSubClient, err = pubsub.NewClient(ctx, pubSubProject)
	failOnError(err, "Pubsub client creation failed")

	//setup subscription
	a.PubSubSubscription = a.PubSubClient.Subscription(pubSubSubscription)
	a.MessageChan = make(chan pubsub.Message)

	return a
}

func (a *Processor) Consume(ctx context.Context) {
	log.Printf("Launching PubSub message listener on subcription %s\n", a.PubSubSubscription.String())

	err := a.PubSubSubscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		a.MessageChan <- *msg
	})
	if err != nil {
		log.Printf("Receive: %v\n", err)
		return
	}
}

func (a *Processor) publishEventToRabbit(message *models.RmMessage, routingKey string, exchange string) error {

	byteMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = a.RabbitChan.Publish(
		exchange,
		routingKey, // routing key (the queue)
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         byteMessage,
			DeliveryMode: 2, // 2 = persistent delivery mode
		})
	if err != nil {
		return err
	}

	log.Printf(" [x] Sent %s", string(byteMessage))
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
