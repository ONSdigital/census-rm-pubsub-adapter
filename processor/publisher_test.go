package processor

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestProcessor_Publish_Success(t *testing.T) {
	// Given
	testOutboundMessageChannel := make(chan *models.OutboundMessage)
	processor := &Processor{
		Logger:           zap.S(),
		Config:           config.TestConfig,
		OutboundMsgChan:  testOutboundMessageChannel,
		RabbitRoutingKey: "test-routing-key",
	}
	mockSourceMessage := new(MockPubSubMessage)
	mockSourceMessage.On("Ack").Return()

	mockChannel := new(MockRabbitChannel)
	mockChannel.On("Publish",
		processor.Config.EventsExchange,
		processor.RabbitRoutingKey,
		true,  // Mandatory
		false, // Immediate
		mock.Anything).Return(nil)
	mockChannel.On("TxCommit").Return(nil)

	outboundMessage := models.OutboundMessage{
		EventMessage: &models.RmMessage{
			Event:   models.RmEvent{},
			Payload: models.RmPayload{},
		},
		SourceMessage: mockSourceMessage,
	}

	// When
	// Send it a message to publish
	go func() { testOutboundMessageChannel <- &outboundMessage }()

	// Run the publisher within a timeout period
	timeout, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	processor.Publish(timeout, mockChannel)

	// Then
	mockChannel.AssertExpectations(t)
	mockSourceMessage.AssertExpectations(t)
}

func TestProcessor_Publish_PublishFailure(t *testing.T) {
	// Given

	testOutboundMessageChannel := make(chan *models.OutboundMessage)
	processor := &Processor{
		Logger:           zap.S(),
		Config:           config.TestConfig,
		OutboundMsgChan:  testOutboundMessageChannel,
		RabbitRoutingKey: "test-routing-key",
	}
	mockSourceMessage := new(MockPubSubMessage)
	mockSourceMessage.On("Nack").Return()

	mockChannel := new(MockRabbitChannel)
	mockChannel.On("Publish",
		processor.Config.EventsExchange,
		processor.RabbitRoutingKey,
		true,  // Mandatory
		false, // Immediate
		mock.Anything).Return(errors.New("Publishing failed"))
	mockChannel.On("TxRollback").Return(nil)

	outboundMessage := models.OutboundMessage{
		EventMessage: &models.RmMessage{
			Event:   models.RmEvent{},
			Payload: models.RmPayload{},
		},
		SourceMessage: mockSourceMessage,
	}

	// When
	// Send it a message to publish
	go func() { testOutboundMessageChannel <- &outboundMessage }()

	// Run the publisher within a timeout period
	timeout, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	processor.Publish(timeout, mockChannel)

	// Then
	mockChannel.AssertExpectations(t)
	mockSourceMessage.AssertExpectations(t)
}

func TestProcessor_Publish_CommitFailure(t *testing.T) {
	// Given
	testOutboundMessageChannel := make(chan *models.OutboundMessage)
	processor := &Processor{
		Logger:           zap.S(),
		Config:           config.TestConfig,
		OutboundMsgChan:  testOutboundMessageChannel,
		RabbitRoutingKey: "test-routing-key",
	}
	mockSourceMessage := new(MockPubSubMessage)
	mockSourceMessage.On("Nack").Return()

	mockChannel := new(MockRabbitChannel)
	mockChannel.On("Publish",
		processor.Config.EventsExchange,
		processor.RabbitRoutingKey,
		true,  // Mandatory
		false, // Immediate
		mock.Anything).Return(nil)
	mockChannel.On("TxCommit").Return(errors.New("Commit failed"))
	mockChannel.On("TxRollback").Return(nil)

	outboundMessage := models.OutboundMessage{
		EventMessage: &models.RmMessage{
			Event:   models.RmEvent{},
			Payload: models.RmPayload{},
		},
		SourceMessage: mockSourceMessage,
	}

	// When
	// Send it a message to publish
	go func() { testOutboundMessageChannel <- &outboundMessage }()

	// Run the publisher within a timeout period
	timeout, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	processor.Publish(timeout, mockChannel)

	// Give it a second to make the calls
	for len(mockSourceMessage.Calls) == 0 {
	}

	// Then
	mockChannel.AssertExpectations(t)
	mockSourceMessage.AssertExpectations(t)
}

// Mocks

type MockPubSubMessage struct {
	mock.Mock
}

func (m *MockPubSubMessage) Ack() {
	m.Called()
}
func (m *MockPubSubMessage) Nack() {
	m.Called()
}

type MockRabbitChannel struct {
	mock.Mock
}

func (m *MockRabbitChannel) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitChannel) Tx() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitChannel) TxCommit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitChannel) TxRollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func (m *MockRabbitChannel) NotifyClose(value chan *amqp.Error) chan *amqp.Error {
	m.Called(value)
	return value
}
