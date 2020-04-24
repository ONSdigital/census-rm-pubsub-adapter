package processor

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/models"
	"log"

	"github.com/pkg/errors"
)

type EqReceiptProcessor struct {
	*Processor
}

func NewEqReceiptProcessor(ctx context.Context, appConfig *config.Configuration) *EqReceiptProcessor {
	eqReceiptProcessor := &EqReceiptProcessor{}
	eqReceiptProcessor.Processor = NewProcessor(ctx, appConfig, appConfig.EqReceiptProject, appConfig.EqReceiptSubscription)
	return eqReceiptProcessor
}

func (a *EqReceiptProcessor) Process(ctx context.Context) {
	for {
		select {
		case msg := <-a.MessageChan:
			var eqReceiptReceived models.EqReceipt
			err := json.Unmarshal(msg.Data, &eqReceiptReceived)
			if err != nil {
				// TODO Log the error and DLQ the message when unmarshalling fails, printing it out is a temporary solution
				log.Println(errors.WithMessagef(err, "Error unmarshalling message: %q", string(msg.Data)))
				msg.Ack()
				return
			}

			log.Printf("Got QID: %q\n", eqReceiptReceived.Metadata.QuestionnaireID)
			rmMessageToSend, err := convertEqReceiptToRmMessage(&eqReceiptReceived)
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

func convertEqReceiptToRmMessage(eqReceipt *models.EqReceipt) (*models.RmMessage, error) {
	if eqReceipt == nil {
		return nil, errors.New("receipt has nil content")
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
