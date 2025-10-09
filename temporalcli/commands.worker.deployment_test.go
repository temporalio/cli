package temporalcli_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type jsonVersionSummariesRowType struct {
	DeploymentName string    `json:"deploymentName"`
	BuildID        string    `json:"BuildID"`
	DrainageStatus string    `json:"drainageStatus"`
	CreateTime     time.Time `json:"createTime"`
}

type jsonRoutingConfigType struct {
	CurrentVersionDeploymentName        string    `json:"currentVersionDeploymentName"`
	CurrentVersionBuildID               string    `json:"currentVersionBuildID"`
	RampingVersionDeploymentName        string    `json:"rampingVersionDeploymentName"`
	RampingVersionBuildID               string    `json:"rampingVersionBuildID"`
	RampingVersionPercentage            float32   `json:"rampingVersionPercentage"`
	CurrentVersionChangedTime           time.Time `json:"currentVersionChangedTime"`
	RampingVersionChangedTime           time.Time `json:"rampingVersionChangedTime"`
	RampingVersionPercentageChangedTime time.Time `json:"rampingVersionPercentageChangedTime"`
}

type jsonDeploymentInfoType struct {
	Name                 string                        `json:"name"`
	CreateTime           time.Time                     `json:"createTime"`
	LastModifierIdentity string                        `json:"lastModifierIdentity"`
	RoutingConfig        jsonRoutingConfigType         `json:"routingConfig"`
	VersionSummaries     []jsonVersionSummariesRowType `json:"versionSummaries"`
	ManagerIdentity      string                        `json:"managerIdentity"`
}

type jsonDrainageInfo struct {
	DrainageStatus  string    `json:"drainageStatus"`
	LastChangedTime time.Time `json:"lastChangedTime"`
	LastCheckedTime time.Time `json:"lastCheckedTime"`
}

type jsonTaskQueueInfoRowType struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type jsonDeploymentVersionInfoType struct {
	Version            string                     `json:"version"`
	CreateTime         time.Time                  `json:"createTime"`
	RoutingChangedTime time.Time                  `json:"routingChangedTime"`
	CurrentSinceTime   time.Time                  `json:"currentSinceTime"`
	RampingSinceTime   time.Time                  `json:"rampingSinceTime"`
	RampPercentage     float32                    `json:"rampPercentage"`
	DrainageInfo       jsonDrainageInfo           `json:"drainageInfo"`
	TaskQueuesInfos    []jsonTaskQueueInfoRowType `json:"taskQueuesInfos"`
	Metadata           map[string]*common.Payload `json:"metadata"`
}

func (s *SharedServerSuite) TestDeployment_Set_Current_Version() {
	deploymentName := uuid.NewString()
	buildId := uuid.NewString()
	version := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId,
	}
	w := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w.Stop()

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "Name", deploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionDeploymentName", version.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionBuildID", version.BuildID)

	// json
	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(deploymentName, jsonOut.Name)
	s.Equal(version.DeploymentName, jsonOut.RoutingConfig.CurrentVersionDeploymentName)
	s.Equal(version.BuildID, jsonOut.RoutingConfig.CurrentVersionBuildID)

	// set metadata
	res = s.Execute(
		"worker", "deployment", "update-metadata-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--metadata", "bar=1",
		"--output", "json",
	)
	s.NoError(res.Err)
	var jsonVersionOut jsonDeploymentVersionInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonVersionOut))

	// "1" is "MQ=="
	s.Equal("MQ==", base64.StdEncoding.EncodeToString(jsonVersionOut.Metadata["bar"].GetData()))
	// "json/plain" is "anNvbi9wbGFpbg=="
	s.Equal("anNvbi9wbGFpbg==", base64.StdEncoding.EncodeToString(jsonVersionOut.Metadata["bar"].GetMetadata()["encoding"]))

	// remove metadata
	res = s.Execute(
		"worker", "deployment", "update-metadata-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--remove-entries", "bar",
		"--output", "json",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--output", "json",
	)
	s.NoError(res.Err)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonVersionOut))
	s.Nil(jsonVersionOut.Metadata)
}

