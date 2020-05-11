package processor

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"reflect"
	"testing"
	"time"
)

func TestConvertQmUndeliveredToRmMessage(t *testing.T) {
	timeCreated, _ := time.Parse("2006-07-08T03:04:05Z", "2008-08-24T00:00:00Z")
	hazyTimeCreated := models.HazyUtcTime{Time: timeCreated}
	qmUndeliveredMessage := models.QmUndelivered{
		DateTime:        &hazyTimeCreated,
		TransactionId:   "abc123xxx",
		QuestionnaireId: "01213213213",
	}

	expectedRabbitMessage := models.RmMessage{
		Event: models.RmEvent{
			Type:          "UNDELIVERED_MAIL_REPORTED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "QM",
			DateTime:      &timeCreated,
			TransactionID: "abc123xxx",
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &models.FulfilmentInformation{
				QuestionnaireId: "01213213213",
			},
		}}

	rabbitMessage, err := convertQmUndeliveredToRmMessage(qmUndeliveredMessage)
	if err != nil {
		t.Errorf("failed: %s", err)
	}

	if !reflect.DeepEqual(expectedRabbitMessage, *rabbitMessage) {
		t.Errorf("Incorrect Rabbit message structure \nexpected:%+v \nactual:%+v", expectedRabbitMessage, rabbitMessage)
	}
}
