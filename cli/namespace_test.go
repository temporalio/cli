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
	"time"

	"github.com/golang/mock/gomock"
	namespacepb "go.temporal.io/api/namespace/v1"
	"go.temporal.io/api/operatorservice/v1"
	replicationpb "go.temporal.io/api/replication/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/primitives/timestamp"
)

func (s *cliAppSuite) TestNamespaceRegister_LocalNamespace() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := s.app.Run([]string{"", "namespace", "register", "--global", "false", cliTestNamespace})
	s.NoError(err)
}

func (s *cliAppSuite) TestNamespaceRegister_GlobalNamespace() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := s.app.Run([]string{"", "namespace", "register", "--global", "true", cliTestNamespace})
	s.NoError(err)
}

func (s *cliAppSuite) TestNamespaceRegister_Data() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := s.app.Run([]string{"", "namespace", "register", "--data", "k1=v1", "--data", "k2=v2", "true", cliTestNamespace})
	s.NoError(err)
}

func (s *cliAppSuite) TestNamespaceRegister_NamespaceExist() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewNamespaceAlreadyExists(""))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "register", "--global", "true", cliTestNamespace})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceRegister_Cluster() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := s.app.Run([]string{"", "namespace", "register", "--cluster", "active", "--cluster", "standby", cliTestNamespace})
	s.NoError(err)
}

func (s *cliAppSuite) TestNamespaceRegister_Failed() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "register", "--global", "true", cliTestNamespace})
	s.Equal(1, errorCode)
}

var describeNamespaceResponseServer = &workflowservice.DescribeNamespaceResponse{
	NamespaceInfo: &namespacepb.NamespaceInfo{
		Name:        "test-namespace",
		Description: "a test namespace",
		OwnerEmail:  "test@uber.com",
	},
	Config: &namespacepb.NamespaceConfig{
		WorkflowExecutionRetentionTtl: timestamp.DurationPtr(3 * time.Hour * 24),
	},
	ReplicationConfig: &replicationpb.NamespaceReplicationConfig{
		ActiveClusterName: "active",
		Clusters: []*replicationpb.ClusterReplicationConfig{
			{
				ClusterName: "active",
			},
			{
				ClusterName: "standby",
			},
		},
	},
}

func (s *cliAppSuite) TestNamespaceUpdate() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, nil).Times(2)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
	err := s.app.Run([]string{"", "namespace", "update", cliTestNamespace})
	s.Nil(err)
	err = s.app.Run([]string{"", "namespace", "update", "--description", "another desc", "--email", "another@uber.com", "--retention", "1", cliTestNamespace})
	s.Nil(err)
}

func (s *cliAppSuite) TestNamespaceUpdate_Data() {
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(describeNamespaceResponseServer, nil).Times(1)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	err := s.app.Run([]string{"", "namespace", "update", "--data", "k1=v1", "--data", "k2=v2", "true", cliTestNamespace})
	s.NoError(err)
}

func (s *cliAppSuite) TestNamespaceUpdate_NamespaceNotExist() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, nil)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewNamespaceNotFound("missing-namespace"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "update", cliTestNamespace})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceUpdate_ActiveClusterFlagNotSet_NamespaceNotExist() {
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewNamespaceNotFound("missing-namespace"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "update", cliTestNamespace})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceUpdate_Cluster() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := s.app.Run([]string{"", "namespace", "update", "--cluster", "active", "--cluster", "standby", cliTestNamespace})
	s.NoError(err)
}

func (s *cliAppSuite) TestNamespaceUpdate_Failed() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, nil)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "update", cliTestNamespace})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceDescribe() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), &workflowservice.DescribeNamespaceRequest{Namespace: cliTestNamespace, Id: ""}).Return(resp, nil)
	err := s.app.Run([]string{"", "namespace", "describe", cliTestNamespace})
	s.Nil(err)
}

func (s *cliAppSuite) TestNamespaceDescribe_ById() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), &workflowservice.DescribeNamespaceRequest{Namespace: "", Id: "nid"}).Return(resp, nil)
	err := s.app.Run([]string{"", "namespace", "describe", "--namespace-id", "nid"})
	s.Nil(err)
}

func (s *cliAppSuite) TestNamespaceDescribe_NamespaceNotExist() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, serviceerror.NewNamespaceNotFound("missing-namespace"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "describe", cliTestNamespace})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceDescribe_Failed() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "describe", cliTestNamespace})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceDelete() {
	s.operatorClient.EXPECT().DeleteNamespace(gomock.Any(), &operatorservice.DeleteNamespaceRequest{Namespace: cliTestNamespace}).Return(&operatorservice.DeleteNamespaceResponse{}, nil)
	err := s.app.Run([]string{"", "namespace", "delete", "--yes", cliTestNamespace})
	s.Nil(err)
}

func (s *cliAppSuite) TestNamespaceDelete_NamespaceNotExist() {
	s.operatorClient.EXPECT().DeleteNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewNamespaceNotFound("missing-namespace"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "delete", "--yes", cliTestNamespace})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceDelete_Failed() {
	s.operatorClient.EXPECT().DeleteNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.RunWithExitCode([]string{"", "namespace", "delete", "--yes", cliTestNamespace})
	s.Equal(1, errorCode)
}
