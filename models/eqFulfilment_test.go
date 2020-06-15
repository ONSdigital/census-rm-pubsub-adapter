package models

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEqFulfilment_Validate(t *testing.T) {
	t.Run("Valid EQ fulfilment request", testEqFulfilmentValidate(
		`{
		"event" : {
			"type" : "FULFILMENT_REQUESTED",
			"source" : "QUESTIONNAIRE_RUNNER",
			"channel" : "EQ",
			"dateTime" : "2011-08-12T20:17:46.384Z",
			"transactionId" : "c45de4dc-3c3b-11e9-b210-d663bd873d93"
		},
		"payload" : {
			"fulfilmentRequest" : {
				"fulfilmentCode": "UACIT1",
				"caseId" : "bbd55984-0dbf-4499-bfa7-0aa4228700e9",
				"individualCaseID": "c8e11403-9392-48e3-bfd5-5e8350ca1c4e",
				"contact": {
					"telNo":"+447890000000"
				}
			}
		}
	}`, true))

	t.Run("Valid EQ fulfilment request without contact block", testEqFulfilmentValidate(
		`{	
		"event" : {
			"type" : "FULFILMENT_REQUESTED",
			"source" : "QUESTIONNAIRE_RUNNER",
			"channel" : "EQ",
			"dateTime" : "2011-08-12T20:17:46.384Z",
			"transactionId" : "c45de4dc-3c3b-11e9-b210-d663bd873d93"
		  },
		"payload" : {
			"fulfilmentRequest" : {
			"fulfilmentCode": "P_UAC_UACIP1",
			"caseId" : "bbd55984-0dbf-4499-bfa7-0aa4228700e9",
			"contact" : {
				"title" : "Mr",
				"forename" : "Testy",
				"surname" : "McTestface"
				}
			}
		}
	}`, true))

	t.Run("Invalid EQ fulfilment request missing transaction ID", testEqFulfilmentValidate(
		`{
		"event" : {
			"type" : "FULFILMENT_REQUESTED",
			"source" : "QUESTIONNAIRE_RUNNER",
			"channel" : "EQ",
			"dateTime" : "2011-08-12T20:17:46.384Z"
		},
		"payload" : {
			"fulfilmentRequest" : {
				"fulfilmentCode": "UACIT1",
				"caseId" : "bbd55984-0dbf-4499-bfa7-0aa4228700e9",
				"contact": {
					"telNo":"+447890000000"
				}
			}
		}
	}`, false))

	t.Run("Invalid EQ fulfilment request missing case ID", testEqFulfilmentValidate(
		`{
		"event" : {
			"type" : "FULFILMENT_REQUESTED",
			"source" : "QUESTIONNAIRE_RUNNER",
			"channel" : "EQ",
			"dateTime" : "2011-08-12T20:17:46.384Z",
			"transactionId" : "c45de4dc-3c3b-11e9-b210-d663bd873d93"
		},
		"payload" : {
			"fulfilmentRequest" : {
				"fulfilmentCode": "UACIT1",
				"contact": {
					"telNo":"+447890000000"
				}
			}
		}
	}`, false))
}

func testEqFulfilmentValidate(msgJson string, valid bool) func(*testing.T) {
	return func(t *testing.T) {
		eqFulfilment := EqFulfilment{}
		if err := json.Unmarshal([]byte(msgJson), &eqFulfilment); err != nil {
			assert.NoError(t, err)
			return
		}
		err := eqFulfilment.Validate()
		if valid {
			assert.NoError(t, err, "Validation failed for valid message")
		} else {
			assert.Error(t, err, "Validate did not error for invalid message")
		}
	}
}
