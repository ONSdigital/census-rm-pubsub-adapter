package models

type FulfilmentInformation struct {
	CaseRef         string `json:"caseRef,omitempty"`
	FulfilmentCode  string `json:"fulfilmentCode,omitempty"`
	QuestionnaireId string `json:"questionnaireId,omitempty"`
}
