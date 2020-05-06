package models

import "errors"

type QmUndelivered struct {
	TransactionId   string      `json:"transactionId"`
	DateTime        HazyUtcTime `json:"dateTime"`
	QuestionnaireId string      `json:"questionnaireId"`
}

func (q QmUndelivered) GetTransactionId() string {
	return q.TransactionId
}

func (q QmUndelivered) Validate() error {
	if q.GetTransactionId() == "" {
		return errors.New("QmUndelivered missing transaction ID")
	}
	if q.DateTime.IsZero() {
		return errors.New("QmUndelivered missing dateTime")
	}
	if q.QuestionnaireId == "" {
		return errors.New("QmUndelivered missing questionnaire ID")
	}
	return nil
}
