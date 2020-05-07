package models

import (
	"errors"
	"time"
)

type EqReceiptMetadata struct {
	TransactionId   string `json:"tx_id"`
	QuestionnaireId string `json:"questionnaire_id"`
	CaseID          string `json:"caseId,omitempty"`
}

type EqReceipt struct {
	TimeCreated *time.Time        `json:"timeCreated"`
	Metadata    EqReceiptMetadata `json:"metadata"`
}

func (e EqReceipt) GetTransactionId() string {
	return e.Metadata.TransactionId
}

func (e EqReceipt) Validate() error {
	if e.GetTransactionId() == "" {
		return errors.New("EqReceipt missing transaction ID")
	}
	if e.TimeCreated == nil {
		return errors.New("EqReceipt missing timeCreated")
	}
	if e.Metadata.QuestionnaireId == "" {
		return errors.New("EqReceipt missing questionnaire ID")
	}
	return nil
}
