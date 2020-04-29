// +build !unitTest

package main

// This test requires dependencies to be running

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/processor"
	"github.com/streadway/amqp"
	"os"
	"testing"
	"time"
)

var ctx context.Context
var cfg *config.Configuration

func TestMain(m *testing.M) {
	ctx = context.Background()
	cfg = &config.Configuration{
		RabbitConnectionString:     "amqp://guest:guest@localhost:7672/",
		ReceiptRoutingKey:          "goTestReceiptQueue",
		UndeliveredRoutingKey:      "goTestUndeliveredQueue",
		EqReceiptProject:           "project",
		EqReceiptSubscription:      "rm-receipt-subscription",
		EqReceiptTopic:             "eq-submission-topic",
		OfflineReceiptProject:      "offline-project",
		OfflineReceiptSubscription: "rm-offline-receipt-subscription",
		OfflineReceiptTopic:        "offline-receipt-topic",
		PpoUndeliveredProject:      "ppo-undelivered-project",
		PpoUndeliveredTopic:        "ppo-undelivered-mail-topic",
		PpoUndeliveredSubscription: "rm-ppo-undelivered-subscription",
		QmUndeliveredProject:       "qm-undelivered-project",
		QmUndeliveredTopic:         "qm-undelivered-mail-topic",
		QmUndeliveredSubscription:  "rm-qm-undelivered-subscription",
	}
	code := m.Run()
	os.Exit(code)
}

func TestEqReceipt(t *testing.T) {
	// Given
	eqReceiptMsg := `{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`
	expectedRabbitMessage := `{"event":{"type":"RESPONSE_RECEIVED","source":"RECEIPT_SERVICE","channel":"EQ","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"response":{"questionnaireId":"01213213213","unreceipt":false}}}`

	eqReceiptProcessor := processor.NewEqReceiptProcessor(ctx, cfg)

	rabbitConn, rabbitChan, err := connectToRabbitChannel()
	defer rabbitConn.Close()
	defer rabbitChan.Close()
	if _, err := rabbitChan.QueuePurge(cfg.ReceiptRoutingKey, true); err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// When
	if messageId, err := publishMessageToPubSub(ctx, eqReceiptMsg, cfg.EqReceiptTopic, cfg.EqReceiptProject); err != nil {
		t.Errorf("PubSub publish fail, project: %s, topic: %s, id: %s, error: %s", cfg.EqReceiptProject, cfg.EqReceiptTopic, messageId, err)
		return
	}

	go eqReceiptProcessor.Consume(ctx)
	go eqReceiptProcessor.Process(ctx)

	rabbitMessage, err := getFirstMessageOnQueue(ctx, cfg.ReceiptRoutingKey, rabbitChan)
	if err != nil {
		t.Error(err)
		return
	}

	// Then
	if rabbitMessage != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessage)
	}
	cancel()
}

func TestOfflineReceipt(t *testing.T) {
	// Given
	offlineReceiptMsg := `{"dateTime": "2008-08-24T00:00:00Z", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`
	expectedRabbitMessage := `{"event":{"type":"RESPONSE_RECEIVED","source":"RECEIPT_SERVICE","channel":"INTEGRATION_TEST","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"response":{"questionnaireId":"01213213213","unreceipt":false}}}`

	offlineReceiptProcessor := processor.NewOfflineReceiptProcessor(ctx, cfg)

	rabbitConn, rabbitChan, err := connectToRabbitChannel()
	defer rabbitConn.Close()
	defer rabbitChan.Close()
	if _, err := rabbitChan.QueuePurge(cfg.ReceiptRoutingKey, true); err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// When
	if messageId, err := publishMessageToPubSub(ctx, offlineReceiptMsg, cfg.OfflineReceiptTopic, cfg.OfflineReceiptProject); err != nil {
		t.Errorf("pubsub publish fail, project: %s, topic: %s, id: %s, error: %s", cfg.OfflineReceiptProject, cfg.OfflineReceiptTopic, messageId, err)
		return
	}

	go offlineReceiptProcessor.Consume(ctx)
	go offlineReceiptProcessor.Process(ctx)

	rabbitMessage, err := getFirstMessageOnQueue(ctx, cfg.ReceiptRoutingKey, rabbitChan)
	if err != nil {
		t.Error(err)
		return
	}

	// Then
	if rabbitMessage != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessage)
	}
	cancel()
}

