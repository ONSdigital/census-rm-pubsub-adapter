package eqreceipt

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models/incoming-pubsub"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models/rabbit"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type App struct {
	RabbitConn   *amqp.Connection
	RabbitChan   *amqp.Channel
	PubSubClient *pubsub.Client
	EqRecieptSub *pubsub.Subscription
	ReceiptChan  chan incoming_pubsub.EqReceipt
}

func (a *App) Setup(ctx context.Context, rabbitConnectionString, projectID string) {

	//set up rabbit connection
	var err error
	a.RabbitConn, err = amqp.Dial(rabbitConnectionString)
	failOnError(err, "Failed to connect to RabbitMQ")

	a.RabbitChan, err = a.RabbitConn.Channel()
	failOnError(err, "Failed to open a channel")

	//setup pubsub connection
	a.PubSubClient, err = pubsub.NewClient(ctx, projectID)
	failOnError(err, "Pubsub client creation failed")
	//setup subcriptions
	a.EqRecieptSub = a.PubSubClient.Subscription("rm-receipt-subscription")
	a.ReceiptChan = make(chan incoming_pubsub.EqReceipt)

}

func (a *App) Produce(ctx context.Context) {
	for {
		select {
		case eqReceiptReceived := <-a.ReceiptChan:
			rmMessageToSend, err := convertEqReceiptToRmMessage(&eqReceiptReceived)
			if err != nil {
				log.Println(errors.Wrap(err, "failed to convert receipt to message"))
			}
			sendRabbitMessage(rmMessageToSend, a.RabbitChan)
		}
	}

}

func (a *App) Consume(ctx context.Context) {
	log.Println("Launched PubSub message listener")

	// Not using the cancel function here (which makes WithCancel a bit redundant!)
	// Ideally the cancel would be deferred/delegated to a signal watcher to
	// enable graceful shutdown that can kill the active go routines.
	cctx, _ := context.WithCancel(ctx)
	err := a.EqRecieptSub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))

		eqReceiptReceived := incoming_pubsub.EqReceipt{}
		json.Unmarshal(msg.Data, &eqReceiptReceived)

		log.Printf("Got QID: %q\n", eqReceiptReceived.Metadata.QuestionnaireID)

		a.ReceiptChan <- eqReceiptReceived
		msg.Ack()

	})
	if err != nil {
		log.Printf("Receive: %v\n", err)
	}
}

func sendRabbitMessage(message *rabbit.RmMessage, ch *amqp.Channel) {

	byteMessage, err := json.Marshal(message)
	failOnError(err, "Failed to marshall data")

	err = ch.Publish(
		"",            // default exchange
		"goTestQueue", // routing key (the queue)
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        byteMessage,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", string(byteMessage))
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func convertEqReceiptToRmMessage(eqReceipt *incoming_pubsub.EqReceipt) (*rabbit.RmMessage, error) {
	if eqReceipt == nil {
		return nil, errors.New("receipt has nil content")
	}

	return &rabbit.RmMessage{
		Event: rabbit.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "EQ",
			DateTime:      eqReceipt.TimeCreated,
			TransactionID: eqReceipt.Metadata.TransactionID,
		},
		Payload: rabbit.RmPayload{
			Response: rabbit.RmResponse{
				QuestionnaireID: eqReceipt.Metadata.QuestionnaireID,
			},
		},
	}, nil
}
