package models

import "errors"

type PpoUndelivered struct {
	TransactionId string       `json:"transactionId"`
	DateTime      *HazyUtcTime `json:"dateTime"`
	CaseRef       string       `json:"caseRef"`
	ProductCode   string       `json:"productCode"`
}

func (p PpoUndelivered) GetTransactionId() string {
	return p.TransactionId
}

func (p PpoUndelivered) Validate() error {
	if p.GetTransactionId() == "" {
		return errors.New("PpoUndelivered missing transaction ID")
	}
	if p.DateTime == nil {
		return errors.New("PpoUndelivered missing dateTime")
	}
	if p.CaseRef == "" {
		return errors.New("PpoUndelivered missing case ref")
	}
	return nil
}
