package models

import "time"

type RmResponse struct {
	CaseID          string `json:"caseId,omitempty"`
	QuestionnaireID string `json:"questionnaireId"`
	Unreceipt       bool   `json:"unreceipt"`
}

type RmPayload struct {
	Response              *RmResponse            `json:"response,omitempty"`
	FulfilmentInformation *FulfilmentInformation `json:"fulfilmentInformation,omitempty"`
}

type RmEvent struct {
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	Channel       string    `json:"channel"`
	DateTime      time.Time `json:"dateTime"`
	TransactionID string    `json:"transactionId"`
}

type RmMessage struct {
	Event   RmEvent   `json:"event"`
	Payload RmPayload `json:"payload"`
}

type FulfilmentInformation struct {
	CaseRef         string `json:"caseRef,omitempty"`
	FulfilmentCode  string `json:"fulfilmentCode,omitempty"`
	QuestionnaireId string `json:"questionnaireId,omitempty"`
}
