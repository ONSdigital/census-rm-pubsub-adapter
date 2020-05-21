package models

import (
	"errors"
	"fmt"
	"github.com/ONSdigital/census-rm-pubsub-adapter/validate"
)

type FulfilmentConfirmed struct {
	TimeCreated     *HazyUtcTime `json:"dateTime" validate:"required"`
	TransactionId   string       `json:"transactionId" validate:"required"`
	ProductCode     string       `json:"productCode" validate:"required"`
	Channel         string       `json:"channel" validate:"required"`
	QuestionnaireId string       `json:"questionnaireId"`
	CaseRef         string       `json:"caseRef"`
}

func (o FulfilmentConfirmed) GetTransactionId() string {
	return o.TransactionId
}

func (o FulfilmentConfirmed) Validate() error {
	err := validate.Validate.Struct(o)

	if err == nil {
		if o.Channel == "QM" && len(o.QuestionnaireId) == 0 {
			err = errors.New(fmt.Sprintf("Missing questionnaire ID in QM message: %T, tx_id: %q", o, o.GetTransactionId()))
		} else if o.Channel == "PPO" && len(o.CaseRef) == 0 {
			err = errors.New(fmt.Sprintf("Missing case ref in PPO message: %T, tx_id: %q", o, o.GetTransactionId()))
		} else {
			err = errors.New(fmt.Sprintf("Unexpected channel: %T, tx_id: %q", o, o.GetTransactionId()))
		}
	}

	return err
}
