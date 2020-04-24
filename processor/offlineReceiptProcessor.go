package processor

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"github.com/pkg/errors"
	"log"
)

type OfflineReceiptProcessor struct {
	*Processor
}

func NewOfflineReceiptProcessor(ctx context.Context, appConfig *config.Configuration) *OfflineReceiptProcessor {
	offlineReceiptProcessor := &OfflineReceiptProcessor{}
	offlineReceiptProcessor.Processor = NewProcessor(ctx, appConfig, appConfig.OfflineReceiptProject, appConfig.OfflineReceiptSubscription)
	return offlineReceiptProcessor
}

func (a *OfflineReceiptProcessor) Process(ctx context.Context) {
	for {
		select {
		case msg := <-a.MessageChan:
			var offlineReceiptReceived models.OfflineReceipt
			err := json.Unmarshal(msg.Data, &offlineReceiptReceived)
			if err != nil {
				// TODO Log the error and DLQ the message when unmarshalling fails, printing it out is a temporary solution
				log.Println(errors.WithMessagef(err, "Error unmarshalling message: %q", string(msg.Data)))
				msg.Ack()
				return
			}

			log.Printf("Got QID: %q\n", offlineReceiptReceived.QuestionnaireID)
			rmMessageToSend, err := convertOfflineReceiptToRmMessage(&offlineReceiptReceived)
			if err != nil {
				log.Println(errors.Wrap(err, "failed to convert receipt to message"))
			}
			err = a.publishEventToRabbit(rmMessageToSend, a.Config.ReceiptRoutingKey, a.Config.EventsExchange)
			if err != nil {
				log.Println(errors.WithMessagef(err, "Failed to publish eq receipt message tx_id: %s", rmMessageToSend.Event.TransactionID))
				msg.Nack()
			} else {
				msg.Ack()
			}
		case <-ctx.Done():
			//stop the loop from consuming messages
			return
		}
	}

}

func convertOfflineReceiptToRmMessage(offlineReceipt *models.OfflineReceipt) (*models.RmMessage, error) {
	if offlineReceipt == nil {
		return nil, errors.New("receipt has nil content")
	}

	return &models.RmMessage{
		Event: models.RmEvent{
			Type:          "RESPONSE_RECEIVED",
			Source:        "RECEIPT_SERVICE",
			Channel:       offlineReceipt.Channel,
			DateTime:      offlineReceipt.TimeCreated,
			TransactionID: offlineReceipt.TransactionID,
		},
		Payload: models.RmPayload{
			Response: models.RmResponse{
				QuestionnaireID: offlineReceipt.QuestionnaireID,
				Unreceipt:       offlineReceipt.Unreceipt,
			},
		},
	}, nil
}
