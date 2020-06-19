package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
)

func NewEqFulfilmentProcessor(ctx context.Context, appConfig *config.Configuration, errChan chan error) (*Processor, error) {
	return NewProcessor(ctx, appConfig, appConfig.EqFulfilmentProject, appConfig.EqFulfilmentSubscription, appConfig.FulfilmentRequestRoutingKey, convertEqFulfilment, unmarshalEqFulfilment, errChan)
}

func unmarshalEqFulfilment(bytes []byte) (models.InboundMessage, error) {
	var eqFulfilment models.EqFulfilment
	if err := json.Unmarshal(bytes, &eqFulfilment); err != nil {
		return nil, err
	}
	if err := eqFulfilment.Validate(); err != nil {
		return nil, err
	}
	return eqFulfilment, nil
}

func convertEqFulfilment(fulfilmentRequest models.InboundMessage) (*models.RmMessage, error) {
	eqFulfilment, ok := fulfilmentRequest.(models.EqFulfilment)
	if !ok {
		return nil, fmt.Errorf("wrong message model given to convertEqFulfilment: %T, only accepts EqFulfilment, tx_id: %q", fulfilmentRequest, fulfilmentRequest.GetTransactionId())
	}
	return &models.RmMessage{
		Event: *eqFulfilment.EqFulfilmentEvent,
		Payload: models.RmPayload{
			FulfilmentRequest: eqFulfilment.EqFulfilmentPayload.FulfilmentRequest,
		},
	}, nil
}
