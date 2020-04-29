package models

import (
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalJSON(t *testing.T) {
	expectedTimeUtc, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2008-08-24T00:00:00Z")

	t.Run("Without TZ", testUnmarshalJSON([]byte(`"2008-08-24T00:00:00"`), &expectedTimeUtc))
	t.Run("With zulu TZ", testUnmarshalJSON([]byte(`"2008-08-24T00:00:00Z"`), &expectedTimeUtc))
	t.Run("With explicit TZ", testUnmarshalJSON([]byte(`"2008-08-24T00:00:00+00:00"`), &expectedTimeUtc))
}

func testUnmarshalJSON(timeBuf []byte, expectedTime *time.Time) func(t *testing.T) {
	return func(t *testing.T) {

		hazyUtcTime := HazyUtcTime{}
		if err := hazyUtcTime.UnmarshalJSON(timeBuf); err != nil {
			t.Error(err)
			return
		}

		if reflect.DeepEqual(expectedTime, hazyUtcTime) {
			t.Errorf("Expected time: %s, got: %s", expectedTime, hazyUtcTime)
		}

		actualTz, _ := hazyUtcTime.Zone()
		if reflect.DeepEqual(0, actualTz) {
			t.Errorf("Expected UTC TZ, got: %s", actualTz)
		}
	}
}
