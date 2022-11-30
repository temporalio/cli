package app_test

import (
	"github.com/golang/mock/gomock"
	"go.temporal.io/api/operatorservice/v1"
)

func (s *cliAppSuite) TestListSearchAttributes() {
	s.operatorClient.EXPECT().ListSearchAttributes(gomock.Any(), gomock.Any()).Return(&operatorservice.ListSearchAttributesResponse{}, nil)
	err := s.app.Run([]string{"", "operator", "search-attribute", "list"})
	s.Nil(err)

	s.operatorClient.EXPECT().ListSearchAttributes(gomock.Any(), gomock.Any()).Return(&operatorservice.ListSearchAttributesResponse{}, nil)
	err = s.app.Run([]string{"", "operator", "search-attribute", "list", "--namespace", cliTestNamespace})
	s.Nil(err)
}
