package models

type RmResponse struct {
	CaseID          *string `json:"caseId"` // Why a reference type?
	QuestionnaireID string  `json:"questionnaireId"`
	Unreceipt       bool    `json:"unreceipt"`
}

type RmPayload struct {
	Response RmResponse `json:"response"`
}

type RmEvent struct {
	Type          string `json:"type"`
	Source        string `json:"source"`
	Channel       string `json:"channel"`
	DateTime      string `json:"dateTime"`
	TransactionID string `json:"transactionId"`
}

type RmMessage struct {
	Event   RmEvent   `json:"event"`
	Payload RmPayload `json:"payload"`
}
