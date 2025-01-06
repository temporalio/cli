package temporalcli_test

import (
	"encoding/base64"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/api/common/v1"
)

type jsonTaskQueuesInfosRowType struct {
		Name string `json:"name"`
		Type string `json:"type"`
		FirstPollerTime time.Time `json:"firstPollerTime"`
}

type jsonDeploymentType struct {
	SeriesName string `json:"seriesName"`
	BuildID string `json:"buildId"`
}

type jsonDeploymentInfoType struct {
	Deployment jsonDeploymentType `json:"deployment"`
	CreateTime time.Time `json:"createTime"`
	IsCurrent bool `json:"isCurrent"`
	TaskQueuesInfos []jsonTaskQueuesInfosRowType `json:"taskQueuesInfos,omitempty"`
	Metadata map[string]*common.Payload `json:"metadata,omitempty"`
}

type jsonDeploymentReachabilityInfoType struct {
	DeploymentInfo jsonDeploymentInfoType `json:"deploymentInfo"`
	Reachability string `json:"reachability"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

type jsonDeploymentListEntryType struct {
	Deployment jsonDeploymentType `json:"deployment"`
	CreateTime time.Time `json:"createTime"`
	IsCurrent bool `json:"isCurrent"`
}

func (s *SharedServerSuite) TestDeployment_Update_Current() {
	seriesName := uuid.NewString()
	buildId := uuid.NewString()

	res := s.Execute(
		"deployment", "update-current",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--deployment-build-id", buildId,
		"--deployment-metadata", "bar=1",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"deployment", "get-current",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "SeriesName", seriesName)
	s.ContainsOnSameLine(res.Stdout.String(), "BuildID", buildId)
	s.ContainsOnSameLine(res.Stdout.String(), "IsCurrent", "true")
	s.ContainsOnSameLine(res.Stdout.String(), "Metadata", "data:\"1\"")

	// json
	res = s.Execute(
		"deployment", "get-current",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(jsonDeploymentType{SeriesName: seriesName, BuildID: buildId}, jsonOut.Deployment)
	s.True(jsonOut.IsCurrent)
	// "1" is "MQ=="
	s.Equal("MQ==", base64.StdEncoding.EncodeToString(jsonOut.Metadata["bar"].GetData()))
	// "json/plain" is "anNvbi9wbGFpbg=="
	s.Equal("anNvbi9wbGFpbg==", base64.StdEncoding.EncodeToString(jsonOut.Metadata["bar"].GetMetadata()["encoding"]))
}

func (s *SharedServerSuite) TestDeployment_List() {
	seriesName := uuid.NewString()
	buildId1 := uuid.NewString()
	buildId2 := uuid.NewString()

	res := s.Execute(
		"deployment", "update-current",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--deployment-build-id", buildId1,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"deployment", "update-current",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--deployment-build-id", buildId2,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"deployment", "list",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), seriesName, buildId1, "now", "false")
	s.ContainsOnSameLine(res.Stdout.String(), seriesName, buildId2, "now", "true")

	// json
	res = s.Execute(
		"deployment", "list",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut []jsonDeploymentListEntryType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	sort.Slice(jsonOut, func(i, j int) bool {
		return jsonOut[i].CreateTime.Before(jsonOut[j].CreateTime)
	})
	s.Equal(len(jsonOut), 2)
	s.Equal(jsonDeploymentType{SeriesName: seriesName, BuildID: buildId1}, jsonOut[0].Deployment)
	s.True(!jsonOut[0].IsCurrent)
	s.Equal(jsonDeploymentType{SeriesName: seriesName, BuildID: buildId2}, jsonOut[1].Deployment)
	s.True(jsonOut[1].IsCurrent)
}

func (s *SharedServerSuite) TestDeployment_Describe_Reachability() {
	seriesName := uuid.NewString()
	buildId1 := uuid.NewString()
	buildId2 := uuid.NewString()

	res := s.Execute(
		"deployment", "update-current",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--deployment-build-id", buildId1,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"deployment", "update-current",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--deployment-build-id", buildId2,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"deployment", "describe",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--deployment-build-id", buildId1,
		"--report-reachability",
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "SeriesName", seriesName)
	s.ContainsOnSameLine(res.Stdout.String(), "BuildID", buildId1)
	s.ContainsOnSameLine(res.Stdout.String(), "IsCurrent", "false")
	s.ContainsOnSameLine(res.Stdout.String(), "Reachability", "unreachable")

	// json
	res = s.Execute(
		"deployment", "describe",
		"--address", s.Address(),
		"--deployment-series-name", seriesName,
		"--deployment-build-id", buildId2,
		"--report-reachability",
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentReachabilityInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(jsonDeploymentType{SeriesName: seriesName, BuildID: buildId2}, jsonOut.DeploymentInfo.Deployment)
	s.True(jsonOut.DeploymentInfo.IsCurrent)
	s.Equal(jsonOut.Reachability, "reachable")
}
