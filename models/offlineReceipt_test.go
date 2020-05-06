package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestOfflineReceipt_Validate(t *testing.T) {

	t.Run("Validate good OfflineReceipt",
		testOfflineReceiptValidate(`{"dateTime": "2008-08-24T00:00:00", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			nil))
	t.Run("Validate missing questionnaire ID",
		testOfflineReceiptValidate(`{"dateTime": "2008-08-24T00:00:00", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx"}`,
			errors.New("OfflineReceipt missing questionnaire ID")))
	t.Run("Validate missing time created",
		testOfflineReceiptValidate(`{"unreceipt" : false, "channel" : "INTEGRATION_TEST", "transactionId": "abc123xxx", "questionnaireId": "01213213213"}`,
			errors.New("OfflineReceipt missing dateTime")))
	t.Run("Validate missing transaction ID",
		testOfflineReceiptValidate(`{"dateTime": "2008-08-24T00:00:00", "unreceipt" : false, "channel" : "INTEGRATION_TEST", "questionnaireId": "01213213213"}`,
			errors.New("OfflineReceipt missing transaction ID")))

}

func testOfflineReceiptValidate(msgJson string, expectedErr error) func(*testing.T) {
	return func(t *testing.T) {
		offlineReceipt := OfflineReceipt{}
		if err := json.Unmarshal([]byte(msgJson), &offlineReceipt); err != nil {
			t.Error(err)
			return
		}
		if err := offlineReceipt.Validate(); expectedErr != nil && err == nil || (err != nil && expectedErr != nil) && err.Error() != expectedErr.Error() {
			t.Errorf("Expected err: %s, got: %s", expectedErr, err)
		}
	}
}
