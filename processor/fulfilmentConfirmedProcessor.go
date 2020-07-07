package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
)

func NewFulfilmentConfirmedProcessor(ctx context.Context, appConfig *config.Configuration, errChan chan Error) (*Processor, error) {
	return NewProcessor(ctx, appConfig, appConfig.FulfilmentConfirmedProject, appConfig.FulfilmentConfirmedSubscription, appConfig.FulfilmentConfirmationRoutingKey, convertFulfilmentConfirmedToRmMessage, unmarshalFulfilmentConfirmed, errChan)
}

func unmarshalFulfilmentConfirmed(data []byte) (models.InboundMessage, error) {
	var fulfilmentConfirmed models.FulfilmentConfirmed
	if err := json.Unmarshal(data, &fulfilmentConfirmed); err != nil {
		return nil, err
	}
	if err := fulfilmentConfirmed.Validate(); err != nil {
		return nil, err
	}
	return fulfilmentConfirmed, nil
}

func convertFulfilmentConfirmedToRmMessage(confirmation models.InboundMessage) (*models.RmMessage, error) {
	fulfilmentConfirmed, ok := confirmation.(models.FulfilmentConfirmed)
	if !ok {
		return nil, fmt.Errorf("wrong message model given to convertFulfilmentConfirmedToRmMessage: %T, only accepts FulfilmentConfirmed, tx_id: %q", confirmation, confirmation.GetTransactionId())
	}

	var payload models.FulfilmentInformation
	if fulfilmentConfirmed.Channel == "QM" {
		payload = models.FulfilmentInformation{
			QuestionnaireId: fulfilmentConfirmed.QuestionnaireId,
			FulfilmentCode:  fulfilmentConfirmed.ProductCode,
		}
	} else {
		// This must be a PPO message, by a process of elimination
		payload = models.FulfilmentInformation{
			CaseRef:        fulfilmentConfirmed.CaseRef,
			FulfilmentCode: fulfilmentConfirmed.ProductCode,
		}
	}

	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "FULFILMENT_CONFIRMED",
			Source:        "RECEIPT_SERVICE",
			Channel:       fulfilmentConfirmed.Channel,
			DateTime:      &fulfilmentConfirmed.TimeCreated.Time,
			TransactionID: fulfilmentConfirmed.TransactionId,
		},
		Payload: models.RmPayload{
			FulfilmentInformation: &payload,
		},
	}, nil
}
