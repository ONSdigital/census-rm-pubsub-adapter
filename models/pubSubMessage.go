package models

type PubSubMessage interface {
	GetTransactionID() string
}
