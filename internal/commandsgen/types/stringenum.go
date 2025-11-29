// NOTE: this file is embedded inside the generated commands.gen.go output

package types

import (
	"fmt"
	"strings"
)

type StringEnum struct {
	Allowed            []string
	Value              string
	ChangedFromDefault bool
}

func NewStringEnum(allowed []string, value string) StringEnum {
	return StringEnum{Allowed: allowed, Value: value}
}

func (s *StringEnum) String() string { return s.Value }

func (s *StringEnum) Set(p string) error {
	for _, allowed := range s.Allowed {
		if p == allowed {
			s.Value = p
			s.ChangedFromDefault = true
			return nil
		}
	}
	return fmt.Errorf("%v is not one of required values of %v", p, strings.Join(s.Allowed, ", "))
}

func (*StringEnum) Type() string { return "string" }

type StringEnumArray struct {
	Allowed map[string]string
	Values  []string
}

func NewStringEnumArray(allowed []string, values []string) StringEnumArray {
	var allowedMap = make(map[string]string)
	for _, str := range allowed {
		allowedMap[strings.ToLower(str)] = str
	}
	return StringEnumArray{Allowed: allowedMap, Values: values}
}

func (s *StringEnumArray) String() string { return strings.Join(s.Values, ",") }

func (s *StringEnumArray) Set(p string) error {
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

func (*StringEnumArray) Type() string { return "string" }
