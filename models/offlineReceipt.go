package models

import "time"

type OfflineReceipt struct {
	TimeCreated     time.Time `json:"dateTime"`
	TransactionId   string    `json:"transactionId"`
	QuestionnaireId string    `json:"questionnaireId"`
	Unreceipt       bool      `json:"unreceipt"`
	Channel         string    `json:"channel"`
}

func (o OfflineReceipt) GetTransactionId() string {
	return o.TransactionId
}
