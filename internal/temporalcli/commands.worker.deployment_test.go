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
	deploymentpb "go.temporal.io/api/deployment/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
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

type jsonVersionStatsType struct {
	ApproximateBacklogCount int64   `json:"approximateBacklogCount"`
	ApproximateBacklogAge   int64   `json:"approximateBacklogAge"` // Duration is serialized as nanoseconds
	BacklogIncreaseRate     float32 `json:"backlogIncreaseRate"`
	TasksAddRate            float32 `json:"tasksAddRate"`
	TasksDispatchRate       float32 `json:"tasksDispatchRate"`
}

type jsonTaskQueueInfoRowType struct {
	Name               string                          `json:"name"`
	Type               string                          `json:"type"`
	Stats              *jsonVersionStatsType           `json:"stats,omitempty"`
	StatsByPriorityKey map[string]jsonVersionStatsType `json:"statsByPriorityKey,omitempty"`
}

type jsonComputeConfigScalingGroupSummary struct {
	TaskQueueTypes []string `json:"taskQueueTypes,omitempty"`
	ProviderType   string   `json:"providerType"`
}

type jsonComputeConfig struct {
	ScalingGroups []jsonComputeConfigScalingGroupSummary `json:"scalingGroups,omitempty"`
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
	ComputeConfig      *jsonComputeConfig         `json:"computeConfig,omitempty"`
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
		"worker", "deployment", "update-version-metadata",
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
		"worker", "deployment", "update-version-metadata",
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

	// --build-id and --unversioned are either/or. there should be an error if
	// the user specifies both flags.
	res = s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--unversioned",
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "specify either --build-id or --unversioned")
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
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), version1.BuildID)
		assert.Contains(t, res.Stdout.String(), version2.BuildID)
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

func (s *SharedServerSuite) TestDeployment_Describe_Version_TaskQueueStats_WithPriority() {
	s.testDeploymentDescribeVersionTaskQueueStats(true)
}

func (s *SharedServerSuite) TestDeployment_Describe_Version_TaskQueueStats_WithoutPriority() {
	s.testDeploymentDescribeVersionTaskQueueStats(false)
}

