// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
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
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *cliAppSuite) TestDescribeCluster() {
	s.frontendClient.EXPECT().GetClusterInfo(gomock.Any(), gomock.Any()).Return(&workflowservice.GetClusterInfoResponse{}, nil).Times(2)
	err := s.app.Run([]string{"", "cluster", "describe"})
	s.NoError(err)

	err = s.app.Run([]string{"", "cluster", "describe", "--fields", "long", "--output", "table"})
	s.NoError(err)
}

func (s *cliAppSuite) TestDescribeSystem() {
	s.frontendClient.EXPECT().GetSystemInfo(gomock.Any(), gomock.Any()).Return(&workflowservice.GetSystemInfoResponse{
		Capabilities: &workflowservice.GetSystemInfoResponse_Capabilities{},
	}, nil).Times(2)
	err := s.app.Run([]string{"", "cluster", "system"})
	s.NoError(err)

	err = s.app.Run([]string{"", "cluster", "system", "--fields", "long", "--output", "table"})
	s.NoError(err)
}

func (s *cliAppSuite) TestUpsertCluster() {
	s.operatorClient.EXPECT().AddOrUpdateRemoteCluster(gomock.Any(), gomock.Any()).Return(&operatorservice.AddOrUpdateRemoteClusterResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "cluster", "upsert", "--frontend-address", "localhost:7233", "--enable-connection", "true"})
	s.NoError(err)
}

func (s *cliAppSuite) TestListCluster() {
	s.operatorClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(&operatorservice.ListClustersResponse{}, nil).Times(2)
	err := s.app.Run([]string{"", "cluster", "list"})
	s.NoError(err)

	err = s.app.Run([]string{"", "cluster", "list", "--fields", "long", "--output", "table"})
	s.NoError(err)
}

func (s *cliAppSuite) TestRemoveCluster() {
	s.operatorClient.EXPECT().RemoveRemoteCluster(gomock.Any(), gomock.Any()).Return(&operatorservice.RemoveRemoteClusterResponse{}, nil).Times(1)
	err := s.app.Run([]string{"", "cluster", "remove", "--name", "test"})
	s.NoError(err)
}
