package app_test

import (
	"github.com/golang/mock/gomock"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *cliAppSuite) TestDescribeCluster() {
	s.frontendClient.EXPECT().GetClusterInfo(gomock.Any(), gomock.Any()).Return(&workflowservice.GetClusterInfoResponse{}, nil).Times(2)
	err := s.app.Run([]string{"", "operator", "cluster", "describe"})
	s.NoError(err)

	err = s.app.Run([]string{"", "operator", "cluster", "describe", "--fields", "long", "--output", "table"})
	s.NoError(err)
}

func (s *cliAppSuite) TestDescribeSystem() {
	s.frontendClient.EXPECT().GetSystemInfo(gomock.Any(), gomock.Any()).Return(&workflowservice.GetSystemInfoResponse{
		Capabilities: &workflowservice.GetSystemInfoResponse_Capabilities{},
	}, nil).Times(2)
	err := s.app.Run([]string{"", "operator", "cluster", "system"})
	s.NoError(err)

	err = s.app.Run([]string{"", "operator", "cluster", "system", "--fields", "long", "--output", "table"})
	s.NoError(err)
}

func (s *cliAppSuite) TestUpsertCluster() {
	s.operatorClient.EXPECT().AddOrUpdateRemoteCluster(gomock.Any(), gomock.Any()).Return(&operatorservice.AddOrUpdateRemoteClusterResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "operator", "cluster", "upsert", "--frontend-address", "localhost:7233", "--enable-connection", "true"})
	s.NoError(err)
}

func (s *cliAppSuite) TestListCluster() {
	s.operatorClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(&operatorservice.ListClustersResponse{}, nil).Times(2)
	err := s.app.Run([]string{"", "operator", "cluster", "list"})
	s.NoError(err)

	err = s.app.Run([]string{"", "operator", "cluster", "list", "--fields", "long", "--output", "table"})
	s.NoError(err)
}

func (s *cliAppSuite) TestRemoveCluster() {
	s.operatorClient.EXPECT().RemoveRemoteCluster(gomock.Any(), gomock.Any()).Return(&operatorservice.RemoveRemoteClusterResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "operator", "cluster", "remove", "--name", "test"})
	s.NoError(err)
}
