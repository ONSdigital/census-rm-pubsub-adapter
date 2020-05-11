package models

import (
	"encoding/json"
	"testing"
)

func TestOfflineReceipt_Validate(t *testing.T) {

	t.Run("Validate good OfflineReceipt",
		testOfflineReceiptValidate(`{"dateTime": "2008-08-24T00:00:00", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			true))
	t.Run("Validate missing questionnaire ID",
		testOfflineReceiptValidate(`{"dateTime": "2008-08-24T00:00:00", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx"}`,
			false))
	t.Run("Validate missing time created",
		testOfflineReceiptValidate(`{"unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			false))
	t.Run("Validate missing transaction ID",
		testOfflineReceiptValidate(`{"dateTime": "2008-08-24T00:00:00", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "questionnaireId": "01213213213"}`,
			false))

}

func testOfflineReceiptValidate(msgJson string, valid bool) func(*testing.T) {
	return func(t *testing.T) {
		offlineReceipt := OfflineReceipt{}
		if err := json.Unmarshal([]byte(msgJson), &offlineReceipt); err != nil {
			t.Error(err)
			return
		}
		if err := offlineReceipt.Validate(); !valid && err == nil || valid && err != nil {
			t.Errorf("Got err: %s", err)
		}
	}
}
