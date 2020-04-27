package processor

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"log"
)

type messageUnmarshaller func([]byte) (models.PubSubMessage, error)

type messageConverter func(message models.PubSubMessage) (*models.RmMessage, error)

type Processor struct {
	RabbitConn         *amqp.Connection
	RabbitChan         *amqp.Channel
	Config             *config.Configuration
	PubSubClient       *pubsub.Client
	PubSubSubscription *pubsub.Subscription
	MessageChan        chan pubsub.Message
	unmarshallMessage  messageUnmarshaller
	convertMessage     messageConverter
}

func NewProcessor(ctx context.Context,
	appConfig *config.Configuration,
	pubSubProject string,
	pubSubSubscription string,
	messageConverter messageConverter,
	messageUnmarshaller messageUnmarshaller) *Processor {
	var err error
	a := &Processor{}
	a.Config = appConfig
	a.convertMessage = messageConverter
	a.unmarshallMessage = messageUnmarshaller

	//set up rabbit connection
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

func (p *Processor) Consume(ctx context.Context) {
	log.Printf("Launching PubSub message listener on subcription %s\n", p.PubSubSubscription.String())

	err := p.PubSubSubscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		p.MessageChan <- *msg
	})
	if err != nil {
		log.Printf("Receive: %v\n", err)
		return
	}
}

func (p *Processor) Process(ctx context.Context) {
	for {
		select {
		case msg := <-p.MessageChan:
			messageReceived, err := p.unmarshallMessage(msg.Data)
			if err != nil {
				// TODO Log the error and DLQ the message when unmarshalling fails, printing it out is p temporary solution
				log.Println(errors.WithMessagef(err, "Error unmarshalling message: %q", string(msg.Data)))
				msg.Ack()
				return
			}
			log.Printf("Got tx_id: %q\n", messageReceived.GetTransactionId())
			rmMessageToSend, err := p.convertMessage(messageReceived)
			if err != nil {
				log.Println(errors.Wrap(err, "failed to convert receipt to message"))
			}
			err = p.publishEventToRabbit(rmMessageToSend, p.Config.ReceiptRoutingKey, p.Config.EventsExchange)
			if err != nil {
				log.Println(errors.WithMessagef(err, "Failed to publish eq receipt message tx_id: %s", rmMessageToSend.Event.TransactionID))
				msg.Nack()
			} else {
				msg.Ack()
			}
		case <-ctx.Done():
			//stop the loop from consuming messages
			return
		}
	}
}

func (p *Processor) publishEventToRabbit(message *models.RmMessage, routingKey string, exchange string) error {

	byteMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = p.RabbitChan.Publish(
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

func (p *Processor) Shutdown() {
	// Shutdown Rabbit
	err := p.RabbitChan.Close()
	if err != nil {
		log.Println(errors.Wrapf(err, "Error closing rabbit channel during shutdown of %s processor", p.PubSubSubscription))
	}
	err = p.RabbitConn.Close()
	if err != nil {
		log.Println(errors.Wrapf(err, "Error closing rabbit connection during shutdown of %s processor", p.PubSubSubscription))
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
