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

package headersprovider

import (
	"context"
)

type HeadersProvider interface {
	GetHeaders(context.Context) (map[string]string, error)
}

var (
	headersProvider HeadersProvider = nil
)

type authHeaderProvider struct {
	value string
}

func (a authHeaderProvider) GetHeaders(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		"Authorization": a.value,
	}, nil
}

func SetAuthorizationHeader(value string) {
	headersProvider = &authHeaderProvider{value: value}
}

func SetCurrent(hp HeadersProvider) {
	headersProvider = hp
}

func GetCurrent() HeadersProvider {
	return headersProvider
}
