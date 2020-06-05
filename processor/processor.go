package processor

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/logger"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type messageUnmarshaller func([]byte) (models.PubSubMessage, error)

type messageConverter func(message models.PubSubMessage) (*models.RmMessage, error)

type Processor struct {
	RabbitConn         *amqp.Connection
	RabbitRoutingKey   string
	RabbitChannels     []*amqp.Channel
	OutboundMsgChan    chan *models.OutboundMessage
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
	p.OutboundMsgChan = make(chan *models.OutboundMessage)
	p.RabbitChannels = make([]*amqp.Channel, 0)
	p.Logger = logger.Logger.With("subscription", pubSubSubscription)

	// Setup PubSub connection
	p.PubSubClient, err = pubsub.NewClient(ctx, pubSubProject)
	if err != nil {
		return nil, errors.Wrap(err, "error settings up PubSub client")
	}

	// Setup subscription
	p.PubSubSubscription = p.PubSubClient.Subscription(pubSubSubscription)

	// Start consuming from PubSub
	p.Logger.Infow("Launching PubSub message receiver")
	go p.Consume(ctx)

	go p.ManagePublishers(ctx)

	return p, nil
}

func (p *Processor) initRabbitConnection() error {
	p.Logger.Debug("Initialising rabbit connection")
	var err error

	// Open the rabbit connection
	p.RabbitConn, err = amqp.Dial(p.Config.RabbitConnectionString)
	if err != nil {
		return errors.Wrap(err, "error connecting to rabbit")
	}
	return nil
}

func (p *Processor) initRabbitChannel(cancelFunc context.CancelFunc) (*amqp.Channel, error) {
	var err error
	var channel *amqp.Channel

	// Open the rabbit channel
	p.Logger.Debugf("Initialising rabbit channel no. %d", len(p.RabbitChannels)+1)
	channel, err = p.RabbitConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "error opening rabbit channel")
	}

	if err := channel.Tx(); err != nil {
		return nil, errors.Wrap(err, "Error making rabbit channel transactional")
	}

	// Set up handler to attempt to reopen channel on channel close
	// Listen for errors on the rabbit channel to handle both channel specific and connection wide exceptions
	channelErrChan := make(chan *amqp.Error)
	go func() {
		channelErr := <-channelErrChan

		// TODO handle reconnecting in graceful processor restart rather than a direct call to reinitialize here
		p.Logger.Errorw("received rabbit channel error", "error", channelErr)
		cancelFunc()
	}()
	channel.NotifyClose(channelErrChan)
	p.RabbitChannels = append(p.RabbitChannels, channel)

	return channel, nil
}

func (p *Processor) startPublishers(ctx context.Context, publisherCancel context.CancelFunc) {
	// Setup one rabbit connection
	if err := p.initRabbitConnection(); err != nil {
		p.Logger.Errorw("Error initialising rabbit connection", "error", err)
		publisherCancel()
		return
	}

	p.RabbitChannels = make([]*amqp.Channel, 0)

	for i := 0; i < p.Config.PublishersPerProcessor; i++ {

		// Open a rabbit channel for each publisher worker
		channel, err := p.initRabbitChannel(publisherCancel)
		if err != nil {
			return
		}
		go p.Publish(ctx, channel)
	}
}

func (p *Processor) ManagePublishers(ctx context.Context) {

	publisherCtx, publisherCancel := context.WithCancel(context.Background())
	p.startPublishers(ctx, publisherCancel)

	for {
		select {
		case <-publisherCtx.Done():
			// Use a publisher context to restart publishers on rabbit connection or channel errors
			// TODO replace this with graceful processor restarting
			publisherCtx, publisherCancel = context.WithCancel(context.Background())
			p.Logger.Info("Restarting publishers")

			// Tidy up the current connection and any open channels
			p.CloseRabbit(true)

			// Restart publishers
			p.startPublishers(ctx, publisherCancel)

			// Sleep for a second here so it doesn't bombard rabbit with reconnection attempts at a ridiculous rate
			time.Sleep(1*time.Second)
		case <-ctx.Done():
			return
		}
	}
}

