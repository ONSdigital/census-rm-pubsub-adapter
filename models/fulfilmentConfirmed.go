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

func (f FulfilmentConfirmed) GetTransactionId() string {
	return f.TransactionId
}

func (f FulfilmentConfirmed) Validate() error {
	err := validate.Validate.Struct(f)

	if err == nil {
		if f.Channel == "QM" && len(f.QuestionnaireId) == 0 {
			err = errors.New(fmt.Sprintf("Missing questionnaire ID in QM message: %T, tx_id: %q", f, f.GetTransactionId()))
		} else if f.Channel == "PPO" && len(f.CaseRef) == 0 {
			err = errors.New(fmt.Sprintf("Missing case ref in PPO message: %T, tx_id: %q", f, f.GetTransactionId()))
		} else if f.Channel != "QM" && f.Channel != "PPO" {
			err = errors.New(fmt.Sprintf("Unexpected channel: %T, tx_id: %q", f, f.GetTransactionId()))
		}
	}

	return err
}
