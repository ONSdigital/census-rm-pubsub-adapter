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

	// Start and manage publisher workers
	go p.ManagePublishers(ctx)

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
