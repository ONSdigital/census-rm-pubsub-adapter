package models

type PubSubMessage interface {
	GetQuestionnaireId() string
	GetTransactionID() string
}
