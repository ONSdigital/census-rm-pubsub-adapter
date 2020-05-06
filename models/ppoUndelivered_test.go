package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestPpoUndelivered_Validate(t *testing.T) {

	t.Run("Validate good PpoUndelivered",
		testPpoUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "caseRef": "0123456789", "productCode": "P_TEST_1"}`,
			nil))
	t.Run("Validate missing case ref",
		testPpoUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "transactionId": "abc123xxx", "productCode": "P_TEST_1"}`,
			errors.New("PpoUndelivered missing case ref")))
	t.Run("Validate missing dateTime",
		testPpoUndeliveredValidate(`{"transactionId": "abc123xxx", "caseRef": "0123456789", "productCode": "P_TEST_1"}`,
			errors.New("PpoUndelivered missing dateTime")))
	t.Run("Validate missing transaction ID",
		testPpoUndeliveredValidate(`{"dateTime": "2008-08-24T00:00:00", "caseRef": "0123456789", "productCode": "P_TEST_1"}`,
			errors.New("PpoUndelivered missing transaction ID")))

}

func testPpoUndeliveredValidate(msgJson string, expectedErr error) func(*testing.T) {
	return func(t *testing.T) {
		ppoUndelivered := PpoUndelivered{}
		if err := json.Unmarshal([]byte(msgJson), &ppoUndelivered); err != nil {
			t.Error(err)
			return
		}
		if err := ppoUndelivered.Validate(); expectedErr != nil && err == nil || (err != nil && expectedErr != nil) && err.Error() != expectedErr.Error() {
			t.Errorf("Expected err: %s, got: %s", expectedErr, err)
		}
	}
}
