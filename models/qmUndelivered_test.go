package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestQmUndelivered_Validate(t *testing.T) {

	t.Run("Validate good QmUndelivered",
		testQmUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			nil))
	t.Run("Validate missing questionnaire ID",
		testQmUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx"}`,
			errors.New("QmUndelivered missing questionnaire ID")))
	t.Run("Validate missing time created",
		testQmUndeliveredValidate(`{"transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			errors.New("QmUndelivered missing dateTime")))
	t.Run("Validate missing transaction ID",
		testQmUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "questionnaireId": "01213213213"}`,
			errors.New("QmUndelivered missing transaction ID")))

}

func testQmUndeliveredValidate(msgJson string, expectedErr error) func(*testing.T) {
	return func(t *testing.T) {
		qmUndelivered := QmUndelivered{}
		if err := json.Unmarshal([]byte(msgJson), &qmUndelivered); err != nil {
			t.Error(err)
			return
		}
		if err := qmUndelivered.Validate(); expectedErr != nil && err == nil || (err != nil && expectedErr != nil) && err.Error() != expectedErr.Error() {
			t.Errorf("Expected err: %s, got: %s", expectedErr, err)
		}
	}
}
