package models

type PpoUndelivered struct {
	TransactionId string      `json:"transactionId"`
	DateTime      HazyUtcTime `json:"dateTime"`
	CaseRef       string      `json:"caseRef"`
	ProductCode   string      `json:"productCode"`
}

func (p PpoUndelivered) GetTransactionId() string {
	return p.TransactionId
}

func (p PpoUndelivered) Validate() bool {
	return p.GetTransactionId() != ""
}
