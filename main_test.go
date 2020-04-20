package main

import (
	"cloud.google.com/go/pubsub"
	"github.com/ONSdigital/census-rm-pubsub-adapter/processor"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
	"os"
	"testing"
	"time"
)

var a processor.App

func TestMain(m *testing.M) {
	ctx := context.Background()
	a.Setup(ctx, "amqp://guest:guest@localhost:6672/", "project")
	code := m.Run()
	os.Exit(code)
}

func TestEqReceipt(t *testing.T) {

	eqRecieptMsg := `{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`
	expectedRabbitMessage := `{"event":{"type":"RESPONSE_RECEIVED","source":"RECEIPT_SERVICE","channel":"EQ","dateTime":"2008-08-24T00:00:00Z","transactionId":"abc123xxx"},"payload":{"response":{"caseId":null,"questionnaireId":"01213213213","unreceipt":false}}}`

	rabbitConn, err := amqp.Dial("amqp://guest:guest@localhost:6672/")
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
	pubSubClient, err := pubsub.NewClient(ctx, "project")
	eqReceiptTopic := pubSubClient.Topic("eq-submission-topic")

	result := eqReceiptTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(eqRecieptMsg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		t.Errorf("Pubsub test message %s publish failed. %s", id, err)
	}

	go a.Consume(ctx)
	go a.Process(ctx)

	msgs, err := rabbitChan.Consume("goTestQueue", "", false, false, false, false, nil)
	if err != nil {
		t.Errorf("Rabbit consume failed: %s", err)
	}

	var rabbitMessages []string
	select {
	case d := <-msgs:
		rabbitMessages = append(rabbitMessages, string(d.Body))
	}

	if string(rabbitMessages[0]) != expectedRabbitMessage {
		t.Errorf("Rabbit messsage incorrect - \nexpected: %s \n actual: %s", expectedRabbitMessage, rabbitMessages[0])
	}
	cancel()

}
