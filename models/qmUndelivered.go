package models

type QmUndelivered struct {
	TransactionId   string      `json:"transactionId"`
	DateTime        HazyUtcTime `json:"dateTime"`
	QuestionnaireId string      `json:"questionnaireId"`
}

func (q QmUndelivered) GetTransactionId() string {
	return q.TransactionId
}

func (q QmUndelivered) Validate() bool {
	return q.GetTransactionId() != ""
}