func (s *SharedServerSuite) testDeploymentDescribeVersionTaskQueueStats(withPriority bool) {
	deploymentName := uuid.NewString()
	buildId := uuid.NewString()
	taskQueue := uuid.NewString()
	version := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        buildId,
	}

	// Create worker directly with explicit versioning behavior
	w1 := worker.New(s.Client, taskQueue, worker.Options{
		DeploymentOptions: worker.DeploymentOptions{
			UseVersioning: true,
			Version:       version,
		},
	})

	// Register a workflow with explicit Pinned versioning behavior
	w1.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input any) (any, error) {
			workflow.GetSignalChannel(ctx, "complete-signal").Receive(ctx, nil)
			return nil, nil
		},
		workflow.RegisterOptions{
			Name:               "TestBacklogWorkflow",
			VersioningBehavior: workflow.VersioningBehaviorPinned,
		},
	)

	s.NoError(w1.Start())

	// Wait for the deployment version to appear
	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	// Set as current version so workflows are routed to this version
	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--yes",
	)
	s.NoError(res.Err)

	// Stop the worker so workflow tasks will backlog
	w1.Stop()

	// Start workflows - they will queue up as workflow tasks since there is no worker
	// Priority keys: 1 (high), 3 (medium/default), 5 (low)
	priorityKeys := []int{1, 3, 5}
	numWorkflows := len(priorityKeys)
	workflowRuns := make([]client.WorkflowRun, numWorkflows)
	for i := 0; i < numWorkflows; i++ {
		opts := client.StartWorkflowOptions{
			TaskQueue: taskQueue,
		}
		if withPriority {
			opts.Priority = temporal.Priority{PriorityKey: priorityKeys[i]}
		}
		run, err := s.Client.ExecuteWorkflow(
			s.Context,
			opts,
			"TestBacklogWorkflow",
			"test-input",
		)
		s.NoError(err)
		workflowRuns[i] = run
		s.T().Logf("Started workflow %d: %s", i, run.GetID())
	}

	// Test 1: Verify gRPC returns non-zero backlog stats (validates server behavior)
	// Use raw gRPC instead of SDK's DeploymentClient to avoid circular dependency
	s.EventuallyWithT(func(t *assert.CollectT) {
		desc, err := s.Client.WorkflowService().DescribeWorkerDeploymentVersion(s.Context, &workflowservice.DescribeWorkerDeploymentVersionRequest{
			Namespace: s.Namespace(),
			DeploymentVersion: &deploymentpb.WorkerDeploymentVersion{
				DeploymentName: version.DeploymentName,
				BuildId:        version.BuildID,
			},
			ReportTaskQueueStats: true,
		})
		assert.NoError(t, err)

		// Find the workflow task queue and check backlog stats (matching SDK test pattern)
		for _, tqInfo := range desc.GetVersionTaskQueues() {
			if tqInfo.GetName() == taskQueue && tqInfo.GetType() == enumspb.TASK_QUEUE_TYPE_WORKFLOW && tqInfo.GetStats() != nil {
				stats := tqInfo.GetStats()
				backlogIncreaseRate := stats.TasksAddRate - stats.TasksDispatchRate
				// Check all stats fields like the SDK test does
				if stats.ApproximateBacklogCount > 0 &&
					stats.ApproximateBacklogAge.AsDuration().Nanoseconds() > 0 &&
					stats.TasksAddRate > 0 &&
					stats.TasksDispatchRate == 0 && // zero task dispatch due to no pollers
					backlogIncreaseRate > 0 {
					return // Success
				}
				t.Errorf("Unexpected backlog stats for tq %s: backlogCount=%d, backlogAge=%v, addRate=%f, dispatchRate=%f, increaseRate=%f",
					taskQueue, stats.ApproximateBacklogCount, stats.ApproximateBacklogAge.AsDuration(),
					stats.TasksAddRate, stats.TasksDispatchRate, backlogIncreaseRate)
				return
			}
		}
		t.Errorf("No workflow task queue with stats found for task queue %s in %d task queues", taskQueue, len(desc.GetVersionTaskQueues()))
	}, 10*time.Second, 500*time.Millisecond)

	// Verify text output has all stats columns with non-zero values
	res = s.Execute(
		"worker", "deployment", "describe-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--report-task-queue-stats",
	)
	s.NoError(res.Err)
	outputWithStats := res.Stdout.String()
	// Verify column headers are present
	s.Contains(outputWithStats, "ApproximateBacklogCount")
	s.Contains(outputWithStats, "ApproximateBacklogAge")
	s.Contains(outputWithStats, "BacklogIncreaseRate")
	s.Contains(outputWithStats, "TasksAddRate")
	s.Contains(outputWithStats, "TasksDispatchRate")
	// Verify the workflow task queue row has the expected backlog count (we started 3 workflows)
	// Format: Name Type ApproximateBacklogCount ApproximateBacklogAge BacklogIncreaseRate TasksAddRate TasksDispatchRate
	s.ContainsOnSameLine(outputWithStats, taskQueue, "workflow", "3")
	if withPriority {
		s.Contains(outputWithStats, "Stats by Priority")
		// Verify each priority key row shows approximate backlog count of 1 (one workflow per priority)
		// The format is: Priority ApproximateBacklogCount ApproximateBacklogAge ...
		// We started one workflow with each priority key (1, 3, 5), so each should have backlog of 1
		for _, priorityKey := range priorityKeys {
			// Check that the priority row contains the priority key followed by backlog count of 1
			nonZeroBacklogIncreaseRate := "."
			nonZeroTasksAddRate := "."
			zeroTaskDispatchRate := "0"
			// Once priority is enabled in all servers by default, check that backlog count of 1 is on each priority row.
			// For servers where priority is not enabled, or workflows that aren't actively using priority keys,
			// all backlog will be sent to the default priority key (3).
			s.ContainsOnSameLine(outputWithStats,
				fmt.Sprintf("%d", priorityKey),
				nonZeroBacklogIncreaseRate,
				nonZeroTasksAddRate,
				zeroTaskDispatchRate,
			)
		}
	} else {
		// Verify that "Stats by Priority" is NOT shown when only the default priority key (3) is used
		s.NotContains(outputWithStats, "Stats by Priority")
	}

	// Test 2: describe-version WITHOUT --report-task-queue-stats should NOT show stats columns
	res = s.Execute(
		"worker", "deployment", "describe-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
	)
	s.NoError(res.Err)
	// Stats columns should not appear in text output
	outputNoStats := res.Stdout.String()
	s.NotContains(outputNoStats, "ApproximateBacklogCount")
	s.NotContains(outputNoStats, "ApproximateBacklogAge")
	s.NotContains(outputNoStats, "BacklogIncreaseRate")
	s.NotContains(outputNoStats, "TasksAddRate")
	s.NotContains(outputNoStats, "TasksDispatchRate")

	// Test 3: JSON output without stats flag - stats should be nil/missing
	res = s.Execute(
		"worker", "deployment", "describe-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--output", "json",
	)
	s.NoError(res.Err)
	var jsonOutNoStats jsonDeploymentVersionInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOutNoStats))
	// Should have task queue info but no stats
	s.Greater(len(jsonOutNoStats.TaskQueuesInfos), 0)
	for _, tq := range jsonOutNoStats.TaskQueuesInfos {
		s.Nil(tq.Stats, "Stats should be nil when --report-task-queue-stats is not provided")
	}

	// Test 4: JSON output with stats flag - verify stats structure is present with non-zero backlog
	res = s.Execute(
		"worker", "deployment", "describe-version",
		"--address", s.Address(),
		"--deployment-name", version.DeploymentName, "--build-id", version.BuildID,
		"--report-task-queue-stats",
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOutWithStats jsonDeploymentVersionInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOutWithStats))
	s.Greater(len(jsonOutWithStats.TaskQueuesInfos), 0, "Should have task queue info")

	// Find the workflow task queue and verify it has non-zero backlog stats
	foundWorkflowTQWithStats := false
	for _, tq := range jsonOutWithStats.TaskQueuesInfos {
		if tq.Name == taskQueue && tq.Type == "workflow" {
			s.NotNil(tq.Stats, "Stats should be present when --report-task-queue-stats is provided")
			if tq.Stats != nil {
				foundWorkflowTQWithStats = true
				// Verify all stats fields are present and backlog count matches what we created
				s.Greater(tq.Stats.ApproximateBacklogCount, int64(0),
					"ApproximateBacklogCount should be non-zero with pending workflows")
				s.Greater(tq.Stats.ApproximateBacklogAge, int64(0),
					"ApproximateBacklogAge should be non-zero with pending workflows")
				s.Greater(tq.Stats.TasksAddRate, float32(0),
					"TasksAddRate should be non-zero after adding tasks")
				// BacklogIncreaseRate may be positive (if backlog is growing)
				// TasksDispatchRate should be 0 since no worker is running
				s.Equal(float32(0), tq.Stats.TasksDispatchRate,
					"TasksDispatchRate should be zero with no worker running")
			}
		}
	}
	s.True(foundWorkflowTQWithStats, "Should find workflow task queue with stats")

	// Cleanup: Restart worker and signal workflows to complete
	w2 := worker.New(s.Client, taskQueue, worker.Options{
		DeploymentOptions: worker.DeploymentOptions{
			UseVersioning: true,
			Version:       version,
		},
	})
	w2.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input any) (any, error) {
			workflow.GetSignalChannel(ctx, "complete-signal").Receive(ctx, nil)
			return nil, nil
		},
		workflow.RegisterOptions{
			Name:               "TestBacklogWorkflow",
			VersioningBehavior: workflow.VersioningBehaviorPinned,
		},
	)
	s.NoError(w2.Start())
	defer w2.Stop()

	// Signal all workflows to complete
	for _, run := range workflowRuns {
		err := s.Client.SignalWorkflow(s.Context, run.GetID(), run.GetRunID(), "complete-signal", nil)
		s.NoError(err)
	}

	// Wait for workflows to complete
	for _, run := range workflowRuns {
		s.NoError(run.Get(s.Context, nil))
	}
}

