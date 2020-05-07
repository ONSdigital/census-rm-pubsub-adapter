package processor

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"reflect"
	"testing"
	"time"
)

func TestConvertPpoUndeliveredToRmMessage(t *testing.T) {
	timeCreated, _ := time.Parse("2006-07-08T03:04:05Z", "2008-08-24T00:00:00Z")
	hazyTimeCreated := models.HazyUtcTime{Time: timeCreated}
	ppoUndeliveredMessage := models.PpoUndelivered{
		DateTime:      &hazyTimeCreated,
		TransactionId: "abc123xxx",
		CaseRef:       "123456789",
		ProductCode:   "P_TEST_1",
	}

	expectedRabbitMessage := models.RmMessage{
		Event: models.RmEvent{
			Type:          "UNDELIVERED_MAIL_REPORTED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "PPO",
			DateTime:      &timeCreated,
			TransactionID: "abc123xxx",
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &models.FulfilmentInformation{
				CaseRef:        "123456789",
				FulfilmentCode: "P_TEST_1",
			},
		}}

	rabbitMessage, err := convertPpoUndeliveredToRmMessage(ppoUndeliveredMessage)
	if err != nil {
		t.Errorf("failed: %s", err)
	}

	if !reflect.DeepEqual(expectedRabbitMessage, *rabbitMessage) {
		t.Errorf("Incorrect Rabbit message structure \nexpected:%+v \nactual:%+v", expectedRabbitMessage, rabbitMessage)
	}
}
