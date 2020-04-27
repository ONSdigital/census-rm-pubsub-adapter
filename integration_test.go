// +build !unitTest

package main

// This test requires dependencies to be running

import (
	"cloud.google.com/go/pubsub"
	"context"
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
		ReceiptRoutingKey:          "goTestQueue",
		EqReceiptProject:           "project",
		EqReceiptSubscription:      "rm-receipt-subscription",
		EqReceiptTopic:             "eq-submission-topic",
		OfflineReceiptProject:      "offline-project",
		OfflineReceiptSubscription: "rm-offline-receipt-subscription",
		OfflineReceiptTopic:        "offline-receipt-topic",
		PpoUndeliveredProject:      "ppo-undelivered-project",
		PpoUndeliveredTopic:        "ppo-undelivered-mail-topic",
		PpoUndeliveredSubscription: "rm-ppo-undelivered-subscription",
		QmUndeliveredProject:      "qm-undelivered-project",
		QmUndeliveredTopic:        "qm-undelivered-mail-topic",
		QmUndeliveredSubscription: "rm-qm-undelivered-subscription",
	}
	code := m.Run()
	os.Exit(code)
}

func TestEqReceipt(t *testing.T) {

	eqReceiptMsg := `{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`
	expectedRabbitMessage := `{"event":{"type":"RESPONSE_RECEIVED","source":"RECEIPT_SERVICE","channel":"EQ","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"response":{"questionnaireId":"01213213213","unreceipt":false}}}`

	eqReceiptProcessor := processor.NewEqReceiptProcessor(ctx, cfg)
	rabbitConn, err := amqp.Dial("amqp://guest:guest@localhost:7672/")
	if err != nil {
		t.Errorf("Rabbit dial fail: %s", err)
	}
	defer rabbitConn.Close()

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		t.Errorf("Rabbit Channel Fail: %s", err)
	}
	defer rabbitChan.Close()
	rabbitChan.QueuePurge("goTestQueue", true)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pubSubClient, err := pubsub.NewClient(ctx, cfg.EqReceiptProject)
	eqReceiptTopic := pubSubClient.Topic(cfg.EqReceiptTopic)

	result := eqReceiptTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(eqReceiptMsg),
	})
	// Block until the result is returned and eqReceiptProcessor server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		t.Errorf("Pubsub test message %s publish failed. %s", id, err)
	}

	go eqReceiptProcessor.Consume(ctx)
	go eqReceiptProcessor.Process(ctx)

	msgs, err := rabbitChan.Consume("goTestQueue", "", false, false, false, false, nil)
	if err != nil {
		t.Errorf("Rabbit consume failed: %s", err)
	}

	var rabbitMessages []string
	select {
	case d := <-msgs:
		rabbitMessages = append(rabbitMessages, string(d.Body))
	case <-ctx.Done():
		t.Errorf("Timed out waiting for the rabbit message")
		return
	}

	if rabbitMessages[0] != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessages[0])
	}
	cancel()
}

func TestOfflineReceipt(t *testing.T) {

	eqReceiptMsg := `{"dateTime": "2008-08-24T00:00:00Z", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "tx_id": "abc123xxx", "questionnaire_id": "01213213213"}`
	expectedRabbitMessage := `{"event":{"type":"RESPONSE_RECEIVED","source":"RECEIPT_SERVICE","channel":"INTEGRATION_TEST","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"response":{"questionnaireId":"01213213213","unreceipt":false}}}`

	offlineReceiptProcessor := processor.NewOfflineReceiptProcessor(ctx, cfg)
	rabbitConn, err := amqp.Dial("amqp://guest:guest@localhost:7672/")
	if err != nil {
		t.Errorf("Rabbit dial fail: %s", err)
	}
	defer rabbitConn.Close()

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		t.Errorf("Rabbit Channel Fail: %s", err)
	}
	defer rabbitChan.Close()
	rabbitChan.QueuePurge("goTestQueue", true)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pubSubClient, err := pubsub.NewClient(ctx, cfg.OfflineReceiptProject)
	eqReceiptTopic := pubSubClient.Topic(cfg.OfflineReceiptTopic)

	result := eqReceiptTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(eqReceiptMsg),
	})
	// Block until the result is returned and offlineReceiptProcessor server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		t.Errorf("Pubsub test message %s publish failed. %s", id, err)
	}

	go offlineReceiptProcessor.Consume(ctx)
	go offlineReceiptProcessor.Process(ctx)

	msgs, err := rabbitChan.Consume("goTestQueue", "", false, false, false, false, nil)
	if err != nil {
		t.Errorf("Rabbit consume failed: %s", err)
	}

	var rabbitMessages []string
	select {
	case d := <-msgs:
		rabbitMessages = append(rabbitMessages, string(d.Body))
	case <-ctx.Done():
		t.Errorf("Timed out waiting for the rabbit message")
		return
	}

	if rabbitMessages[0] != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessages[0])
	}
	cancel()
}

