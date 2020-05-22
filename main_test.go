// +build !unitTest

package main

// This test requires dependencies to be running

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/streadway/amqp"
	"os"
	"runtime"
	"testing"
	"time"
)

var ctx context.Context
var cfg *config.Configuration

func TestMain(m *testing.M) {
	// These tests interact with data in backing services so cannot be run in parallel
	runtime.GOMAXPROCS(1)
	ctx = context.Background()
	cfg = &config.Configuration{
		RabbitConnectionString:          "amqp://guest:guest@localhost:7672/",
		ReceiptRoutingKey:               "goTestReceiptQueue",
		UndeliveredRoutingKey:           "goTestUndeliveredQueue",
		FulfilmentRoutingKey:            "goTestFulfilmentConfirmedQueue",
		DlqRoutingKey:                   "goTestQuarantineQueue",
		EqReceiptProject:                "project",
		EqReceiptSubscription:           "rm-receipt-subscription",
		EqReceiptTopic:                  "eq-submission-topic",
		OfflineReceiptProject:           "offline-project",
		OfflineReceiptSubscription:      "rm-offline-receipt-subscription",
		OfflineReceiptTopic:             "offline-receipt-topic",
		PpoUndeliveredProject:           "ppo-undelivered-project",
		PpoUndeliveredTopic:             "ppo-undelivered-mail-topic",
		PpoUndeliveredSubscription:      "rm-ppo-undelivered-subscription",
		QmUndeliveredProject:            "qm-undelivered-project",
		QmUndeliveredTopic:              "qm-undelivered-mail-topic",
		QmUndeliveredSubscription:       "rm-qm-undelivered-subscription",
		FulfilmentConfirmedProject:      "fulfilment-confirmed-project",
		FulfilmentConfirmedSubscription: "fulfilment-subscription",
		FulfilmentConfirmedTopic:        "fulfilment-topic",
	}
	code := m.Run()
	os.Exit(code)
}

func TestMessageProcessing(t *testing.T) {
	t.Run("Test EQ receipting", testMessageProcessing(
		`{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`,
		`{"event":{"type":"RESPONSE_RECEIVED","source":"RECEIPT_SERVICE","channel":"EQ","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"response":{"questionnaireId":"01213213213","unreceipt":false}}}`,
		cfg.EqReceiptTopic, cfg.EqReceiptProject, cfg.ReceiptRoutingKey))

	t.Run("Test Offline receipting", testMessageProcessing(
		`{"dateTime": "2008-08-24T00:00:00", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
		`{"event":{"type":"RESPONSE_RECEIVED","source":"RECEIPT_SERVICE","channel":"INTEGRATION_TEST","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"response":{"questionnaireId":"01213213213","unreceipt":false}}}`,
		cfg.OfflineReceiptTopic, cfg.OfflineReceiptProject, cfg.ReceiptRoutingKey))

	t.Run("Test PPO undelivered mail", testMessageProcessing(
		`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "caseRef": "0123456789", "productCode": "P_TEST_1"}`,
		`{"event":{"type":"UNDELIVERED_MAIL_REPORTED","source":"RECEIPT_SERVICE","channel":"PPO","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"fulfilmentInformation":{"caseRef":"0123456789","fulfilmentCode":"P_TEST_1"}}}`,
		cfg.PpoUndeliveredTopic, cfg.PpoUndeliveredProject, cfg.UndeliveredRoutingKey))

	t.Run("Test QM undelivered mail", testMessageProcessing(
		`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
		`{"event":{"type":"UNDELIVERED_MAIL_REPORTED","source":"RECEIPT_SERVICE","channel":"QM","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"fulfilmentInformation":{"questionnaireId":"01213213213"}}}`,
		cfg.QmUndeliveredTopic, cfg.QmUndeliveredProject, cfg.UndeliveredRoutingKey))

	t.Run("Test QM fulfilment confirmation", testMessageProcessing(
		`{"dateTime":"2019-08-03T14:30:01","questionnaireId":"1100000000112","productCode":"P_OR_H1","channel":"QM","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
		`{"event":{"type":"FULFILMENT_CONFIRMED","source":"RECEIPT_SERVICE","channel":"QM","dateTime":"2019-08-03T14:30:01Z","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"},"payload":{"fulfilmentInformation":{"fulfilmentCode":"P_OR_H1","questionnaireId":"1100000000112"}}}`,
		cfg.FulfilmentConfirmedTopic, cfg.FulfilmentConfirmedProject, cfg.FulfilmentRoutingKey))

	t.Run("Test PPO fulfilment confirmation", testMessageProcessing(
		`{"dateTime":"2019-08-03T14:30:01","caseRef":"12345678","productCode":"P_OR_H1","channel":"PPO","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
		`{"event":{"type":"FULFILMENT_CONFIRMED","source":"RECEIPT_SERVICE","channel":"PPO","dateTime":"2019-08-03T14:30:01Z","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"},"payload":{"fulfilmentInformation":{"caseRef":"12345678","fulfilmentCode":"P_OR_H1"}}}`,
		cfg.FulfilmentConfirmedTopic, cfg.FulfilmentConfirmedProject, cfg.FulfilmentRoutingKey))
}

func TestMessageQuarantining(t *testing.T) {
	t.Run("Test bad non JSON message is quarantined", testMessageProcessing(
		`bad_message`,
		`bad_message`,
		cfg.EqReceiptTopic, cfg.EqReceiptProject, cfg.DlqRoutingKey))

	t.Run("Test bad message missing transaction ID is quarantined", testMessageProcessing(
		`{"thisMessage": "is_missing_tx_id"}`,
		`{"thisMessage": "is_missing_tx_id"}`,
		cfg.EqReceiptTopic, cfg.EqReceiptProject, cfg.DlqRoutingKey))
}

func testMessageProcessing(messageToSend string, expectedRabbitMessage string, topic string, project string, rabbitRoutingKey string) func(t *testing.T) {
	return func(t *testing.T) {
		if _, err := StartProcessors(ctx, cfg, make(chan error)); err != nil {
			t.Error(err)
			return
		}

		rabbitConn, rabbitCh, err := connectToRabbitChannel()
		defer rabbitCh.Close()
		defer rabbitConn.Close()
		if _, err := rabbitCh.QueuePurge(rabbitRoutingKey, false); err != nil {
			t.Error(err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// When
		if messageId, err := publishMessageToPubSub(ctx, messageToSend, topic, project); err != nil {
			t.Errorf("PubSub publish fail, project: %s, topic: %s, id: %s, error: %s", project, topic, messageId, err)
			return
		}

		rabbitMessage, err := getFirstMessageOnQueue(ctx, rabbitRoutingKey, rabbitCh)
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
}

func TestStartProcessors(t *testing.T) {
	processors, err := StartProcessors(ctx, cfg, make(chan error))
	if err != nil {
		t.Error(err)
		return
	}

	if len(processors) != 5 {
		t.Errorf("StartProcessors should return 5 processors, actually returned %d", len(processors))
	}
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
