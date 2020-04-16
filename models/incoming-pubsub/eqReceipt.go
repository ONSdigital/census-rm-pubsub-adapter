package incoming_pubsub

type EqReceiptMetadata struct {
	TransactionID   string `json:"tx_id"`
	QuestionnaireID string `json:"questionnaire_id"`
}

type EqReceipt struct {
	TimeCreated string            `json:"timeCreated"`
	Metadata    EqReceiptMetadata `json:"metadata"`
}
