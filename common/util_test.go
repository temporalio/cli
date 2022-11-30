package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func (s *utilSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(utilSuite))
}

type utilSuite struct {
	*require.Assertions
	suite.Suite
}

func (s *utilSuite) TestStringToEnum_MapCaseInsensitive() {
	enumValues := map[string]int32{
		"Unspecified": 0,
		"Transfer":    1,
		"Timer":       2,
		"Replication": 3,
	}

	result, err := StringToEnum("timeR", enumValues)
	s.NoError(err)
	s.Equal(result, int32(2)) // Timer
}

func (s *utilSuite) TestStringToEnum_MapNonExisting() {
	enumValues := map[string]int32{
		"Unspecified": 0,
		"Transfer":    1,
		"Timer":       2,
		"Replication": 3,
	}

	result, err := StringToEnum("Timer2", enumValues)
	s.Error(err)
	s.Equal(result, int32(0))
}

func (s *utilSuite) TestStringToEnum_MapEmptyValue() {
	enumValues := map[string]int32{
		"Unspecified": 0,
		"Transfer":    1,
		"Timer":       2,
		"Replication": 3,
	}

	result, err := StringToEnum("", enumValues)
	s.NoError(err)
	s.Equal(result, int32(0))
}

func (s *utilSuite) TestStringToEnum_MapEmptyEnum() {
	enumValues := map[string]int32{}

	result, err := StringToEnum("Timer", enumValues)
	s.Error(err)
	s.Equal(result, int32(0))
}

func (s *utilSuite) TestParseKeyValuePairs() {
	tests := map[string]struct {
		input   []string
		want    map[string]string
		wantErr bool
	}{
		"simple values": {
			input: []string{
				"key1=value1",
				"key2=value2",
				"key3=value3=with=equal",
				"key4=value4:with-symbols",
				"key5=",
				`key6={"Auth":{"Enabled":false,"Options":["audience","organization"]},"ShowTemporalSystemNamespace":true}`,
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3=with=equal",
				"key4": "value4:with-symbols",
				"key5": "",
				"key6": `{"Auth":{"Enabled":false,"Options":["audience","organization"]},"ShowTemporalSystemNamespace":true}`,
			},
		},
		"no values": {
			input: []string{},
			want:  map[string]string{},
		},
		"empty": {
			input:   []string{""},
			wantErr: true,
		},
		"no separator": {
			input:   []string{"key:value"},
			wantErr: true,
		},
		"no key": {
			input:   []string{"=value"},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		s.Run(name, func() {
			got, err := SplitKeyValuePairs(tt.input)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.Equal(tt.want, got)
			}
		})
	}
}

// TestParseTime tests the parsing of date argument in UTC and UnixNano formats
func (s *utilSuite) TestParseTime() {
	t, err := ParseTime("", time.Date(1978, 8, 22, 0, 0, 0, 0, time.UTC), time.Now().UTC())
	s.NoError(err)
	s.Equal("1978-08-22 00:00:00 +0000 UTC", t.String())

	t, err = ParseTime("2018-06-07T15:04:05+07:00", time.Time{}, time.Now())
	s.NoError(err)
	s.Equal("2018-06-07T15:04:05+07:00", t.Format(time.RFC3339))

	expected, err := time.Parse(defaultDateTimeFormat, "2018-06-07T15:04:05+07:00")
	s.NoError(err)

	t, err = ParseTime("1528358645000000000", time.Time{}, time.Now().UTC())
	s.NoError(err)
	s.Equal(expected.UTC(), t)
}

// TestParseTimeDateRange tests the parsing of date argument in time range format, N<duration>
// where N is the integral multiplier, and duration can be second/minute/hour/day/week/month/year
func (s *utilSuite) TestParseTimeDateRange() {
	now := time.Now().UTC()
	tests := []struct {
		timeStr  string    // input
		defVal   time.Time // input
		expected time.Time // expected unix nano (approx)
	}{
		{
			timeStr:  "1s",
			defVal:   time.Time{},
			expected: now.Add(-time.Second),
		},
		{
			timeStr:  "100second",
			defVal:   time.Time{},
			expected: now.Add(-100 * time.Second),
		},
		{
			timeStr:  "2m",
			defVal:   time.Time{},
			expected: now.Add(-2 * time.Minute),
		},
		{
			timeStr:  "200minute",
			defVal:   time.Time{},
			expected: now.Add(-200 * time.Minute),
		},
		{
			timeStr:  "3h",
			defVal:   time.Time{},
			expected: now.Add(-3 * time.Hour),
		},
		{
			timeStr:  "1000hour",
			defVal:   time.Time{},
			expected: now.Add(-1000 * time.Hour),
		},
		{
			timeStr:  "5d",
			defVal:   time.Time{},
			expected: now.Add(-5 * day),
		},
		{
			timeStr:  "25day",
			defVal:   time.Time{},
			expected: now.Add(-25 * day),
		},
		{
			timeStr:  "5w",
			defVal:   time.Time{},
			expected: now.Add(-5 * week),
		},
		{
			timeStr:  "52week",
			defVal:   time.Time{},
			expected: now.Add(-52 * week),
		},
		{
			timeStr:  "3M",
			defVal:   time.Time{},
			expected: now.Add(-3 * month),
		},
		{
			timeStr:  "6month",
			defVal:   time.Time{},
			expected: now.Add(-6 * month),
		},
		{
			timeStr:  "1y",
			defVal:   time.Time{},
			expected: now.Add(-year),
		},
		{
			timeStr:  "7year",
			defVal:   time.Time{},
			expected: now.Add(-7 * year),
		},
		{
			timeStr:  "100y", // epoch time will be returned as that's the minimum unix timestamp possible
			defVal:   time.Time{},
			expected: time.Unix(0, 0).UTC(),
		},
	}
	const delta = 5 * time.Millisecond
	for _, te := range tests {
		parsedTime, err := ParseTime(te.timeStr, te.defVal, now)
		s.NoError(err)

		s.True(te.expected.Before(parsedTime) || te.expected == parsedTime, "Case: %s. %d must be less or equal than parsed %d", te.timeStr, te.expected, parsedTime)
		s.True(te.expected.Add(delta).After(parsedTime) || te.expected.Add(delta) == parsedTime, "Case: %s. %d must be greater or equal than parsed %d", te.timeStr, te.expected, parsedTime)
	}
}