func (s *SharedServerSuite) TestCreateWorkerDeployment() {
	deploymentName := uuid.NewString()

	res := s.Execute(
		"worker", "deployment", "create",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully created worker deployment")

	// Wait for the deployment to appear
	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe",
			"--address", s.Address(),
			"--name", deploymentName,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	// Attempting to create a WD with the same name should fail with a conflict
	// error.
	res = s.Execute(
		"worker", "deployment", "create",
		"--address", s.Address(),
		"--name", deploymentName,
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "already exists")
}

func (s *SharedServerSuite) TestCreateWorkerDeploymentVersion_EmptyComputeConfig() {
	deploymentName := uuid.NewString()
	taskQueue := uuid.NewString()

	lazyCreatedBuildID := uuid.NewString()
	lazyCreatedVer := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        lazyCreatedBuildID,
	}

	// Create worker with explicit versioning. This will end up creating a
	// WorkerDeployment with the specified name. We will then manually create a
	// worker deployment version using the `temporal worker deployment
	// create-version` command.
	w1 := worker.New(s.Client, taskQueue, worker.Options{
		DeploymentOptions: worker.DeploymentOptions{
			UseVersioning: true,
			Version:       lazyCreatedVer,
		},
	})

	// Register a workflow with explicit Pinned versioning behavior to trigger
	// creation of the worker deployment.
	w1.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input any) (any, error) {
			workflow.GetSignalChannel(ctx, "complete-signal").Receive(ctx, nil)
			return nil, nil
		},
		workflow.RegisterOptions{
			Name:               "TestCreateWorkerDeploymentVersion_NoComputeConfig",
			VersioningBehavior: workflow.VersioningBehaviorPinned,
		},
	)

	s.NoError(w1.Start())

	// Wait for the lazily-created deployment to appear
	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", deploymentName,
			"--build-id", lazyCreatedBuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	// Now that we know the worker deployment exists (because the above
	// lazily-created worker deployment version ended up creating it), we will
	// manually create a new worker deployment version using the `temporal
	// worker deployment create-version` CLI command.
	noComputeConfigBuildID := uuid.NewString()

	res := s.Execute(
		"worker", "deployment", "create-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", noComputeConfigBuildID,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully created worker deployment version")

	// Wait for the deployment version to appear
	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", deploymentName,
			"--build-id", noComputeConfigBuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	// Check that there is no compute config returned for this WDV
	res = s.Execute(
		"worker", "deployment", "describe-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", noComputeConfigBuildID,
		"--output", "json",
	)
	s.NoError(res.Err)
	var jsonOut jsonDeploymentVersionInfoType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Nil(jsonOut.ComputeConfig, "ComputeConfig should be nil.")

	// Attempting to create a WDV with the same BuildID should fail with a
	// conflict error.
	res = s.Execute(
		"worker", "deployment", "create-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", noComputeConfigBuildID,
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "already exists")
}

