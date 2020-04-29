package models

import (
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalJSON(t *testing.T) {
	expectedTimeUtc, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2008-08-24T00:00:00Z")
	expectedTimeBst, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2008-08-24T00:00:00+01:00")

	t.Run("Without TZ", testUnmarshalJSON([]byte(`"2008-08-24T00:00:00"`), &expectedTimeUtc))
	t.Run("With zulu TZ", testUnmarshalJSON([]byte(`"2008-08-24T00:00:00Z"`), &expectedTimeUtc))
	t.Run("With explicit UTC TZ", testUnmarshalJSON([]byte(`"2008-08-24T00:00:00+00:00"`), &expectedTimeUtc))
	t.Run("With explicit non zero TZ", testUnmarshalJSON([]byte(`"2008-08-24T00:00:00+01:00"`), &expectedTimeBst))
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

		_, actualTzOffset := hazyUtcTime.Zone()
		_, expectedTzOffset := expectedTime.Zone()
		if actualTzOffset != expectedTzOffset {
			t.Errorf("Expected TZ offset: %d, got: %d", expectedTzOffset, actualTzOffset)
		}
	}
}
