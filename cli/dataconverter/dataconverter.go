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

package dataconverter

import (
	"net/http"
	"strings"

	"go.temporal.io/sdk/converter"
)

var (
	dataConverter = converter.GetDefaultDataConverter()
)

func SetCurrent(dc converter.DataConverter) {
	dataConverter = dc
}

func SetRemoteEndpoint(endpoint string, namespace string, auth string) {
	endpoint = strings.ReplaceAll(endpoint, "{namespace}", namespace)

	dataConverter = converter.NewRemoteDataConverter(
		converter.GetDefaultDataConverter(),
		converter.RemoteDataConverterOptions{
			Endpoint: endpoint,
			ModifyRequest: func(req *http.Request) error {
				req.Header.Set("X-Namespace", namespace)
				if auth != "" {
					req.Header.Set("Authorization", auth)
				}

				return nil
			},
		},
	)
}

func GetCurrent() converter.DataConverter {
	return dataConverter
}
