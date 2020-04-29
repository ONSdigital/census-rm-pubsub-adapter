package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
)

func NewPpoUndeliveredProcessor(ctx context.Context, appConfig *config.Configuration) *Processor {
	return NewProcessor(ctx, appConfig, appConfig.PpoUndeliveredProject, appConfig.PpoUndeliveredSubscription, appConfig.UndeliveredRoutingKey, convertPpoUndeliveredToRmMessage, unmarshalPpoUndelivered)
}

func unmarshalPpoUndelivered(data []byte) (models.PubSubMessage, error) {
	var ppoUndelivered models.PpoUndelivered
	if err := json.Unmarshal(data, &ppoUndelivered); err != nil {
		return nil, err
	}
	return ppoUndelivered, nil
}

func convertPpoUndeliveredToRmMessage(message models.PubSubMessage) (*models.RmMessage, error) {
	ppoUndelivered, ok := message.(models.PpoUndelivered)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Wrong message model given to convertPpoUndeliveredToRmMessage: %T, only accepts ppoUndelivered, tx_id: %q", message, message.GetTransactionId()))
	}

	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "UNDELIVERED_MAIL_REPORTED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "PPO",
			DateTime:      ppoUndelivered.DateTime.Time,
			TransactionID: ppoUndelivered.TransactionId,
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &models.FulfilmentInformation{
				CaseRef:        ppoUndelivered.CaseRef,
				FulfilmentCode: ppoUndelivered.ProductCode,
			},
		},
	}, nil
}
