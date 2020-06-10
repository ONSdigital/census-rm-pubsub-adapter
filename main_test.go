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
	"github.com/ONSdigital/census-rm-pubsub-adapter/processor"
	"github.com/ONSdigital/census-rm-pubsub-adapter/readiness"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
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
	cfg = config.TestConfig
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
		cfg.FulfilmentConfirmedTopic, cfg.FulfilmentConfirmedProject, cfg.FulfilmentConfirmationRoutingKey))

	t.Run("Test PPO fulfilment confirmation", testMessageProcessing(
		`{"dateTime":"2019-08-03T14:30:01","caseRef":"12345678","productCode":"P_OR_H1","channel":"PPO","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
		`{"event":{"type":"FULFILMENT_CONFIRMED","source":"RECEIPT_SERVICE","channel":"PPO","dateTime":"2019-08-03T14:30:01Z","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"},"payload":{"fulfilmentInformation":{"caseRef":"12345678","fulfilmentCode":"P_OR_H1"}}}`,
		cfg.FulfilmentConfirmedTopic, cfg.FulfilmentConfirmedProject, cfg.FulfilmentConfirmationRoutingKey))

	t.Run("Test EQ fulfilment request", testMessageProcessing(
		`{
		"event" : {
			"type" : "FULFILMENT_REQUESTED",
			"source" : "QUESTIONNAIRE_RUNNER",
			"channel" : "EQ",
			"dateTime" : "2011-08-12T20:17:46.384Z",
			"transactionId" : "c45de4dc-3c3b-11e9-b210-d663bd873d93"
		},
		"payload" : {
			"fulfilmentRequest" : {
				"fulfilmentCode": "UACIT1",
				"caseId" : "bbd55984-0dbf-4499-bfa7-0aa4228700e9",
				"individualCaseId" : "8e8ebf71-d9c6-4efa-a693-ae24e7116e98",
				"contact": {
					"telNo":"+447890000000"
				}
			}
		}
	}`,
		`{"event":{"type":"FULFILMENT_REQUESTED","source":"QUESTIONNAIRE_RUNNER","channel":"EQ","dateTime":"2011-08-12T20:17:46.384Z","transactionId":"c45de4dc-3c3b-11e9-b210-d663bd873d93"},"payload":{"fulfilmentRequest":{"fulfilmentCode":"UACIT1","caseId":"bbd55984-0dbf-4499-bfa7-0aa4228700e9","individualCaseId":"8e8ebf71-d9c6-4efa-a693-ae24e7116e98","contact":{"telNo":"+447890000000"}}}}`,
		cfg.EqFulfilmentTopic, cfg.EqFulfilmentProject, cfg.FulfilmentRequestRoutingKey))

	t.Run("Test EQ fulfilment request empty contact", testMessageProcessing(
		`{
		"event" : {
			"type" : "FULFILMENT_REQUESTED",
			"source" : "QUESTIONNAIRE_RUNNER",
			"channel" : "EQ",
			"dateTime" : "2011-08-12T20:17:46.384Z",
			"transactionId" : "c45de4dc-3c3b-11e9-b210-d663bd873d93"
		},
		"payload" : {
			"fulfilmentRequest" : {
				"fulfilmentCode": "P_UAC_UACIP1",
				"caseId" : "bbd55984-0dbf-4499-bfa7-0aa4228700e9",
				"individualCaseId" : "8e8ebf71-d9c6-4efa-a693-ae24e7116e98",
				"contact": {}
			}
		}
	}`,
		`{"event":{"type":"FULFILMENT_REQUESTED","source":"QUESTIONNAIRE_RUNNER","channel":"EQ","dateTime":"2011-08-12T20:17:46.384Z","transactionId":"c45de4dc-3c3b-11e9-b210-d663bd873d93"},"payload":{"fulfilmentRequest":{"fulfilmentCode":"P_UAC_UACIP1","caseId":"bbd55984-0dbf-4499-bfa7-0aa4228700e9","individualCaseId":"8e8ebf71-d9c6-4efa-a693-ae24e7116e98","contact":{}}}}`,
		cfg.EqFulfilmentTopic, cfg.EqFulfilmentProject, cfg.FulfilmentRequestRoutingKey))

}

func testMessageProcessing(messageToSend string, expectedRabbitMessage string, topic string, project string, rabbitRoutingKey string) func(t *testing.T) {
	return func(t *testing.T) {
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := StartProcessors(timeout, cfg, make(chan processor.Error)); err != nil {
			assert.NoError(t, err)
			return
		}

		rabbitConn, rabbitCh, err := connectToRabbitChannel()
		assert.NoError(t, err)
		defer rabbitCh.Close()
		defer rabbitConn.Close()

		if _, err := rabbitCh.QueuePurge(rabbitRoutingKey, false); err != nil {
			assert.NoError(t, err)
			return
		}

		// When
		if messageId, err := publishMessageToPubSub(timeout, messageToSend, topic, project); err != nil {
			t.Errorf("PubSub publish fail, project: %s, topic: %s, id: %s, error: %s", project, topic, messageId, err)
			return
		}

		rabbitMessage, err := getFirstMessageOnQueue(timeout, rabbitRoutingKey, rabbitCh)
		if !assert.NoErrorf(t, err, "Did not find message on queue %s", rabbitRoutingKey) {
			return
		}

		// Then
		assert.Equal(t, expectedRabbitMessage, rabbitMessage)
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

	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	mockResult := "Success!"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(mockResult))
		requests = append(requests, r)
		requestBody, _ = ioutil.ReadAll(r.Body)
	}))
	defer srv.Close()

	cfg.QuarantineMessageUrl = srv.URL

	if _, err := StartProcessors(timeout, cfg, make(chan processor.Error)); err != nil {
		assert.NoError(t, err)
		return
	}

	// When
	if _, err := publishMessageToPubSub(timeout, messageToSend, cfg.EqReceiptTopic, cfg.EqReceiptProject); err != nil {
		t.Errorf("Failed to publish message to PubSub, err: %s", err)
		return
	}

	// Allow a second for the processor to process the message
	time.Sleep(1 * time.Second)

	if !assert.Len(t, requests, 1, "Unexpected number of calls to Exception Manager") {
		return
	}

	var quarantineBody models.MessageToQuarantine
	err := json.Unmarshal(requestBody, &quarantineBody)
	if err != nil {
		t.Errorf("Could not decode request body sent to Exception Manager")
		return
	}

	assert.Equal(t, "application/json", quarantineBody.ContentType, "Dodgy content type")
	assert.Equal(t, "Error unmarshalling message", quarantineBody.ExceptionClass, "Dodgy exception class")
	assert.Equal(t, messageToSend, string(quarantineBody.MessagePayload), "Dodgy message payload")
	assert.Equal(t, cfg.EqReceiptSubscription, quarantineBody.Queue, "Dodgy quarantine queue")
	assert.Equal(t, "none", quarantineBody.RoutingKey, "Dodgy routing key")
	assert.Equal(t, "Pubsub Adapter", quarantineBody.Service, "Dodgy routing key")
	assert.Len(t, quarantineBody.MessageHash, 64, "Dodgy message hash")
	assert.Len(t, quarantineBody.Headers, 1, "Dodgy headers")
	assert.Contains(t, quarantineBody.Headers, "pubSubId", "Dodgy headers, missing pubSubId")
	assert.False(t, len(quarantineBody.Headers["pubSubId"]) == 0, "Dodgy pubSubId header, expected non-zero length")
}

func TestStartProcessors(t *testing.T) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	processors, err := StartProcessors(timeout, cfg, make(chan processor.Error))
	if err != nil {
		assert.NoError(t, err)
		return
	}

	assert.Len(t, processors, 6, "StartProcessors should return 6 processors")
}

func TestRabbitReconnectOnChannelDeath(t *testing.T) {
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errChan := make(chan processor.Error)

	// Initialise up the processors normally
	processors, err := StartProcessors(timeout, cfg, errChan)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	ready := readiness.New(timeout, cfg.ReadinessFilePath)
	assert.NoError(t, ready.Ready())

	// Start the run loop
	go RunLoop(timeout, cfg, nil, errChan, ready)

	// Take the first testProcessor
	testProcessor := processors[0]

	// Pick one of the processors rabbit channels
	var channel processor.RabbitChannel
	for channel == nil {
		select {
		case <-timeout.Done():
			t.Error()
			return
		default:
			if len(testProcessor.RabbitChannels) > 0 {
				channel = testProcessor.RabbitChannels[0]
			}
		}
	}

	// Subscribe another test channel to the close notifications
	channelErrChan := make(chan *amqp.Error)
	channel.NotifyClose(channelErrChan)

	// Check the processors rabbit channel can publish
	if err := publishToRabbit(channel, cfg.EventsExchange, cfg.ReceiptRoutingKey, `{"test":"message should publish before"}`); err != nil {
		assert.NoError(t, err)
		return
	}

	// Break the channel by publishing a mandatory message that can't be routed
	// NB: This is not a typical scenario this feature is designed around as the app or rabbit would have to be
	// mis-configured for this to occur, and the channel closing is only an undesirable side effect.
	// It is, however, the only viable way of inducing a channel close that I could think of using to exercise this code.
	if err := publishToRabbit(channel, "this_exchange_should_not_exist", cfg.ReceiptRoutingKey, `{"test":"message should fail"}`); err != nil {
		assert.NoError(t, err)
		return
	}

	// Wait for the unpublishable message to kill the channel with a timeout
	select {
	case <-timeout.Done():
		assert.Fail(t, "Timed out waiting for induced rabbit channel closure")
		return
	case <-channelErrChan:
	}

	// Try to successfully publish a message within the timeout
	success := make(chan bool)
	go attemptPublishOnProcessorsChannel(timeout, testProcessor, success)

	select {
	case <-timeout.Done():
		t.Error("Failed to publish message with processors channel within the timeout")
		return
	case <-success:
		return
	}
}

func TestRabbitReconnectOnBadConnection(t *testing.T) {
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Induce a connection failure by using the wrong connection string
	brokenCfg := *cfg
	brokenCfg.RabbitConnectionString = "bad-connection-string"

	errChan := make(chan processor.Error)

	// Initialise up the processors normally
	processors, err := StartProcessors(timeout, &brokenCfg, errChan)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	ready := readiness.New(timeout, cfg.ReadinessFilePath)
	assert.NoError(t, ready.Ready())

	// Start the run loop
	go RunLoop(timeout, cfg, nil, errChan, ready)

	// Take the first processor
	testProcessor := processors[0]

	// Give it a second to attempt connection and fail
	time.Sleep(1 * time.Second)

	// Fix the config
	testProcessor.Config.RabbitConnectionString = cfg.RabbitConnectionString

	// Try to successfully publish a message using the processors channel within the timeout
	success := make(chan bool)
	go attemptPublishOnProcessorsChannel(timeout, testProcessor, success)

	select {
	case <-timeout.Done():
		t.Error("Failed to publish message with processors channel within the timeout")
		return
	case <-success:
		return
	}
}

func TestUnsuccessfulRestartsTimeOut(t *testing.T) {
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Induce a connection failure by using the wrong connection string
	brokenCfg := *cfg
	brokenCfg.RabbitConnectionString = "bad-connection-string"

	// Turn down the restart timeout to 1 second for the test
	brokenCfg.RestartTimeout = 1

	errChan := make(chan processor.Error)

	// Initialise up the processors normally
	_, err := StartProcessors(timeout, &brokenCfg, errChan)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	ready := readiness.New(timeout, cfg.ReadinessFilePath)
	assert.NoError(t, ready.Ready())

	// Start the run loop in a goroutine, send success when it exits
	success := make(chan bool)
	go func() {
		RunLoop(timeout, &brokenCfg, nil, errChan, ready)
		success <- true
	}()

	select {
	case <-success:
		return
	case <-timeout.Done():
		assert.Fail(t, "Test timed out before the run loop exited on the timeout")
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

func publishToRabbit(channel processor.RabbitChannel, exchange string, routingKey string, message string) error {
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

func attemptPublishOnProcessorsChannel(ctx context.Context, testProcessor *processor.Processor, success chan bool) {
	for {
		select {
		case <-ctx.Done():
			// Kill this goroutine if the test times out
			return
		default:
			// Repeatedly try to publish a message using the processors channel
			if len(testProcessor.RabbitChannels) == 0 {
				continue
			}
			channel := testProcessor.RabbitChannels[0]
			if err := publishToRabbit(channel, cfg.EventsExchange, cfg.ReceiptRoutingKey, `{"test":"message should publish after"}`); err == nil {
				// We have successfully published a message with the processors re-opened rabbit channel
				success <- true
				return
			}
		}
	}
}