func (s *SharedServerSuite) TestDeployment_Set_Current_Version_AllowNoPollers() {
	deploymentName := uuid.NewString()
	buildId := uuid.NewString()
	version := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId,
	}

	// with --allow-no-pollers, no need to have a worker polling on this version
	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--allow-no-pollers",
		"--yes",
	)
	s.NoError(res.Err)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "Name", deploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionDeploymentName", version.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionBuildID", version.BuildID)

	// json
	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(deploymentName, jsonOut.Name)
	s.Equal(version.DeploymentName, jsonOut.RoutingConfig.CurrentVersionDeploymentName)
	s.Equal(version.BuildID, jsonOut.RoutingConfig.CurrentVersionBuildID)
}

func (s *SharedServerSuite) TestDeployment_Set_Ramping_Version_AllowNoPollers() {
	deploymentName := uuid.NewString()
	buildId := uuid.NewString()
	version := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId,
	}

	// with --allow-no-pollers, no need to have a worker polling on this version
	res := s.Execute(
		"worker", "deployment", "set-ramping-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--percentage", "5",
		"--allow-no-pollers",
		"--yes",
	)
	s.NoError(res.Err)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "Name", deploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionDeploymentName", version.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionBuildID", version.BuildID)

	// json
	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(deploymentName, jsonOut.Name)
	s.Equal(version.DeploymentName, jsonOut.RoutingConfig.RampingVersionDeploymentName)
	s.Equal(version.BuildID, jsonOut.RoutingConfig.RampingVersionBuildID)
}

func filterByNamePrefix(jsonOut []jsonDeploymentInfoType, prefix string) []jsonDeploymentInfoType {
	result := []jsonDeploymentInfoType{}
	for i := range jsonOut {
		if strings.HasPrefix(jsonOut[i].Name, prefix) {
			result = append(result, jsonOut[i])
		}
	}
	return result
}

func (s *SharedServerSuite) TestDeployment_List() {
	prefix := "deployment_list_"
	deploymentName1 := prefix + "a_" + uuid.NewString()
	deploymentName2 := prefix + "b_" + uuid.NewString()
	buildId1 := uuid.NewString()
	buildId2 := uuid.NewString()
	version1 := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName1,
		BuildID:        buildId1,
	}
	version2 := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName2,
		BuildID:        buildId2,
	}

	w1 := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version1,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w1.Stop()

	w2 := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version2,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w2.Stop()

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName1)
		assert.Contains(t, res.Stdout.String(), deploymentName2)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		)
		assert.NoError(t, res.Err)
		res = s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version2.DeploymentName, "--build-id", version2.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version2.DeploymentName, "--build-id", version2.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res = s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		s.NoError(res.Err)
	}, 10*time.Second, 100*time.Millisecond)

	s.ContainsOnSameLine(res.Stdout.String(), deploymentName1, version1.BuildID)
	s.ContainsOnSameLine(res.Stdout.String(), deploymentName2, version2.BuildID)

	// json
	res = s.Execute(
		"worker", "deployment", "list",
		"--address", s.Address(),
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut []jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	jsonOut = filterByNamePrefix(jsonOut, prefix)
	sort.Slice(jsonOut, func(i, j int) bool {
		return jsonOut[i].Name < jsonOut[j].Name
	})
	s.Equal(2, len(jsonOut))
	s.Equal(deploymentName1, jsonOut[0].Name)
	s.Equal(version1.DeploymentName, jsonOut[0].RoutingConfig.CurrentVersionDeploymentName)
	s.Equal(version1.BuildID, jsonOut[0].RoutingConfig.CurrentVersionBuildID)
	s.Equal(deploymentName2, jsonOut[1].Name)
	s.Equal(version2.DeploymentName, jsonOut[1].RoutingConfig.CurrentVersionDeploymentName)
	s.Equal(version2.BuildID, jsonOut[1].RoutingConfig.CurrentVersionBuildID)
}

