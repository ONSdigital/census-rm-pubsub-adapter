package models

import "time"

type PpoUndelivered struct {
	TransactionId string    `json:"transactionId"`
	DateTime      time.Time `json:"dateTime"`
	CaseRef       string    `json:"caseRef"`
	ProductCode   string    `json:"productCode"`
}

func (p PpoUndelivered) GetTransactionId() string {
	return p.TransactionId
}
