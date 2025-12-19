package cliext

import (
	"bytes"
	"fmt"
	"sort"
)

// inlineStringMap wraps a map to marshal as an inline TOML table.
type inlineStringMap map[string]string

func (m inlineStringMap) MarshalTOML() ([]byte, error) {
	if len(m) == 0 {
		return nil, nil
	}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{ ")
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "%s = %q", k, m[k])
	}
	buf.WriteString(" }")
	return buf.Bytes(), nil
}
