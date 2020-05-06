package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestEqReceipt_Validate(t *testing.T) {

	t.Run("Validate good EqReceipt",
		testEqReceiptValidate(`{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`,
			nil))
	t.Run("Validate missing questionnaire ID",
		testEqReceiptValidate(`{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx"}}`,
			errors.New("EqReceipt missing questionnaire ID")))
	t.Run("Validate missing time created",
		testEqReceiptValidate(`{"metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`,
			errors.New("EqReceipt missing timeCreated")))
	t.Run("Validate missing transaction ID",
		testEqReceiptValidate(`{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"questionnaire_id": "01213213213"}}`,
			errors.New("EqReceipt missing transaction ID")))

}

func testEqReceiptValidate(msgJson string, expectedErr error) func(*testing.T) {
	return func(t *testing.T) {
		eqReceipt := EqReceipt{}
		if err := json.Unmarshal([]byte(msgJson), &eqReceipt); err != nil {
			t.Error(err)
			return
		}
		if err := eqReceipt.Validate(); expectedErr != nil && err == nil || (err != nil && expectedErr != nil) && err.Error() != expectedErr.Error() {
			t.Errorf("Expected err: %s, got: %s", expectedErr, err)
		}
	}
}
