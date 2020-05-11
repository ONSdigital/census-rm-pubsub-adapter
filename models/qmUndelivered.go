package models

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/validate"
)

type QmUndelivered struct {
	TransactionId   string       `json:"transactionId" validate:"required"`
	DateTime        *HazyUtcTime `json:"dateTime" validate:"required"`
	QuestionnaireId string       `json:"questionnaireId" validate:"required"`
}

func (q QmUndelivered) GetTransactionId() string {
	return q.TransactionId
}

func (q QmUndelivered) Validate() error {
	return validate.Validate.Struct(q)
}
