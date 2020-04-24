package models

import "time"

type EqReceiptMetadata struct {
	TransactionID   string `json:"tx_id"`
	QuestionnaireID string `json:"questionnaire_id"`
	CaseID          string `json:"caseId,omitempty"`
}

type EqReceipt struct {
	TimeCreated time.Time         `json:"timeCreated"`
	Metadata    EqReceiptMetadata `json:"metadata"`
}
