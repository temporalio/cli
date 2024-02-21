package temporalcli_test

import (
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
)

func (s *SharedServerSuite) TestOperator_SearchAttribute_Create_Already_Exists() {
	/*
		This test try to create:
		1, a system search attribute, that should fail.
		2, a custom search attribute with different type, that should fail.
		3, a custom search attribute with same type, that should pass.
	*/
	// Create System search attribute (WorkflowId)
	res := s.Execute(
		"operator", "search-attribute", "create",
		"--address", s.Address(),
		"--name", "WorkflowId",
		"--type", "Keyword",
	)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "unable to add search attributes: Search attribute WorkflowId is reserved by system.")

	// Create Custom search attribute with different type
	res2 := s.Execute(
		"operator", "search-attribute", "create",
		"--address", s.Address(),
		"--name", "CustomKeywordField",
		"--type", "Int",
	)
	s.Error(res2.Err)
	s.Contains(res2.Err.Error(), "search attribute CustomKeywordField already exists and has different type Keyword")

	// Create Custom search attribute with same type
	res3 := s.Execute(
		"operator", "search-attribute", "create",
		"--address", s.Address(),
		"--name", "CustomKeywordField",
		"--type", "Keyword",
	)
	s.NoError(res3.Err)
}

func (s *SharedServerSuite) TestOperator_SearchAttribute() {
	/*
		This test case first create 2 new custom search attributes, and verify the creation by list them.
		Then we delete one of the newly created SA, and verify the deletion by list again.
	*/
	res := s.Execute(
		"operator", "search-attribute", "create",
		"--address", s.Address(),
		"--name", "MySearchAttributeForTestCreateKeyword",
		"--type", "Keyword",
		"--name", "MySearchAttributeForTestCreateInt",
		"--type", "Int",
	)
	s.NoError(res.Err)

	// verify the creation succeed
	res2 := s.Execute(
		"operator", "search-attribute", "list",
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res2.Err)
	var jsonOut operatorservice.ListSearchAttributesResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res2.Stdout.Bytes(), &jsonOut, true))
	s.Equal(enums.INDEXED_VALUE_TYPE_KEYWORD, jsonOut.CustomAttributes["MySearchAttributeForTestCreateKeyword"])
	s.Equal(enums.INDEXED_VALUE_TYPE_INT, jsonOut.CustomAttributes["MySearchAttributeForTestCreateInt"])

	// Remove it
	res3 := s.Execute(
		"operator", "search-attribute", "remove",
		"--address", s.Address(),
		"--name", "MySearchAttributeForTestCreateKeyword",
		"--yes",
	)
	s.NoError(res3.Err)

	// verify deletion
	res4 := s.Execute(
		"operator", "search-attribute", "list",
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res2.Err)
	var jsonOut2 operatorservice.ListSearchAttributesResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res4.Stdout.Bytes(), &jsonOut2, true))
	// Int SA still exists
	s.Equal(enums.INDEXED_VALUE_TYPE_INT, jsonOut2.CustomAttributes["MySearchAttributeForTestCreateInt"])
	// Keyword SA no longer exists
	_, exists := jsonOut2.CustomAttributes["MySearchAttributeForTestCreateKeyword"]
	s.False(exists)

	// also verify some system attributes from list result
	s.Equal(enums.INDEXED_VALUE_TYPE_DATETIME, jsonOut.SystemAttributes["StartTime"])
	s.Equal(enums.INDEXED_VALUE_TYPE_KEYWORD, jsonOut.SystemAttributes["WorkflowId"])
}