func TestPpoUndelivered(t *testing.T) {
	// Given
	ppoUndeliveredMsg := `{"dateTime": "2008-08-24T00:00:00Z", "transactionId": "abc123xxx", "caseRef": "0123456789", "productCode": "P_TEST_1"}`
	expectedRabbitMessage := `{"event":{"type":"UNDELIVERED_MAIL_REPORTED","source":"RECEIPT_SERVICE","channel":"PPO","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"fulfilmentInformation":{"caseRef":"0123456789","fulfilmentCode":"P_TEST_1"}}}`

	ppoUndeliveredProcessor := processor.NewPpoUndeliveredProcessor(ctx, cfg)

	rabbitConn, rabbitChan, err := connectToRabbitChannel()
	defer rabbitConn.Close()
	defer rabbitChan.Close()
	if _, err := rabbitChan.QueuePurge(cfg.UndeliveredRoutingKey, true); err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// When
	if messageId, err := publishMessageToPubSub(ctx, ppoUndeliveredMsg, cfg.PpoUndeliveredTopic, cfg.PpoUndeliveredProject); err != nil {
		t.Errorf("pubsub publish fail, project: %s, topic: %s, id: %s, error: %s", cfg.PpoUndeliveredProject, cfg.PpoUndeliveredTopic, messageId, err)
		return
	}

	go ppoUndeliveredProcessor.Consume(ctx)
	go ppoUndeliveredProcessor.Process(ctx)

	rabbitMessage, err := getFirstMessageOnQueue(ctx, cfg.UndeliveredRoutingKey, rabbitChan)
	if err != nil {
		t.Error(err)
		return
	}

	// Then
	if rabbitMessage != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessage)
	}
	cancel()
}

func TestQmUndelivered(t *testing.T) {
	// Given
	qmUndeliveredMsg := `{"dateTime": "2008-08-24T00:00:00Z", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`
	expectedRabbitMessage := `{"event":{"type":"UNDELIVERED_MAIL_REPORTED","source":"RECEIPT_SERVICE","channel":"QM","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"fulfilmentInformation":{"questionnaireId":"01213213213"}}}`

	qmUndeliveredProcessor := processor.NewQmUndeliveredProcessor(ctx, cfg)

	rabbitConn, rabbitChan, err := connectToRabbitChannel()
	defer rabbitConn.Close()
	defer rabbitChan.Close()
	if _, err := rabbitChan.QueuePurge(cfg.UndeliveredRoutingKey, true); err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// When
	if messageId, err := publishMessageToPubSub(ctx, qmUndeliveredMsg, cfg.QmUndeliveredTopic, cfg.QmUndeliveredProject); err != nil {
		t.Errorf("pubsub publish fail, project: %s, topic: %s, id: %s, error: %s", cfg.QmUndeliveredProject, cfg.QmUndeliveredTopic, messageId, err)
		return
	}

	go qmUndeliveredProcessor.Consume(ctx)
	go qmUndeliveredProcessor.Process(ctx)

	rabbitMessage, err := getFirstMessageOnQueue(ctx, cfg.UndeliveredRoutingKey, rabbitChan)
	if err != nil {
		t.Error(err)
		return
	}

	// Then
	if rabbitMessage != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessage)
	}
	cancel()
}

func publishMessageToPubSub(ctx context.Context, msg string, topic string, project string) (id string, err error) {
	pubSubClient, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return "", err
	}
	eqReceiptTopic := pubSubClient.Topic(topic)

	result := eqReceiptTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// Block until the result is returned and eqReceiptProcessor server-generated
	// ID is returned for the published message.
	return result.Get(ctx)
}

func getFirstMessageOnQueue(ctx context.Context, queue string, ch *amqp.Channel) (message string, err error) {
	msgChan, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return "", err
	}

	select {
	case d := <-msgChan:
		message = string(d.Body)
		return message, nil
	case <-ctx.Done():
		return "", errors.New("timed out waiting for the rabbit message")
	}
}

func connectToRabbitChannel() (conn *amqp.Connection, ch *amqp.Channel, err error) {
	rabbitConn, err := amqp.Dial(cfg.RabbitConnectionString)
	if err != nil {
		return nil, nil, err
	}

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return rabbitConn, rabbitChan, nil
}
