package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
)

type qmUndeliveredProcessor struct {
	*Processor
}

func NewQmUndeliveredProcessor(ctx context.Context, appConfig *config.Configuration) *qmUndeliveredProcessor {
	qmUndeliveredProcessor := &qmUndeliveredProcessor{}
	qmUndeliveredProcessor.Processor = NewProcessor(ctx, appConfig, appConfig.QmUndeliveredProject, appConfig.QmUndeliveredSubscription, convertQmUndeliveredToRmMessage, unmarshalQmUndelivered)
	return qmUndeliveredProcessor
}

func unmarshalQmUndelivered(data []byte) (models.PubSubMessage, error) {
	var qmUndelivered models.QmUndelivered
	err := json.Unmarshal(data, &qmUndelivered)
	if err != nil {
		return nil, err
	}
	return qmUndelivered, nil
}

func convertQmUndeliveredToRmMessage(message models.PubSubMessage) (*models.RmMessage, error) {
	qmUndelivered, ok := message.(models.QmUndelivered)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Wrong message model given to convertQmUndeliveredToRmMessage: %T, only accepts qmUndelivered, tx_id: %q", message, message.GetTransactionId()))
	}

	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "UNDELIVERED_MAIL_REPORTED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "QM",
			DateTime:      qmUndelivered.DateTime,
			TransactionID: qmUndelivered.TransactionId,
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &models.FulfilmentInformation{
				QuestionnaireId: qmUndelivered.QuestionnaireId,
			},
		},
	}, nil
}
