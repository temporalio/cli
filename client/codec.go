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
