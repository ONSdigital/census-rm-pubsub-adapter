package eqreceipt

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/models/incoming-pubsub"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models/rabbit"
	"reflect"
	"testing"
)

func TestConvertEqReceiptToRmMessage(t *testing.T) {
	eqReceiptMessage := incoming_pubsub.EqReceipt{"2008-08-24T00:00:00Z",
		incoming_pubsub.EqReceiptMetadata{
			TransactionID: "abc123xxx", QuestionnaireID: "01213213213"}}

	expectedRabbitMessage := rabbit.RmMessage{
		Event: rabbit.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "EQ",
			DateTime:      "2008-08-24T00:00:00Z",
			TransactionID: "abc123xxx",
		},
		Payload: rabbit.RmPayload{
			Response: rabbit.RmResponse{
				QuestionnaireID: "01213213213",
			},
		}}

	rabbitMessage, err := convertEqReceiptToRmMessage(&eqReceiptMessage)
	if err != nil {
		t.Errorf("failed: %s", err)
	}

	if !reflect.DeepEqual(expectedRabbitMessage, *rabbitMessage) {
		t.Errorf("Incorrect Rabbit message structure \nexpected:%+v \nactual:%+v", expectedRabbitMessage, rabbitMessage)
	}
}
