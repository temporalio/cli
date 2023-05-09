package app_test

import (
	"github.com/golang/mock/gomock"
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
