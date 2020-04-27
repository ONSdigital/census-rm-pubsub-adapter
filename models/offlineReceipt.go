package models

import "time"

type OfflineReceipt struct {
	TimeCreated     time.Time `json:"dateTime"`
	TransactionId   string    `json:"tx_id"`
	QuestionnaireId string    `json:"questionnaire_id"`
	Unreceipt       bool      `json:"unreceipt"`
	Channel         string    `json:"channel"`
}

func (o OfflineReceipt) GetTransactionId() string {
	return o.TransactionId
}
