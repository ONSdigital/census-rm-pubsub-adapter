package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/pkg/errors"
)

type EqReceiptProcessor struct {
	*Processor
}

func NewEqReceiptProcessor(ctx context.Context, appConfig *config.Configuration) *EqReceiptProcessor {
	eqReceiptProcessor := &EqReceiptProcessor{}
	eqReceiptProcessor.Processor = NewProcessor(ctx, appConfig, appConfig.EqReceiptProject, appConfig.EqReceiptSubscription, convertEqReceiptToRmMessage, unmarshalEqReceipt)
	return eqReceiptProcessor
}

func unmarshalEqReceipt(data []byte) (models.PubSubMessage, error) {
	var eqReceipt models.EqReceipt
	err := json.Unmarshal(data, &eqReceipt)
	if err != nil {
		return nil, err
	}
	return eqReceipt, nil
}

func convertEqReceiptToRmMessage(receipt models.PubSubMessage) (*models.RmMessage, error) {
	eqReceipt, ok := receipt.(models.EqReceipt)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Wrong message model given to convertEqReceiptToRmMessage: %T, only accepts EqReceipt, tx_id %q", receipt, receipt.GetQuestionnaireId()))
	}

	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       "EQ",
			DateTime:      eqReceipt.TimeCreated,
			TransactionID: eqReceipt.Metadata.TransactionID,
		},
		Payload: models.RmPayload{
			Response: models.RmResponse{
				QuestionnaireID: eqReceipt.Metadata.QuestionnaireID,
				CaseID:          eqReceipt.Metadata.CaseID,
			},
		},
	}, nil
}
