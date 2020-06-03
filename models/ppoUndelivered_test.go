package models

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPpoUndelivered_Validate(t *testing.T) {

	t.Run("Validate good PpoUndelivered",
		testPpoUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "caseRef": "0123456789", "productCode": "P_TEST_1"}`,
			true))
	t.Run("Validate missing case ref",
		testPpoUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "productCode": "P_TEST_1"}`,
			false))
	t.Run("Validate missing dateTime",
		testPpoUndeliveredValidate(`{"transactionId": "abc123xxx", "caseRef": "0123456789", "productCode": "P_TEST_1"}`,
			false))
	t.Run("Validate missing transaction ID",
		testPpoUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "caseRef": "0123456789", "productCode": "P_TEST_1"}`,
			false))

}

func testPpoUndeliveredValidate(msgJson string, valid bool) func(*testing.T) {
	return func(t *testing.T) {
		ppoUndelivered := PpoUndelivered{}
		if err := json.Unmarshal([]byte(msgJson), &ppoUndelivered); err != nil {
			t.Error(err)
			return
		}
		err := ppoUndelivered.Validate()
		if valid {
			assert.NoError(t, err, "Validation failed for valid message")
		} else {
			assert.Error(t, err, "Validate did not error for invalid message")
		}
	}
}
