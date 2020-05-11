package models

import (
	"encoding/json"
	"testing"
)

func TestEqReceipt_Validate(t *testing.T) {

	t.Run("Validate good EqReceipt",
		testEqReceiptValidate(`{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`,
			true))
	t.Run("Validate missing questionnaire ID",
		testEqReceiptValidate(`{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"tx_id": "abc123xxx"}}`,
			false))
	t.Run("Validate missing time created",
		testEqReceiptValidate(`{"metadata": {"tx_id": "abc123xxx", "questionnaire_id": "01213213213"}}`,
			false))
	t.Run("Validate missing transaction ID",
		testEqReceiptValidate(`{"timeCreated": "2008-08-24T00:00:00Z", "metadata": {"questionnaire_id": "01213213213"}}`,
			false))

}

func testEqReceiptValidate(msgJson string, valid bool) func(*testing.T) {
	return func(t *testing.T) {
		eqReceipt := EqReceipt{}
		if err := json.Unmarshal([]byte(msgJson), &eqReceipt); err != nil {
			t.Error(err)
			return
		}
		if err := eqReceipt.Validate(); !valid && err == nil || valid && err != nil {
			t.Errorf("Got err: %s", err)
		}
	}
}