func (s *SharedServerSuite) TestCreateWorkerDeploymentVersion_Errors() {
	deploymentName := uuid.NewString()
	taskQueue := uuid.NewString()

	lazyCreatedBuildID := uuid.NewString()
	lazyCreatedVer := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        lazyCreatedBuildID,
	}

	// Create worker with explicit versioning. This will end up creating a
	// WorkerDeployment with the specified name. We will then manually create a
	// worker deployment version using the `temporal worker deployment
	// create-version` command.
	w1 := worker.New(s.Client, taskQueue, worker.Options{
		DeploymentOptions: worker.DeploymentOptions{
			UseVersioning: true,
			Version:       lazyCreatedVer,
		},
	})

	// Register a workflow with explicit Pinned versioning behavior to trigger
	// creation of the worker deployment.
	w1.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input any) (any, error) {
			workflow.GetSignalChannel(ctx, "complete-signal").Receive(ctx, nil)
			return nil, nil
		},
		workflow.RegisterOptions{
			Name:               "TestCreateWorkerDeploymentVersion_Errors",
			VersioningBehavior: workflow.VersioningBehaviorPinned,
		},
	)

	s.NoError(w1.Start())

	// Wait for the lazily-created deployment to appear
	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", deploymentName,
			"--build-id", lazyCreatedBuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	// Create some WDVs with invalid compute config parameters.
	assumeRoleFailureBuildID := uuid.NewString()

	invokeARN := "arn:aws:lambda:us-east-1:123456789012:function:MyExampleFunction:1"
	assumeRoleARN := "arn:aws:iam::123456789012:role/MyServiceRole"
	assumeRoleExternalID := "external-id"

	res := s.Execute(
		"worker", "deployment", "create-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", assumeRoleFailureBuildID,
		"--aws-lambda-function-arn", invokeARN,
		"--aws-lambda-assume-role-arn", assumeRoleARN,
		"--aws-lambda-assume-role-external-id", assumeRoleExternalID,
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "failed to assume role arn:aws:iam::123456789012:role/MyServiceRole: operation error STS: AssumeRole")

	missingExternalIDBuildID := uuid.NewString()

	res = s.Execute(
		"worker", "deployment", "create-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", missingExternalIDBuildID,
		"--aws-lambda-function-arn", invokeARN,
		"--aws-lambda-assume-role-arn", assumeRoleARN,
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "missing required AWS Lambda provider detail: role_external_id")

	missingAssumeRoleBuildID := uuid.NewString()

	res = s.Execute(
		"worker", "deployment", "create-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", missingAssumeRoleBuildID,
		"--aws-lambda-function-arn", invokeARN,
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "missing required AWS Lambda provider detail: role")
}

