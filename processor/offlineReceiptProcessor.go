package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
)

func NewOfflineReceiptProcessor(ctx context.Context, appConfig *config.Configuration, errChan chan error) (*Processor, error) {
	return NewProcessor(ctx, appConfig, appConfig.OfflineReceiptProject, appConfig.OfflineReceiptSubscription, appConfig.ReceiptRoutingKey, convertOfflineReceiptToRmMessage, unmarshalOfflineReceipt, errChan)
}

func unmarshalOfflineReceipt(data []byte) (models.InboundMessage, error) {
	var offlineReceipt models.OfflineReceipt
	if err := json.Unmarshal(data, &offlineReceipt); err != nil {
		return nil, err
	}
	if err := offlineReceipt.Validate(); err != nil {
		return nil, err
	}
	return offlineReceipt, nil
}

func convertOfflineReceiptToRmMessage(receipt models.InboundMessage) (*models.RmMessage, error) {
	offlineReceipt, ok := receipt.(models.OfflineReceipt)
	if !ok {
		return nil, fmt.Errorf("wrong message model given to convertEqReceiptToRmMessage: %T, only accepts EqReceipt, tx_id: %q", receipt, receipt.GetTransactionId())
	}
	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       offlineReceipt.Channel,
			DateTime:      &offlineReceipt.TimeCreated.Time,
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
