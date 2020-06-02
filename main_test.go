// +build !unitTest

package main

// This test requires dependencies to be running

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"errors"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/streadway/amqp"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
		PublishersPerProcessor:          1,
		RabbitConnectionString:          "amqp://guest:guest@localhost:7672/",
		ReceiptRoutingKey:               "goTestReceiptQueue",
		UndeliveredRoutingKey:           "goTestUndeliveredQueue",
		FulfilmentRoutingKey:            "goTestFulfilmentConfirmedQueue",
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

		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// When
		if messageId, err := publishMessageToPubSub(timeout, messageToSend, topic, project); err != nil {
			t.Errorf("PubSub publish fail, project: %s, topic: %s, id: %s, error: %s", project, topic, messageId, err)
			return
		}

		rabbitMessage, err := getFirstMessageOnQueue(timeout, rabbitRoutingKey, rabbitCh)
		if err != nil {
			t.Error(err)
			return
		}

		// Then
		if rabbitMessage != expectedRabbitMessage {
			t.Errorf("Rabbit messsage incorrect - \nexpected: %s \nactual: %s", expectedRabbitMessage, rabbitMessage)
		}
	}
}

func TestMessageQuarantiningBadJson(t *testing.T) {
	testMessageQuarantining("bad_message", "Test bad non JSON message is quarantined", t)
}

func TestMessageQuarantiningMissingTxnId(t *testing.T) {
	testMessageQuarantining(`{"thisMessage": "is_missing_tx_id"}`, "Test bad message missing transaction ID is quarantined", t)
}

func testMessageQuarantining(messageToSend string, testDescription string, t *testing.T) {
	var requests []*http.Request
	var requestBody []byte

	mockResult := "Success!"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(mockResult))
		requests = append(requests, r)
		requestBody, _ = ioutil.ReadAll(r.Body)
	}))
	defer srv.Close()

	cfg.QuarantineMessageUrl = srv.URL

	t.Run(testDescription, testMessageProcessingQuarantine(
		messageToSend,
		cfg.EqReceiptTopic, cfg.EqReceiptProject))

	// Allow a second for the processor to process the message
	time.Sleep(1 * time.Second)

	if len(requests) != 1 {
		t.Errorf("Unexpected number of calls to Exception Manager")
		return
	}

	var quarantineBody models.MessageToQuarantine
	err := json.Unmarshal(requestBody, &quarantineBody)
	if err != nil {
		t.Errorf("Could not decode request body sent to Exception Manager")
		return
	}

	if !assertEqual("application/json", quarantineBody.ContentType, "Dodgy content type", t) ||
		!assertEqual("Error unmarshalling message", quarantineBody.ExceptionClass, "Dodgy exception class", t) ||
		!assertEqual(1, len(quarantineBody.Headers), "Dodgy headers", t) ||
		!assertEqual(64, len(quarantineBody.MessageHash), "Dodgy message hash", t) ||
		!assertEqual(messageToSend, string(quarantineBody.MessagePayload), "Dodgy message payload", t) ||
		!assertEqual(cfg.EqReceiptSubscription, quarantineBody.Queue, "Dodgy quarantine queue", t) ||
		!assertEqual("none", quarantineBody.RoutingKey, "Dodgy routing key", t) ||
		!assertEqual("Pubsub Adapter", quarantineBody.Service, "Dodgy routing key", t) {
		return
	}

	if len(quarantineBody.Headers["pubSubId"]) == 0 {
		t.Errorf("Dodgy pubSubId header, expected non-zero length")
		return
	}
}

func assertEqual(expected interface{}, actual interface{}, feedback string, t *testing.T) bool {
	if expected != actual {
		t.Errorf("%v, expected %v, actual %v", feedback, expected, actual)
		return false
	}

	return true
}

func testMessageProcessingQuarantine(messageToSend string, topic string, project string) func(t *testing.T) {
	return func(t *testing.T) {
		if _, err := StartProcessors(ctx, cfg, make(chan error)); err != nil {
			t.Error(err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// When
		if messageId, err := publishMessageToPubSub(ctx, messageToSend, topic, project); err != nil {
			t.Errorf("PubSub publish fail, project: %s, topic: %s, id: %s, error: %s", project, topic, messageId, err)
			return
		}

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

func TestRabbitReconnect(t *testing.T) {
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start up the processors normally
	processors, err := StartProcessors(ctx, cfg, make(chan error))
	if err != nil {
		t.Error(err)
		return
	}

	// Take the first processor
	processor := processors[0]

	// Pick one of the processors rabbit channels
	var channel *amqp.Channel
	for channel == nil {
		select {
		case <-timeout.Done():
			t.Error()
			return
		default:
			if len(processor.RabbitChannels) > 0 {
				channel = processor.RabbitChannels[0]
			}
		}
	}
	channelErrChan := make(chan *amqp.Error)
	channel.NotifyClose(channelErrChan)

	// Check the processors rabbit channel can publish
	if err := publishToRabbit(channel, cfg.EventsExchange, cfg.ReceiptRoutingKey, `{"test":"message should publish before"}`); err != nil {
		t.Error(err)
		return
	}

	// Break the channel by publishing a mandatory message that can't be routed
	// NB: This is not a typical scenario this feature is designed around as the app or rabbit would have to be
	// mis-configured for this to occur, and the channel closing is only an undesirable side effect.
	// It is, however, the only viable way of inducing a channel close that I could think of using to exercise this code.
	if err := publishToRabbit(channel, "this_exchange_should_not_exist", cfg.ReceiptRoutingKey, `{"test":"message should fail"}`); err != nil {
		t.Error(err)
		return
	}

	// Wait for the unpublishable message to kill the channel with a timeout
	select {
	case <-timeout.Done():
		t.Error("Timed out waiting for induced rabbit channel closure")
	case <-channelErrChan:
	}

	// Try to successfully publish a message within the timeout
	success := make(chan bool)
	go func() {
		for {
			select {
			case <-timeout.Done():
				// Kill this goroutine if the test times out
				return
			default:
				// Repeatedly try to publish a message using the processors channel
				if len(processor.RabbitChannels) == 0 {
					continue
				}
				channel := processor.RabbitChannels[0]
				if err := publishToRabbit(channel, cfg.EventsExchange, cfg.ReceiptRoutingKey, `{"test":"message should publish after"}`); err == nil {
					// We have successfully published a message with the processors re-opened rabbit channel
					success <- true
					return
				}
			}
		}
	}()

	select {
	case <-timeout.Done():
		t.Error("Failed to publish message with processors channel within the timeout")
		return
	case <-success:
		return
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

func publishToRabbit(channel *amqp.Channel, exchange string, routingKey string, message string) error {
	return channel.Publish(
		exchange,
		routingKey,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         []byte(message),
			DeliveryMode: 2, // 2 = persistent delivery mode
		})
}
