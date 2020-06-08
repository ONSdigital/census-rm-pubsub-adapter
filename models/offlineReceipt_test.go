package models

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
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
			assert.NoError(t, err)
			return
		}
		err := offlineReceipt.Validate()
		if valid {
			assert.NoError(t, err, "Validation failed for valid message")
		} else {
			assert.Error(t, err, "Validate did not error for invalid message")
		}
	}
}
