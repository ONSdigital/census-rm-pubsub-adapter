package models

import "errors"

type OfflineReceipt struct {
	TimeCreated     *HazyUtcTime `json:"dateTime"`
	TransactionId   string       `json:"transactionId"`
	QuestionnaireId string       `json:"questionnaireId"`
	Unreceipt       bool         `json:"unreceipt"`
	Channel         string       `json:"channel"`
}

func (o OfflineReceipt) GetTransactionId() string {
	return o.TransactionId
}

func (o OfflineReceipt) Validate() error {
	if o.GetTransactionId() == "" {
		return errors.New("OfflineReceipt missing transaction ID")
	}
	if o.TimeCreated == nil {
		return errors.New("OfflineReceipt missing dateTime")
	}
	if o.QuestionnaireId == "" {
		return errors.New("OfflineReceipt missing questionnaire ID")
	}
	return nil
}
