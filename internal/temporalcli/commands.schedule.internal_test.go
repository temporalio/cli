package temporalcli

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
)

func TestScheduleUpdateSearchAttributes_IntegerPrecision(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int64
	}{
		{"9007199254740993", 9007199254740993}, // Not exactly representable as float64.
		{"9223372036854775807", math.MaxInt64},
		{"-9223372036854775808", math.MinInt64},
	} {
		t.Run(tc.input, func(t *testing.T) {
			c := TemporalScheduleUpdateCommand{
				ScheduleSearchAttributesOptions: ScheduleSearchAttributesOptions{
					ScheduleSearchAttribute: []string{"Count=" + tc.input},
				},
			}
			attributes, err := c.searchAttributes()
			require.NoError(t, err)
			value, ok := attributes.GetInt64(temporal.NewSearchAttributeKeyInt64("Count"))
			require.True(t, ok)
			require.Equal(t, tc.want, value)
		})
	}
}
