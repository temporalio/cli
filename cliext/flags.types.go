package cliext

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FlagStringEnum is a flag type that restricts values to a predefined set.
type FlagStringEnum struct {
	Allowed            []string
	Value              string
	ChangedFromDefault bool
}

// NewFlagStringEnum creates a new FlagStringEnum with the given allowed values and default.
func NewFlagStringEnum(allowed []string, value string) FlagStringEnum {
	return FlagStringEnum{Allowed: allowed, Value: value}
}

func (s *FlagStringEnum) String() string { return s.Value }

func (s *FlagStringEnum) Set(p string) error {
	for _, allowed := range s.Allowed {
		if p == allowed {
			s.Value = p
			s.ChangedFromDefault = true
			return nil
		}
	}
	return fmt.Errorf("%v is not one of required values of %v", p, strings.Join(s.Allowed, ", "))
}

func (*FlagStringEnum) Type() string { return "string" }

// FlagStringEnumArray is a flag type that accumulates multiple values from a predefined set.
type FlagStringEnumArray struct {
	Allowed map[string]string
	Values  []string
}

// NewFlagStringEnumArray creates a new FlagStringEnumArray with the given allowed values and defaults.
func NewFlagStringEnumArray(allowed []string, values []string) FlagStringEnumArray {
	var allowedMap = make(map[string]string)
	for _, str := range allowed {
		allowedMap[strings.ToLower(str)] = str
	}
	return FlagStringEnumArray{Allowed: allowedMap, Values: values}
}

func (s *FlagStringEnumArray) String() string { return strings.Join(s.Values, ",") }

func (s *FlagStringEnumArray) Set(p string) error {
	val, ok := s.Allowed[strings.ToLower(p)]
	if !ok {
		values := make([]string, 0, len(s.Allowed))
		for _, v := range s.Allowed {
			values = append(values, v)
		}
		return fmt.Errorf("invalid value: %s, allowed values are: %s", p, strings.Join(values, ", "))
	}
	s.Values = append(s.Values, val)
	return nil
}

func (*FlagStringEnumArray) Type() string { return "string" }

// FlagDuration extends time.Duration with support for days ("d" suffix).
type FlagDuration time.Duration

var reFlagDays = regexp.MustCompile(`(\d+(\.\d*)?|(\.\d+))d`)

// ParseFlagDuration is like time.ParseDuration, but supports unit "d" for days
// (always interpreted as exactly 24 hours).
func ParseFlagDuration(s string) (time.Duration, error) {
	s = reFlagDays.ReplaceAllStringFunc(s, func(v string) string {
		fv, err := strconv.ParseFloat(strings.TrimSuffix(v, "d"), 64)
		if err != nil {
			return v // will cause time.ParseDuration to return an error
		}
		return fmt.Sprintf("%fh", 24*fv)
	})
	return time.ParseDuration(s)
}

// MustParseFlagDuration parses a duration string and panics on error.
// Used in generated code for compile-time constants.
func MustParseFlagDuration(s string) FlagDuration {
	dur, err := ParseFlagDuration(s)
	if err != nil {
		panic(fmt.Sprintf("invalid duration %q: %v", s, err))
	}
	return FlagDuration(dur)
}

func (d FlagDuration) Duration() time.Duration {
	return time.Duration(d)
}

func (d *FlagDuration) String() string {
	return d.Duration().String()
}

func (d *FlagDuration) Set(s string) error {
	p, err := ParseFlagDuration(s)
	if err != nil {
		return err
	}
	*d = FlagDuration(p)
	return nil
}

func (d *FlagDuration) Type() string {
	return "duration"
}

// FlagTimestamp wraps time.Time with RFC3339 format for flags.
type FlagTimestamp time.Time

func (t FlagTimestamp) Time() time.Time {
	return time.Time(t)
}

func (t *FlagTimestamp) String() string {
	return t.Time().Format(time.RFC3339)
}

func (t *FlagTimestamp) Set(s string) error {
	p, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = FlagTimestamp(p)
	return nil
}

func (t *FlagTimestamp) Type() string {
	return "timestamp"
}
