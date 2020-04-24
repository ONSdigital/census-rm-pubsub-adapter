package processor

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"reflect"
	"testing"
	"time"
)

func TestConvertEqReceiptToRmMessage(t *testing.T) {
	timeCreated, _ := time.Parse("2006-07-08T03:04:05Z", "2008-08-24T00:00:00Z")
	eqReceiptMessage := models.EqReceipt{timeCreated,
		models.EqReceiptMetadata{
			TransactionID: "abc123xxx", QuestionnaireID: "01213213213"}}

	expectedRabbitMessage := models.RmMessage{
		Event: models.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "EQ",
			DateTime:      timeCreated,
			TransactionID: "abc123xxx",
		},
		Payload: models.RmPayload{
			Response: models.RmResponse{
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
