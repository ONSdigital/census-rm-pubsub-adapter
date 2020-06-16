package processor

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConvertEqFulfilmentToRmMessage(t *testing.T) {
	timeCreated, _ := time.Parse("2006-07-08T03:04:05Z", "2008-08-24T00:00:00Z")
	eqFulfilment := models.EqFulfilment{
		EqFulfilmentEvent: &models.RmEvent{
			Type:          "FULFILMENT_REQUESTED",
			Source:        "QUESTIONNAIRE_RUNNER",
			Channel:       "EQ",
			DateTime:      &timeCreated,
			TransactionID: "abc123xxx",
		},
		EqFulfilmentPayload: &models.EqFulfilmentPayload{
			FulfilmentRequest: &models.FulfilmentRequest{
				FulfilmentCode: "ABC_XYZ_123",
				CaseID:         "test",
				Contact: &models.Contact{
					TelNo: "123",
				},
			},
		},
	}

	expectedRabbitMessage := models.RmMessage{
		Event: *eqFulfilment.EqFulfilmentEvent,
		Payload: models.RmPayload{
			FulfilmentRequest: eqFulfilment.EqFulfilmentPayload.FulfilmentRequest,
		},
	}

	rabbitMessage, err := convertEqFulfilment(eqFulfilment)
	if err != nil {
		t.Errorf("failed: %s", err)
	}

	assert.Equal(t, expectedRabbitMessage, *rabbitMessage, "Incorrect Rabbit message structure")
}