func (p *Processor) Publish(ctx context.Context, channel *amqp.Channel) {
	for {
		select {
		case outboundMessage := <-p.OutboundMsgChan:

			ctxLogger := p.Logger.With("transactionId", outboundMessage.EventMessage.Event.TransactionID)
			if err := p.publishEventToRabbit(outboundMessage.EventMessage, p.RabbitRoutingKey, p.Config.EventsExchange, channel); err != nil {
				ctxLogger.Errorw("Failed to publish message", "error", err)
				outboundMessage.SourceMessage.Nack()
				if err := channel.TxRollback(); err != nil {
					ctxLogger.Errorw("Error rolling back rabbit transaction after failed message publish", "error", err)
				}
				continue
			}
			if err := channel.TxCommit(); err != nil {
				ctxLogger.Errorw("Failed to commit transaction to publish message", "error", err)
				outboundMessage.SourceMessage.Nack()
				if err := channel.TxRollback(); err != nil {
					ctxLogger.Errorw("Error rolling back rabbit transaction", "error", err)
				}
				continue
			}

			outboundMessage.SourceMessage.Ack()

		case <-ctx.Done():
			return
		}
	}
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
		if err := p.quarantineMessage(msg); err != nil {
			ctxLogger.Errorw("Error quarantining bad message, nacking", "error", err, "data", string(msg.Data))
			msg.Nack()
			return
		}
		ctxLogger.Debugw("Acking quarantined message", "msgData", string(msg.Data))
		msg.Ack()
		return
	}
	ctxLogger = ctxLogger.With("transactionId", messageReceived.GetTransactionId())
	ctxLogger.Debugw("Processing message")
	rmMessageToSend, err := p.convertMessage(messageReceived)
	if err != nil {
		ctxLogger.Errorw("Error converting message", "error", err)
		return
	}
	ctxLogger.Debugw("Sending outbound message to publish", "msgData", string(msg.Data))
	p.OutboundMsgChan <- &models.OutboundMessage{SourceMessage: msg, EventMessage: rmMessageToSend}
}

func (p *Processor) publishEventToRabbit(message *models.RmMessage, routingKey string, exchange string, channel *amqp.Channel) error {

	byteMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if err := channel.Publish(
		exchange,
		routingKey, // routing key (the queue)
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         byteMessage,
			DeliveryMode: 2, // 2 = persistent delivery mode
		}); err != nil {
		return err
	}

	p.Logger.Debugw("Published message", "routingKey", routingKey, "transactionId", message.Event.TransactionID)
	return nil
}

func (p *Processor) quarantineMessage(message *pubsub.Message) error {
	headers := map[string]string{
		"pubSubId": message.ID,
	}

	for key, value := range message.Attributes {
		headers[key] = value
	}

	msgToQuarantine := models.MessageToQuarantine{
		MessageHash:    fmt.Sprintf("%x", sha256.Sum256(message.Data)),
		MessagePayload: message.Data,
		Service:        "Pubsub Adapter",
		Queue:          p.PubSubSubscription.ID(),
		ExceptionClass: "Error unmarshalling message",
		RoutingKey:     "none",
		ContentType:    "application/json",
		Headers:        headers,
	}

	jsonValue, err := json.Marshal(msgToQuarantine)
	if err != nil {
		return err
	}

	_, err = http.Post(p.Config.QuarantineMessageUrl, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	p.Logger.Debugw("Quarantined message", "data", string(message.Data))
	return nil
}

func (p *Processor) CloseRabbit(errOk bool) {
	for _, channel := range p.RabbitChannels {
		if err := channel.Close(); err != nil && !errOk {
			p.Logger.Errorw("Error closing rabbit channel", "error", err)
		}
	}

	if p.RabbitConn != nil {
		if err := p.RabbitConn.Close(); err != nil && !errOk {
			p.Logger.Errorw("Error closing rabbit connection ", "error", err)
		}
	}
}
