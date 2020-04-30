package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/pkg/errors"
)

func NewOfflineReceiptProcessor(ctx context.Context, appConfig *config.Configuration) *Processor {
	return NewProcessor(ctx, appConfig, appConfig.OfflineReceiptProject, appConfig.OfflineReceiptSubscription, appConfig.ReceiptRoutingKey, convertOfflineReceiptToRmMessage, unmarshalOfflineReceipt)
}

func unmarshalOfflineReceipt(data []byte) (models.PubSubMessage, error) {
	var offlineReceipt models.OfflineReceipt
	err := json.Unmarshal(data, &offlineReceipt)
	if err != nil {
		return nil, err
	}
	return offlineReceipt, nil
}

func convertOfflineReceiptToRmMessage(receipt models.PubSubMessage) (*models.RmMessage, error) {
	offlineReceipt, ok := receipt.(models.OfflineReceipt)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Wrong message model given to convertEqReceiptToRmMessage: %T, only accepts EqReceipt, tx_id: %q", receipt, receipt.GetTransactionId()))
	}
	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       offlineReceipt.Channel,
			DateTime:      offlineReceipt.TimeCreated.Time,
			TransactionID: offlineReceipt.TransactionId,
		},
		Payload: models.RmPayload{
			Response: &models.RmResponse{
				QuestionnaireID: offlineReceipt.QuestionnaireId,
				Unreceipt:       offlineReceipt.Unreceipt,
			},
		},
	}, nil
}
