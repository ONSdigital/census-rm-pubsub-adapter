package models

import "time"

type QmUndelivered struct {
	TransactionId   string    `json:"transactionId"`
	DateTime        time.Time `json:"dateTime"`
	QuestionnaireId string    `json:"questionnaireId"`
}

func (q QmUndelivered) GetTransactionId() string {
	return q.TransactionId
}
