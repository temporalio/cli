package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/temporalio/cli/internal/commandsgen/types"
)

type ParseDurationSuite struct {
	suite.Suite
}

func TestParseDurationSuite(t *testing.T) {
	suite.Run(t, new(ParseDurationSuite))
}

func (s *ParseDurationSuite) TestParseDuration() {
	for _, c := range []struct {
		input    string
		expected time.Duration // -1 means error
	}{
		{"1h", time.Hour},
		{"3m30s", 3*time.Minute + 30*time.Second},
		{"1d", 24 * time.Hour},
		{"3d", 3 * 24 * time.Hour},
		{"5d6h15m", 5*24*time.Hour + 6*time.Hour + 15*time.Minute},
		{"5.25d15m", 5*24*time.Hour + 6*time.Hour + 15*time.Minute},
		{".5d", 12 * time.Hour},
		{"-10d12.25h", -(10*24*time.Hour + 12*time.Hour + 15*time.Minute)},
		{"3m2h1d", 3*time.Minute + 2*time.Hour + 1*24*time.Hour},
		{"8m7h6d5d4h3m", 8*time.Minute + 7*time.Hour + 6*24*time.Hour + 5*24*time.Hour + 4*time.Hour + 3*time.Minute},
		{"7", -1},         // error
		{"", -1},          // error
		{"10000000h", -1}, // error out of bounds
	} {
		got, err := types.ParseDuration(c.input)
		if c.expected == -1 {
			s.Error(err)
		} else {
			s.Equal(c.expected, got)
		}
	}
}
