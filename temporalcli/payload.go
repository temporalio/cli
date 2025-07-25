package temporalcli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"go.temporal.io/api/common/v1"
)

func CreatePayloads(data [][]byte, metadata map[string][][]byte, isBase64 bool) (*common.Payloads, error) {
	ret := &common.Payloads{Payloads: make([]*common.Payload, len(data))}
	for i, in := range data {
		var metadataForIndex = map[string][]byte{}
		for k, vals := range metadata {
			if len(vals) == 0 {
				continue
			}
			v := vals[0]
			if len(vals) > i {
				v = vals[i]
			}
			// If it's JSON, validate it
			if k == "encoding" && strings.HasPrefix(string(v), "json/") && !json.Valid(in) {
				return nil, fmt.Errorf("input #%v is not valid JSON", i+1)
			}
			metadataForIndex[k] = v
		}
		// Decode base64 if base64'd (std encoding only for now)
		if isBase64 {
			var err error
			if in, err = base64.StdEncoding.DecodeString(string(in)); err != nil {
				return nil, fmt.Errorf("input #%v is not valid base64", i+1)
			}
		}
		ret.Payloads[i] = &common.Payload{Data: in, Metadata: metadataForIndex}
	}
	return ret, nil
}
