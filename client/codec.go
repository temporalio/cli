// The MIT License
//
// Copyright (c) 2021 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Vendored code from sdk-go.

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

// RemotePayloadCodecOptions are options for RemotePayloadCodec.
// Client is optional.
type RemotePayloadCodecOptions struct {
	Endpoint      string
	ModifyRequest func(*http.Request) error
	Client        http.Client
}

type remotePayloadCodec struct {
	options RemotePayloadCodecOptions
}

const remotePayloadCodecEncodePath = "/encode"
const remotePayloadCodecDecodePath = "/decode"

// NewRemotePayloadCodec creates a PayloadCodec using the remote endpoint configured by RemotePayloadCodecOptions.
func NewRemotePayloadCodec(options RemotePayloadCodecOptions) converter.PayloadCodec {
	return &remotePayloadCodec{options}
}

// Encode uses the remote payload codec endpoint to encode payloads.
func (pc *remotePayloadCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return pc.encodeOrDecode(pc.options.Endpoint+remotePayloadCodecEncodePath, payloads)
}

// Decode uses the remote payload codec endpoint to decode payloads.
func (pc *remotePayloadCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return pc.encodeOrDecode(pc.options.Endpoint+remotePayloadCodecDecodePath, payloads)
}

func (pc *remotePayloadCodec) encodeOrDecode(endpoint string, payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	requestPayloads, err := json.Marshal(commonpb.Payloads{Payloads: payloads})
	if err != nil {
		return payloads, fmt.Errorf("unable to marshal payloads: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestPayloads))
	if err != nil {
		return payloads, fmt.Errorf("unable to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if pc.options.ModifyRequest != nil {
		err = pc.options.ModifyRequest(req)
		if err != nil {
			return payloads, err
		}
	}

	response, err := pc.options.Client.Do(req)
	if err != nil {
		return payloads, err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode == 200 {
		bs, err := io.ReadAll(response.Body)
		if err != nil {
			return payloads, fmt.Errorf("failed to read response body: %w", err)
		}
		var resultPayloads commonpb.Payloads
		err = jsonpb.UnmarshalString(string(bs), &resultPayloads)
		if err != nil {
			return payloads, fmt.Errorf("unable to unmarshal payloads: %w", err)
		}
		if len(payloads) != len(resultPayloads.Payloads) {
			return payloads, fmt.Errorf("received %d payloads from remote codec, expected %d", len(resultPayloads.Payloads), len(payloads))
		}
		return resultPayloads.Payloads, nil
	}

	message, _ := io.ReadAll(response.Body)
	return payloads, fmt.Errorf("%s: %s", http.StatusText(response.StatusCode), message)
}

type remoteDataConverter struct {
	parent       converter.DataConverter
	payloadCodec converter.PayloadCodec
}

// NewRemoteDataConverter wraps the given parent DataConverter and performs
// encoding/decoding on the payload via the remote endpoint.
func NewRemoteDataConverter(parent converter.DataConverter, options converter.RemoteDataConverterOptions) converter.DataConverter {
	options.Endpoint = strings.TrimSuffix(options.Endpoint, "/")
	payloadCodec := NewRemotePayloadCodec(RemotePayloadCodecOptions(options))
	return &remoteDataConverter{parent, payloadCodec}
}

// ToPayload implements DataConverter.ToPayload performing remote encoding on the
// result of the parent's ToPayload call.
func (rdc *remoteDataConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	payload, err := rdc.parent.ToPayload(value)
	if payload == nil || err != nil {
		return payload, err
	}
	encodedPayloads, err := rdc.payloadCodec.Encode([]*commonpb.Payload{payload})
	if err != nil {
		return payload, err
	}
	return encodedPayloads[0], err
}

// ToPayloads implements DataConverter.ToPayloads performing remote encoding on the
// result of the parent's ToPayloads call.
func (rdc *remoteDataConverter) ToPayloads(value ...interface{}) (*commonpb.Payloads, error) {
	payloads, err := rdc.parent.ToPayloads(value...)
	if payloads == nil || err != nil {
		return payloads, err
	}
	encodedPayloads, err := rdc.payloadCodec.Encode(payloads.Payloads)
	return &commonpb.Payloads{Payloads: encodedPayloads}, err
}

// FromPayload implements DataConverter.FromPayload performing remote decoding on the
// given payload before sending to the parent FromPayload.
func (rdc *remoteDataConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	decodedPayloads, err := rdc.payloadCodec.Decode([]*commonpb.Payload{payload})
	if err != nil {
		return err
	}
	return rdc.parent.FromPayload(decodedPayloads[0], valuePtr)
}

// FromPayloads implements DataConverter.FromPayloads performing remote decoding on the
// given payloads before sending to the parent FromPayloads.
func (rdc *remoteDataConverter) FromPayloads(payloads *commonpb.Payloads, valuePtrs ...interface{}) error {
	if payloads == nil {
		return rdc.parent.FromPayloads(payloads, valuePtrs...)
	}

	decodedPayloads, err := rdc.payloadCodec.Decode(payloads.Payloads)
	if err != nil {
		return err
	}
	return rdc.parent.FromPayloads(&commonpb.Payloads{Payloads: decodedPayloads}, valuePtrs...)
}

// ToString implements DataConverter.ToString performing remote decoding on the given
// payload before sending to the parent ToString.
func (rdc *remoteDataConverter) ToString(payload *commonpb.Payload) string {
	if payload == nil {
		return rdc.parent.ToString(payload)
	}

	decodedPayloads, err := rdc.payloadCodec.Decode([]*commonpb.Payload{payload})
	if err != nil {
		return err.Error()
	}
	return rdc.parent.ToString(decodedPayloads[0])
}

// ToStrings implements DataConverter.ToStrings using ToString for each value.
func (rdc *remoteDataConverter) ToStrings(payloads *commonpb.Payloads) []string {
	if payloads == nil {
		return nil
	}

	strs := make([]string, len(payloads.Payloads))
	// Perform decoding one by one here so that we return individual errors
	for i, payload := range payloads.Payloads {
		strs[i] = rdc.ToString(payload)
	}
	return strs
}
