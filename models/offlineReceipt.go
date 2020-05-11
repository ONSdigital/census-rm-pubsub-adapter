package models

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/validate"
)

type OfflineReceipt struct {
	TimeCreated     *HazyUtcTime `json:"dateTime" validate:"required"`
	TransactionId   string       `json:"transactionId" validate:"required"`
	QuestionnaireId string       `json:"questionnaireId" validate:"required"`
	Unreceipt       bool         `json:"unreceipt"`
	Channel         string       `json:"channel"`
}

func (o OfflineReceipt) GetTransactionId() string {
	return o.TransactionId
}

func (o OfflineReceipt) Validate() error {
	return validate.Validate.Struct(o)
}
