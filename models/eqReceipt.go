package models

import (
	"time"
)

type EqReceiptMetadata struct {
	TransactionId   string `json:"tx_id"`
	QuestionnaireId string `json:"questionnaire_id"`
	CaseID          string `json:"caseId,omitempty"`
}

type EqReceipt struct {
	TimeCreated time.Time         `json:"timeCreated"`
	Metadata    EqReceiptMetadata `json:"metadata"`
}

func (e EqReceipt) GetTransactionId() string {
	return e.Metadata.TransactionId
}

func (e EqReceipt) Validate() bool {
	return e.GetTransactionId() != ""
}
