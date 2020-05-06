package processor

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/logger"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type messageUnmarshaller func([]byte) (models.PubSubMessage, error)

type messageConverter func(message models.PubSubMessage) (*models.RmMessage, error)

type Processor struct {
	RabbitConn         *amqp.Connection
	RabbitChannel      *amqp.Channel
	RabbitRoutingKey   string
	Config             *config.Configuration
	PubSubClient       *pubsub.Client
	PubSubSubscription *pubsub.Subscription
	unmarshallMessage  messageUnmarshaller
	convertMessage     messageConverter
	ErrChan            chan error
	Logger             *zap.SugaredLogger
}

func NewProcessor(ctx context.Context,
	appConfig *config.Configuration,
	pubSubProject string,
	pubSubSubscription string,
	routingKey string,
	messageConverter messageConverter,
	messageUnmarshaller messageUnmarshaller, errChan chan error) (*Processor, error) {
	var err error
	p := &Processor{}
	p.Config = appConfig
	p.RabbitRoutingKey = routingKey
	p.convertMessage = messageConverter
	p.unmarshallMessage = messageUnmarshaller
	p.ErrChan = errChan
	p.Logger = logger.Logger.With("subscription", pubSubSubscription)

	// Set up rabbit connection
	p.RabbitConn, err = amqp.Dial(appConfig.RabbitConnectionString)
	if err != nil {
		return nil, err
	}

	p.RabbitChannel, err = p.RabbitConn.Channel()
	if err != nil {
		return nil, err
	}

	// Setup PubSub connection
	p.PubSubClient, err = pubsub.NewClient(ctx, pubSubProject)
	if err != nil {
		return nil, err
	}

	// Setup subscription
	p.PubSubSubscription = p.PubSubClient.Subscription(pubSubSubscription)

	// Start consuming from PubSub
	p.Logger.Infow("Launching PubSub message receiver")
	go p.Consume(ctx)

	return p, nil
}

func (p *Processor) Consume(ctx context.Context) {
	err := p.PubSubSubscription.Receive(ctx, p.Process)
	if err != nil {
		p.Logger.Errorw("Error in consumer", "error", err)
		p.ErrChan <- err
	}
}

func (p *Processor) Process(_ context.Context, msg *pubsub.Message) {
	ctxLogger := p.Logger.With("msgId", msg.ID)
	messageReceived, err := p.unmarshallMessage(msg.Data)
	if err != nil {
		ctxLogger.Errorw("Error unmarshalling message, quarantining", "error", err, "data", string(msg.Data))
		err = p.quarantineMessageInRabbit(msg)
		if err != nil {
			ctxLogger.Errorw("Error quarantining bad message, nacking", "error", err, "data", string(msg.Data))
			msg.Nack()
		} else {
			ctxLogger.Debugw("Acking quarantined message", "msgData", string(msg.Data))
			msg.Ack()
		}
		return
	}
	ctxLogger = ctxLogger.With("transactionId", messageReceived.GetTransactionId())
	ctxLogger.Debugw("Processing message")
	rmMessageToSend, err := p.convertMessage(messageReceived)
	if err != nil {
		ctxLogger.Errorw("Error converting message", "error", err)
	}
	err = p.publishEventToRabbit(rmMessageToSend, p.RabbitRoutingKey, p.Config.EventsExchange)
	if err != nil {
		ctxLogger.Errorw("Failed to publish message", "error", err)
		msg.Nack()
	} else {
		ctxLogger.Debugw("Acking message", "msgData", string(msg.Data))
		msg.Ack()
	}
}

func (p *Processor) publishEventToRabbit(message *models.RmMessage, routingKey string, exchange string) error {

	byteMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = p.RabbitChannel.Publish(
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

	p.Logger.Debugw("Published message", "routingKey", routingKey, "transactionId", message.Event.TransactionID)
	return nil
}

func (p *Processor) quarantineMessageInRabbit(message *pubsub.Message) error {
	headers := amqp.Table{
		"pubSubId": message.ID,
	}
	for key, value := range message.Attributes {
		headers[key] = value
	}
	err := p.RabbitChannel.Publish(
		"",
		p.Config.DlqRoutingKey, // routing key (the queue)
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message.Data,
			Headers:      headers,
			DeliveryMode: 2, // 2 = persistent delivery mode
		})
	if err != nil {
		return err
	}

	p.Logger.Debugw("Quarantined message", "DlqRoutingKey", p.Config.DlqRoutingKey, "data", string(message.Data))
	return nil
}

func (p *Processor) CloseRabbit() {
	if err := p.RabbitChannel.Close(); err != nil {
		p.Logger.Errorw("Error closing rabbit channel during shutdown of processor", "error", err)
	}
	if err := p.RabbitConn.Close(); err != nil {
		p.Logger.Errorw("Error closing rabbit connection during shutdown of processor", "error", err)
	}
}
