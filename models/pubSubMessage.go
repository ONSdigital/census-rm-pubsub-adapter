package models

type PubSubMessage interface {
	GetTransactionId() string
	Validate() error
}
