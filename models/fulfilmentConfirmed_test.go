package models

import (
	"encoding/json"
	"testing"
)

func TestFulfilmentConfirmed_Validate_Validate(t *testing.T) {

	t.Run("Validate good QM FulfilmentConfirmed",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","questionnaireId":"1100000000112","productCode":"P_OR_H1","channel":"QM","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			true))
	t.Run("Validate good PPO FulfilmentConfirmed",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","caseRef":"12345678","productCode":"P_OR_H1","channel":"PPO","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			true))
	t.Run("Validate missing product code",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","questionnaireId":"1100000000112","channel":"QM","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			false))
	t.Run("Validate missing time created",
		testFulfilmentConfirmedValidate(`{"questionnaireId":"1100000000112","productCode":"P_OR_H1","channel":"QM","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			false))
	t.Run("Validate missing transaction ID",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","questionnaireId":"1100000000112","productCode":"P_OR_H1","channel":"QM","type":"FULFILMENT_CONFIRMED"}`,
			false))
	t.Run("Validate missing channel",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","questionnaireId":"1100000000112","productCode":"P_OR_H1","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			false))
	t.Run("Validate missing questionnaire ID",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","productCode":"P_OR_H1","channel":"QM","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			false))
	t.Run("Validate missing case ref",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","productCode":"P_OR_H1","channel":"PPO","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			false))
	t.Run("Validate dodgy channel",
		testFulfilmentConfirmedValidate(`{"dateTime":"2019-08-03T14:30:01","questionnaireId":"1100000000112","productCode":"P_OR_H1","channel":"NOODLEBANANA","type":"FULFILMENT_CONFIRMED","transactionId":"92971ad5-c534-48af-a8b3-92484b14ceef"}`,
			false))
}

func testFulfilmentConfirmedValidate(msgJson string, valid bool) func(*testing.T) {
	return func(t *testing.T) {
		fulfilmentConfirmed := FulfilmentConfirmed{}
		if err := json.Unmarshal([]byte(msgJson), &fulfilmentConfirmed); err != nil {
			t.Error(err)
			return
		}
		if err := fulfilmentConfirmed.Validate(); !valid && err == nil || valid && err != nil {
			t.Errorf("Got err: %s", err)
		}
	}
}
