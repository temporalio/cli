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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *cliAppSuite) TestDescribeBatchJob() {
	s.frontendClient.EXPECT().DescribeBatchOperation(gomock.Any(), gomock.Any()).Return(&workflowservice.DescribeBatchOperationResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "batch", "describe", "--job-id", "test"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListBatchJobs() {
	s.frontendClient.EXPECT().ListBatchOperations(gomock.Any(), gomock.Any()).Return(&workflowservice.ListBatchOperationsResponse{}, nil).Times(2)
	err := s.app.Run([]string{"", "batch", "list"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	err = s.app.Run([]string{"", "batch", "list", "--fields", "long"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestStopBatchJob() {
	s.frontendClient.EXPECT().StopBatchOperation(gomock.Any(), gomock.Any()).Return(&workflowservice.StopBatchOperationResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "batch", "terminate", "--job-id", "test", "--reason", "test"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestStartBatchJob_Signal() {
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{Count: 5}, nil).Once()
	s.frontendClient.EXPECT().StartBatchOperation(gomock.Any(), gomock.Any()).Return(&workflowservice.StartBatchOperationResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "workflow", "signal", "--name", "test-signal", "--query", "WorkflowType='test-type'", "--reason", "test-reason", "--input", "test-input", "--yes"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestStartBatchJob_Terminate() {
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{Count: 5}, nil).Once()
	s.frontendClient.EXPECT().StartBatchOperation(gomock.Any(), gomock.Any()).Return(&workflowservice.StartBatchOperationResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "workflow", "terminate", "--query", "WorkflowType='test-type'", "--reason", "test-reason", "--yes"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestStartBatchJob_Cancel() {
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{Count: 5}, nil).Once()
	s.frontendClient.EXPECT().StartBatchOperation(gomock.Any(), gomock.Any()).Return(&workflowservice.StartBatchOperationResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "workflow", "cancel", "--query", "WorkflowType='test-type'", "--reason", "test-reason", "--yes"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}
