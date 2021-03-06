package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
)

func NewQmUndeliveredProcessor(ctx context.Context, appConfig *config.Configuration, errChan chan Error) (*Processor, error) {
	return NewProcessor(ctx, appConfig, appConfig.QmUndeliveredProject, appConfig.QmUndeliveredSubscription, appConfig.UndeliveredRoutingKey, convertQmUndeliveredToRmMessage, unmarshalQmUndelivered, errChan)
}

func unmarshalQmUndelivered(data []byte) (models.InboundMessage, error) {
	var qmUndelivered models.QmUndelivered
	if err := json.Unmarshal(data, &qmUndelivered); err != nil {
		return nil, err
	}
	if err := qmUndelivered.Validate(); err != nil {
		return nil, err
	}
	return qmUndelivered, nil
}

func convertQmUndeliveredToRmMessage(message models.InboundMessage) (*models.RmMessage, error) {
	qmUndelivered, ok := message.(models.QmUndelivered)
	if !ok {
		return nil, fmt.Errorf("wrong message model given to convertQmUndeliveredToRmMessage: %T, only accepts qmUndelivered, tx_id: %q", message, message.GetTransactionId())
	}

	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "UNDELIVERED_MAIL_REPORTED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "QM",
			DateTime:      &qmUndelivered.DateTime.Time,
			TransactionID: qmUndelivered.TransactionId,
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &models.FulfilmentInformation{
				QuestionnaireId: qmUndelivered.QuestionnaireId,
			},
		},
	}, nil
}
