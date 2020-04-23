package processor

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type App struct {
	RabbitConn   *amqp.Connection
	RabbitChan   *amqp.Channel
	PubSubClient *pubsub.Client
	EqReceiptSub *pubsub.Subscription
	MessageChan  chan pubsub.Message
}

func New(ctx context.Context, rabbitConnectionString, projectId string, subscriptionId string) *App {

	//set up rabbit connection
	var err error
	a := &App{}
	a.RabbitConn, err = amqp.Dial(rabbitConnectionString)
	failOnError(err, "Failed to connect to RabbitMQ")

	a.RabbitChan, err = a.RabbitConn.Channel()
	failOnError(err, "Failed to open a channel")

	//setup pubsub connection
	a.PubSubClient, err = pubsub.NewClient(ctx, projectId)
	failOnError(err, "Pubsub client creation failed")

	//setup subscription
	a.EqReceiptSub = a.PubSubClient.Subscription(subscriptionId)
	a.MessageChan = make(chan pubsub.Message)

	return a
}

func (a *App) Process(ctx context.Context) {
	for {
		select {
		case msg := <-a.MessageChan:
			var eqReceiptReceived models.EqReceipt
			err := json.Unmarshal(msg.Data, &eqReceiptReceived)
			if err != nil {
				// TODO Log the error and DLQ the message when unmarshalling fails, printing it out is a temporary solution
				log.Println(errors.WithMessagef(err, "Error unmarshalling message: %q", string(msg.Data)))
				msg.Ack()
				return
			}

			log.Printf("Got QID: %q\n", eqReceiptReceived.Metadata.QuestionnaireID)
			rmMessageToSend, err := convertEqReceiptToRmMessage(&eqReceiptReceived)
			if err != nil {
				log.Println(errors.Wrap(err, "failed to convert receipt to message"))
			}
			sendRabbitMessage(rmMessageToSend, a.RabbitChan)
			msg.Ack()
		case <-ctx.Done():
			//stop the loop from consuming messages
			return
		}
	}

}

func (a *App) Consume(ctx context.Context) {
	log.Println("Launched PubSub message listener")

	err := a.EqReceiptSub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		a.MessageChan <- *msg
	})
	if err != nil {
		log.Printf("Receive: %v\n", err)
	}
}

func sendRabbitMessage(message *models.RmMessage, ch *amqp.Channel) {

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

func convertEqReceiptToRmMessage(eqReceipt *models.EqReceipt) (*models.RmMessage, error) {
	if eqReceipt == nil {
		return nil, errors.New("receipt has nil content")
	}

	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "EQ",
			DateTime:      eqReceipt.TimeCreated,
			TransactionID: eqReceipt.Metadata.TransactionID,
		},
		Payload: models.RmPayload{
			Response: models.RmResponse{
				QuestionnaireID: eqReceipt.Metadata.QuestionnaireID,
			},
		},
	}, nil
}
