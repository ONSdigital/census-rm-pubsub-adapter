package models

import (
	"github.com/ONSdigital/census-rm-pubsub-adapter/validate"
)

type PpoUndelivered struct {
	TransactionId string       `json:"transactionId" validate:"required"`
	DateTime      *HazyUtcTime `json:"dateTime" validate:"required"`
	CaseRef       string       `json:"caseRef" validate:"required"`
	ProductCode   string       `json:"productCode"`
}

func (p PpoUndelivered) GetTransactionId() string {
	return p.TransactionId
}

func (p PpoUndelivered) Validate() error {
	return validate.Validate.Struct(p)
}
