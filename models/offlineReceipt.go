package models

import "time"

type OfflineReceipt struct {
	TimeCreated     time.Time `json:"dateTime"`
	TransactionID   string    `json:"tx_id"`
	QuestionnaireID string    `json:"questionnaire_id"`
	Unreceipt       bool      `json:"unreceipt"`
	Channel         string    `json:"channel"`
}

func (o OfflineReceipt) GetQuestionnaireId() string {
	return o.QuestionnaireID
}

func (o OfflineReceipt) GetTransactionID() string {
	return o.TransactionID
}