func (s *SharedServerSuite) TestDeployment_Describe_Drainage() {
	deploymentName := uuid.NewString()
	buildId1 := "a" + uuid.NewString()
	buildId2 := "b" + uuid.NewString()
	version1 := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId1,
	}
	version2 := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId2,
	}

	w1 := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version1,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w1.Stop()

	w2 := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version2,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w2.Stop()

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		)
		assert.NoError(t, res.Err)
		res = s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version2.DeploymentName, "--build-id", version2.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionDeploymentName", version1.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionBuildID", version1.BuildID)

	fmt.Print("hello")
	res = s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version2.DeploymentName, "--build-id", version2.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionDeploymentName", version2.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionBuildID", version2.BuildID)
	s.ContainsOnSameLine(res.Stdout.String(), version1.DeploymentName, "draining")
	s.ContainsOnSameLine(res.Stdout.String(), version2.DeploymentName, "unspecified")

	// json
	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(deploymentName, jsonOut.Name)
	sort.Slice(jsonOut.VersionSummaries, func(i, j int) bool {
		return jsonOut.VersionSummaries[i].BuildID < jsonOut.VersionSummaries[j].BuildID
	})

	s.Equal(2, len(jsonOut.VersionSummaries))
	s.Equal("draining", jsonOut.VersionSummaries[0].DrainageStatus)
	s.Equal(version1.BuildID, jsonOut.VersionSummaries[0].BuildID)
	s.Equal("unspecified", jsonOut.VersionSummaries[1].DrainageStatus)
	s.Equal(version2.BuildID, jsonOut.VersionSummaries[1].BuildID)
}

func (s *SharedServerSuite) TestDeployment_Ramping() {
	deploymentName := uuid.NewString()
	buildId1 := "a" + uuid.NewString()
	buildId2 := "b" + uuid.NewString()
	version1 := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId1,
	}
	version2 := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId2,
	}

	w1 := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version1,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w1.Stop()

	w2 := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version2,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w2.Stop()

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		)
		assert.NoError(t, res.Err)
		res = s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version2.DeploymentName, "--build-id", version2.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "set-ramping-version",
		"--address", s.Address(),
		"--deployment-name", version2.DeploymentName, "--build-id", version2.BuildID,
		"--percentage", "12.5",
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionDeploymentName", version1.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionBuildID", version1.BuildID)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionDeploymentName", version2.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionBuildID", version2.BuildID)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionPercentage", "12.5")

	// setting version2 as current also removes the ramp
	res = s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version2.DeploymentName, "--build-id", version2.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(deploymentName, jsonOut.Name)
	s.Empty(jsonOut.RoutingConfig.RampingVersionBuildID)
	s.Equal(version2.BuildID, jsonOut.RoutingConfig.CurrentVersionBuildID)

	//same with explicit delete
	res = s.Execute(
		"worker", "deployment", "set-ramping-version",
		"--address", s.Address(),
		"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		"--percentage", "10.1",
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionDeploymentName", version2.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "CurrentVersionBuildID", version2.BuildID)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionDeploymentName", version1.DeploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionBuildID", version1.BuildID)
	s.ContainsOnSameLine(res.Stdout.String(), "RampingVersionPercentage", "10.1")

	res = s.Execute(
		"worker", "deployment", "set-ramping-version",
		"--address", s.Address(),
		"--deployment-name", version1.DeploymentName, "--build-id", version1.BuildID,
		"--delete",
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
		"--output", "json",
	)
	s.NoError(res.Err)

	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(deploymentName, jsonOut.Name)
	s.Equal(float32(0), jsonOut.RoutingConfig.RampingVersionPercentage)
	s.Equal(version2.BuildID, jsonOut.RoutingConfig.CurrentVersionBuildID)
}

func (s *SharedServerSuite) TestDeployment_Set_Manager_Identity() {
	deploymentName := uuid.NewString()
	BuildID := uuid.NewString()
	testIdentity := uuid.NewString()
	version := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        BuildID,
	}
	w := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
	})
	defer w.Stop()

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res := s.Execute(
		"worker", "deployment", "manager-identity", "set",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--manager-identity", testIdentity,
		"--yes",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "Name", deploymentName)
	s.ContainsOnSameLine(res.Stdout.String(), "ManagerIdentity", testIdentity)

	// json
	res = s.Execute(
		"worker", "deployment", "describe",
		"--address", s.Address(),
		"--name", deploymentName,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut jsonDeploymentInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(deploymentName, jsonOut.Name)
	s.Equal(testIdentity, jsonOut.ManagerIdentity)
}
