package temporalcli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type StringEnum struct {
	Allowed []string
	Value   string
}

func NewStringEnum(allowed []string, value string) StringEnum {
	return StringEnum{Allowed: allowed, Value: value}
}

func (s *StringEnum) String() string { return s.Value }

func (s *StringEnum) Set(p string) error {
	for _, allowed := range s.Allowed {
		if p == allowed {
			s.Value = p
			return nil
		}
	}
	return fmt.Errorf("%v is not one of required values of %v", p, strings.Join(s.Allowed, ", "))
}

func (*StringEnum) Type() string { return "string" }

func stringToProtoEnum[T ~int32](s string, maps ...map[string]int32) (T, error) {
	// Go over each map looking, if not there, use first map to build set of
	// strings required
	for _, m := range maps {
		for k, v := range m {
			if strings.EqualFold(k, s) {
				return T(v), nil
			}
		}
	}
	keys := make([]string, 0, len(maps[0]))
	for k := range maps[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return 0, fmt.Errorf("unknown value %q, expected one of: %v", s, strings.Join(keys, ", "))
}

func stringKeysValues(s []string) (map[string]string, error) {
	ret := make(map[string]string, len(s))
	for _, item := range s {
		pieces := strings.SplitN(item, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("missing expected '=' in %q", item)
		}
		ret[pieces[0]] = pieces[1]
	}
	return ret, nil
}

func stringKeysJSONValues(s []string) (map[string]any, error) {
	ret := make(map[string]any, len(s))
	for _, item := range s {
		pieces := strings.SplitN(item, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("missing expected '=' in %q", item)
		}
		var v any
		if err := json.Unmarshal([]byte(pieces[1]), &v); err != nil {
			return nil, fmt.Errorf("invalid JSON value for key %q: %w", pieces[0], err)
		}
		ret[pieces[0]] = v
	}
	return ret, nil
}
