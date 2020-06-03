package models

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQmUndelivered_Validate(t *testing.T) {

	t.Run("Validate good QmUndelivered",
		testQmUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			true))
	t.Run("Validate missing questionnaire ID",
		testQmUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx"}`,
			false))
	t.Run("Validate missing time created",
		testQmUndeliveredValidate(`{"transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			false))
	t.Run("Validate missing transaction ID",
		testQmUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "questionnaireId": "01213213213"}`,
			false))

}

func testQmUndeliveredValidate(msgJson string, valid bool) func(*testing.T) {
	return func(t *testing.T) {
		qmUndelivered := QmUndelivered{}
		if err := json.Unmarshal([]byte(msgJson), &qmUndelivered); err != nil {
			t.Error(err)
			return
		}
		err := qmUndelivered.Validate()
		if valid {
			assert.NoError(t, err, "Validation failed for valid message")
		} else {
			assert.Error(t, err, "Validate did not error for invalid message")
		}
	}
}
