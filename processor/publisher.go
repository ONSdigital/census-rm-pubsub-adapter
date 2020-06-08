package processor

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type RabbitChannel interface {
	Close() error
	Tx() error
	TxCommit() error
	TxRollback() error
	NotifyClose(chan *amqp.Error) chan *amqp.Error
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
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

func (p *Processor) initRabbitChannel(cancelFunc context.CancelFunc) (RabbitChannel, error) {
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

	p.RabbitChannels = make([]RabbitChannel, 0)

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
		case <-ctx.Done():
			return
		}
	}
}

func (p *Processor) Publish(ctx context.Context, channel RabbitChannel) {
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

func (p *Processor) publishEventToRabbit(message *models.RmMessage, routingKey string, exchange string, channel RabbitChannel) error {

	byteMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if err := channel.Publish(
		exchange,
		routingKey,
		true,       // mandatory
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

func (p *Processor) CloseRabbit(errOk bool) {
	if p.RabbitConn != nil {
		if err := p.RabbitConn.Close(); err != nil && !errOk {
			p.Logger.Errorw("Error closing rabbit connection", "error", err)
		}
	}
}