func TestPpoUndelivered(t *testing.T) {

	eqReceiptMsg := `{"dateTime": "2008-08-24T00:00:00Z", "transactionId": "abc123xxx", "caseRef": "0123456789", "productCode": "P_TEST_1"}`
	expectedRabbitMessage := `{"event":{"type":"UNDELIVERED_MAIL_REPORTED","source":"RECEIPT_SERVICE","channel":"PPO","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"fulfilmentInformation":{"caseRef":"0123456789","fulfilmentCode":"P_TEST_1"}}}`

	ppoUndeliveredProcessor := processor.NewPpoUndeliveredProcessor(ctx, cfg)
	rabbitConn, err := amqp.Dial("amqp://guest:guest@localhost:7672/")
	if err != nil {
		t.Errorf("Rabbit dial fail: %s", err)
	}
	defer rabbitConn.Close()

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		t.Errorf("Rabbit Channel Fail: %s", err)
	}
	defer rabbitChan.Close()
	rabbitChan.QueuePurge("goTestQueue", true)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pubSubClient, err := pubsub.NewClient(ctx, cfg.PpoUndeliveredProject)
	eqReceiptTopic := pubSubClient.Topic(cfg.PpoUndeliveredTopic)

	result := eqReceiptTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(eqReceiptMsg),
	})
	// Block until the result is returned and ppoUndeliveredProcessor server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		t.Errorf("Pubsub test message %s publish failed. %s", id, err)
	}

	go ppoUndeliveredProcessor.Consume(ctx)
	go ppoUndeliveredProcessor.Process(ctx)

	msgs, err := rabbitChan.Consume("goTestQueue", "", false, false, false, false, nil)
	if err != nil {
		t.Errorf("Rabbit consume failed: %s", err)
	}

	var rabbitMessages []string
	select {
	case d := <-msgs:
		rabbitMessages = append(rabbitMessages, string(d.Body))
	case <-ctx.Done():
		t.Errorf("Timed out waiting for the rabbit message")
		return
	}

	if rabbitMessages[0] != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessages[0])
	}
	cancel()
}

func TestQmUndelivered(t *testing.T) {

	eqReceiptMsg := `{"dateTime": "2008-08-24T00:00:00Z", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`
	expectedRabbitMessage := `{"event":{"type":"UNDELIVERED_MAIL_REPORTED","source":"RECEIPT_SERVICE","channel":"QM","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"fulfilmentInformation":{"questionnaireId":"01213213213"}}}`

	qmUndeliveredProcessor := processor.NewQmUndeliveredProcessor(ctx, cfg)
	rabbitConn, err := amqp.Dial("amqp://guest:guest@localhost:7672/")
	if err != nil {
		t.Errorf("Rabbit dial fail: %s", err)
	}
	defer rabbitConn.Close()

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		t.Errorf("Rabbit Channel Fail: %s", err)
	}
	defer rabbitChan.Close()
	rabbitChan.QueuePurge("goTestQueue", true)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pubSubClient, err := pubsub.NewClient(ctx, cfg.QmUndeliveredProject)
	eqReceiptTopic := pubSubClient.Topic(cfg.QmUndeliveredTopic)

	result := eqReceiptTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(eqReceiptMsg),
	})
	// Block until the result is returned and qmUndeliveredProcessor server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		t.Errorf("Pubsub test message %s publish failed. %s", id, err)
	}

	go qmUndeliveredProcessor.Consume(ctx)
	go qmUndeliveredProcessor.Process(ctx)

	msgs, err := rabbitChan.Consume("goTestQueue", "", false, false, false, false, nil)
	if err != nil {
		t.Errorf("Rabbit consume failed: %s", err)
	}

	var rabbitMessages []string
	select {
	case d := <-msgs:
		rabbitMessages = append(rabbitMessages, string(d.Body))
	case <-ctx.Done():
		t.Errorf("Timed out waiting for the rabbit message")
		return
	}

	if rabbitMessages[0] != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessages[0])
	}
	cancel()
}