// TODO(jaypipes): Enable this test when we have a way of ensuring AWS resource
// fixtures since the CLI test harness uses a real Temporal Server and a real
// Temporal Server validates any supplied AWS Lambda Function and Assume Role
// ARNs are good...
func (s *SharedServerSuite) TestCreateWorkerDeploymentVersion_LambdaComputeConfig() {
	s.T().Skip("AWS Lambda Function and Assume Role fixtures needed.")
	deploymentName := uuid.NewString()
	taskQueue := uuid.NewString()

	lazyCreatedBuildID := uuid.NewString()
	lazyCreatedVer := worker.WorkerDeploymentVersion{
		DeploymentName: deploymentName,
		BuildID:        lazyCreatedBuildID,
	}

	// Create worker with explicit versioning. This will end up creating a
	// WorkerDeployment with the specified name. We will then manually create a
	// worker deployment version using the `temporal worker deployment
	// create-version` command.
	w1 := worker.New(s.Client, taskQueue, worker.Options{
		DeploymentOptions: worker.DeploymentOptions{
			UseVersioning: true,
			Version:       lazyCreatedVer,
		},
	})

	// Register a workflow with explicit Pinned versioning behavior to trigger
	// creation of the worker deployment.
	w1.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input any) (any, error) {
			workflow.GetSignalChannel(ctx, "complete-signal").Receive(ctx, nil)
			return nil, nil
		},
		workflow.RegisterOptions{
			Name:               "TestCreateWorkerDeploymentVersion_LambdaComputeConfig",
			VersioningBehavior: workflow.VersioningBehaviorPinned,
		},
	)

	s.NoError(w1.Start())

	// Now that we know the worker deployment exists (because the above
	// lazily-created worker deployment version ended up creating it), we will
	// manually create a new worker deployment version using the `temporal
	// worker deployment create-version` CLI command.
	//
	// Create a WDV with a valid Compute Config specified and verify that the
	// compute config provider is displayed in the output of `temporal worker
	// deployment describe-version`
	computeConfigBuildID := uuid.NewString()

	invokeARN := "arn:aws:lambda:us-east-1:123456789012:function:MyExampleFunction:1"
	assumeRoleARN := "arn:aws:iam::123456789012:role/MyServiceRole"
	assumeRoleExternalID := "external-id"

	res := s.Execute(
		"worker", "deployment", "create-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", computeConfigBuildID,
		"--aws-lambda-function-arn", invokeARN,
		"--aws-lambda-assume-role-arn", assumeRoleARN,
		"--aws-lambda-assume-role-external-id", assumeRoleExternalID,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully created worker deployment version")

	// Wait for the deployment version to appear
	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--deployment-name", deploymentName,
			"--build-id", computeConfigBuildID,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	// Check that there is a compute config returned for this WDV
	res = s.Execute(
		"worker", "deployment", "describe-version",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", computeConfigBuildID,
		"--output", "json",
	)
	s.NoError(res.Err)
	jsonOut := jsonDeploymentVersionInfoType{}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.NotNil(jsonOut.ComputeConfig, "ComputeConfig should not be nil.")
}
