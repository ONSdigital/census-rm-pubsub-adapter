package models

import "github.com/ONSdigital/census-rm-pubsub-adapter/validate"

type EqFulfilment struct {
	EqFulfilmentEvent   *RmEvent             `json:"event" validate:"required"`
	EqFulfilmentPayload *EqFulfilmentPayload `json:"payload" validate:"required"`
}

type EqFulfilmentPayload struct {
	FulfilmentRequest *FulfilmentRequest `json:"fulfilmentRequest" validate:"required"`
}

type FulfilmentRequest struct {
	FulfilmentCode   string   `json:"fulfilmentCode" validate:"required"`
	CaseID           string   `json:"caseId" validate:"required"`
	IndividualCaseID string   `json:"individualCaseId"`
	Contact          *Contact `json:"contact" validate:"required"`
}

type Contact struct {
	TelNo    string `json:"telNo,omitempty"`
	Title    string `json:"title,omitempty"`
	Forename string `json:"forename,omitempty"`
	Surname  string `json:"surname,omitempty"`
}

func (e EqFulfilment) GetTransactionId() string {
	return e.EqFulfilmentEvent.TransactionID
}

func (e EqFulfilment) Validate() error {
	return validate.Validate.Struct(e)
}
