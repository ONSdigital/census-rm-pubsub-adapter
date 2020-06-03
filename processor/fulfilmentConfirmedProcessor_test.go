package processor

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestConvertQMFulfilmentConfirmedToRmMessage(t *testing.T) {
	timeCreated, _ := time.Parse("2006-07-08T03:04:05Z", "2008-08-24T00:00:00Z")
	hazyTimeCreated := models.HazyUtcTime{Time: timeCreated}
	offlineReceiptMessage := models.FulfilmentConfirmed{
		TimeCreated:     &hazyTimeCreated,
		TransactionId:   "abc123xxx",
		QuestionnaireId: "01213213213",
		Channel:         "QM",
		ProductCode:     "ABC_XYZ_123",
	}

	expectedRabbitMessage := models.RmMessage{
		Event: models.RmEvent{
			Type:          "FULFILMENT_CONFIRMED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "QM",
			DateTime:      &timeCreated,
			TransactionID: "abc123xxx",
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &models.FulfilmentInformation{
				QuestionnaireId: "01213213213",
				FulfilmentCode:  "ABC_XYZ_123",
			},
		}}

	rabbitMessage, err := convertFulfilmentConfirmedToRmMessage(offlineReceiptMessage)
	if err != nil {
		t.Errorf("failed: %s", err)
	}

	assert.Equal(t, expectedRabbitMessage, *rabbitMessage, "Incorrect Rabbit message structure")
}

func TestConvertPPOFulfilmentConfirmedToRmMessage(t *testing.T) {
	timeCreated, _ := time.Parse("2006-07-08T03:04:05Z", "2008-08-24T00:00:00Z")
	hazyTimeCreated := models.HazyUtcTime{Time: timeCreated}
	offlineReceiptMessage := models.FulfilmentConfirmed{
		TimeCreated:   &hazyTimeCreated,
		TransactionId: "abc123xxx",
		CaseRef:       "666123999",
		Channel:       "PPO",
		ProductCode:   "ABC_XYZ_123",
	}

	expectedRabbitMessage := models.RmMessage{
		Event: models.RmEvent{
			Type:          "FULFILMENT_CONFIRMED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "PPO",
			DateTime:      &timeCreated,
			TransactionID: "abc123xxx",
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &models.FulfilmentInformation{
				CaseRef:        "666123999",
				FulfilmentCode: "ABC_XYZ_123",
			},
		}}

	rabbitMessage, err := convertFulfilmentConfirmedToRmMessage(offlineReceiptMessage)
	if err != nil {
		t.Errorf("failed: %s", err)
	}

	if !reflect.DeepEqual(expectedRabbitMessage, *rabbitMessage) {
		t.Errorf("Incorrect Rabbit message structure \nexpected:%+v \nactual:%+v", expectedRabbitMessage, rabbitMessage)
	}
}
