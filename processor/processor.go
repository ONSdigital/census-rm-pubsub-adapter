package processor

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/logger"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/streadway/amqp"
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
	MessageChan        chan *pubsub.Message
	unmarshallMessage  messageUnmarshaller
	convertMessage     messageConverter
	ErrChan            chan error
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
	p.MessageChan = make(chan *pubsub.Message)

	// Start processing messages on the channel
	ctxLogger := logger.Logger.With("subscription", p.PubSubSubscription.ID())
	ctxLogger.Infow("Launching message processor")
	go p.Process(ctx)

	// Start consuming from PubSub
	ctxLogger.Infow("Launching PubSub message receiver")
	go p.Consume(ctx)

	return p, nil
}

func (p *Processor) Consume(ctx context.Context) {
	err := p.PubSubSubscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {

		logger.Logger.Debugw("Consumer got msg", "data", string(msg.Data), "subscription", p.PubSubSubscription)
		p.MessageChan <- msg
	})
	if err != nil {
		p.ErrChan <- err
	}
}

func (p *Processor) Process(ctx context.Context) {
	for {
		select {
		case msg := <-p.MessageChan:
			messageReceived, err := p.unmarshallMessage(msg.Data)
			if err != nil {
				logger.Logger.Errorw("Error unmarshalling message, quarantining", "error", err, "data", string(msg.Data))
				err = p.quarantineMessageInRabbit(msg)
				if err != nil {
					logger.Logger.Errorw("Error quarantining bad message, nacking", "error", err, "data", string(msg.Data))
					msg.Nack()
				} else {
					logger.Logger.Infow("ACKING", "messageId", msg.ID, "msgData", msg.Data)
					msg.Ack()
				}
				continue
			}
			ctxLogger := logger.Logger.With("transactionId", messageReceived.GetTransactionId())
			ctxLogger.Debugw("Processing message")
			rmMessageToSend, err := p.convertMessage(messageReceived)
			if err != nil {
				ctxLogger.Errorw("Failed to convert message", "error", err)
			}
			err = p.publishEventToRabbit(rmMessageToSend, p.RabbitRoutingKey, p.Config.EventsExchange)
			if err != nil {
				ctxLogger.Errorw("Failed to publish message", "error", err)
				msg.Nack()
			} else {
				logger.Logger.Infow("ACKING", "messageId", msg.ID, "msgData", msg.Data)
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

	logger.Logger.Debugw("Published message", "routingKey", routingKey, "transactionId", message.Event.TransactionID)
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

	logger.Logger.Debugw("Quarantined message", "DlqRoutingKey", p.Config.DlqRoutingKey, "data", string(message.Data))
	return nil
}

func (p *Processor) CloseRabbit() {
	if err := p.RabbitChannel.Close(); err != nil {
		logger.Logger.Errorw("Error closing rabbit channel during shutdown of processor", "subscription", p.PubSubSubscription, "error", err)
	}
	if err := p.RabbitConn.Close(); err != nil {
		logger.Logger.Errorw("Error closing rabbit connection during shutdown of processor", "subscription", p.PubSubSubscription, "error", err)
	}
}
