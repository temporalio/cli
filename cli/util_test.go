// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
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

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	enumspb "go.temporal.io/api/enums/v1"
)

func (s *utilSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(utilSuite))
}

type utilSuite struct {
	*require.Assertions
	suite.Suite
}

func (s *utilSuite) TestStringToEnum_MapCaseInsensitive() {
	enumValues := map[string]int32{
		"Unspecified": 0,
		"Transfer":    1,
		"Timer":       2,
		"Replication": 3,
	}

	result, err := stringToEnum("timeR", enumValues)
	s.NoError(err)
	s.Equal(result, int32(2)) // Timer
}

func (s *utilSuite) TestStringToEnum_MapNonExisting() {
	enumValues := map[string]int32{
		"Unspecified": 0,
		"Transfer":    1,
		"Timer":       2,
		"Replication": 3,
	}

	result, err := stringToEnum("Timer2", enumValues)
	s.Error(err)
	s.Equal(result, int32(0))
}

func (s *utilSuite) TestStringToEnum_MapEmptyValue() {
	enumValues := map[string]int32{
		"Unspecified": 0,
		"Transfer":    1,
		"Timer":       2,
		"Replication": 3,
	}

	result, err := stringToEnum("", enumValues)
	s.NoError(err)
	s.Equal(result, int32(0))
}

func (s *utilSuite) TestStringToEnum_MapEmptyEnum() {
	enumValues := map[string]int32{}

	result, err := stringToEnum("Timer", enumValues)
	s.Error(err)
	s.Equal(result, int32(0))
}

func (s *utilSuite) TestParseFoldStatusList() {
	tests := map[string]struct {
		value   string
		want    []enumspb.WorkflowExecutionStatus
		wantErr bool
	}{
		"default values": {
			value: "completed,canceled,terminated",
			want: []enumspb.WorkflowExecutionStatus{
				enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED,
				enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED,
			},
		},
		"no values": {
			value: "",
			want:  nil,
		},
		"invalid": {
			value:   "Foobar",
			wantErr: true,
		},
		"title case": {
			value: "Running,Completed,Failed,Canceled,Terminated,ContinuedAsNew,TimedOut",
			want: []enumspb.WorkflowExecutionStatus{
				enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
				enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				enumspb.WORKFLOW_EXECUTION_STATUS_FAILED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED,
				enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
				enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
			},
		},
		"upper case": {
			value: "RUNNING,COMPLETED,FAILED,CANCELED,TERMINATED,CONTINUEDASNEW,TIMEDOUT",
			want: []enumspb.WorkflowExecutionStatus{
				enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
				enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				enumspb.WORKFLOW_EXECUTION_STATUS_FAILED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED,
				enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
				enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
			},
		},
	}
	for name, tt := range tests {
		s.Run(name, func() {
			got, err := parseFoldStatusList(tt.value)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.Equal(tt.want, got)
			}
		})
	}
}

func (s *utilSuite) TestParseKeyValuePairs() {
	tests := map[string]struct {
		input   []string
		want    map[string]string
		wantErr bool
	}{
		"simple values": {
			input: []string{
				"key1=value1",
				"key2=value2",
				"key3=value3=with=equal",
				"key4=value4:with-symbols",
				"key5=",
				`key6={"Auth":{"Enabled":false,"Options":["audience","organization"]},"ShowTemporalSystemNamespace":true}`,
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3=with=equal",
				"key4": "value4:with-symbols",
				"key5": "",
				"key6": `{"Auth":{"Enabled":false,"Options":["audience","organization"]},"ShowTemporalSystemNamespace":true}`,
			},
		},
		"no values": {
			input: []string{},
			want:  map[string]string{},
		},
		"empty": {
			input:   []string{""},
			wantErr: true,
		},
		"no separator": {
			input:   []string{"key:value"},
			wantErr: true,
		},
		"no key": {
			input:   []string{"=value"},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		s.Run(name, func() {
			got, err := SplitKeyValuePairs(tt.input)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.Equal(tt.want, got)
			}
		})
	}
}
