package temporalcli_test

import (
	"encoding/json"

	"github.com/google/uuid"
)

func (s *SharedServerSuite) TestTaskQueue_BuildId() {
	buildIdTaskQueue := uuid.NewString()
	res := s.Execute(
		"task-queue", "update-build-ids", "add-new-default",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--build-id", "1.0",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-ids", "add-new-default",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--build-id", "2.0",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-ids", "add-new-compatible",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--build-id", "1.1",
		"--existing-compatible-build-id", "1.0",
	)
	s.NoError(res.Err)

	// Text
	res = s.Execute(
		"task-queue", "get-build-ids",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "[1.0 1.1]", "1.1", "false")
	s.ContainsOnSameLine(res.Stdout.String(), "[2.0]", "2.0", "true")

	// json
	res = s.Execute(
		"task-queue", "get-build-ids",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)
	type rowType struct {
		BuildIds      []string
		DefaultForSet string
		IsDefaultSet  bool
	}
	var jsonOut []rowType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal([]rowType{
		rowType{
			BuildIds:      []string{"1.0", "1.1"},
			DefaultForSet: "1.1",
			IsDefaultSet:  false,
		},
		rowType{
			BuildIds:      []string{"2.0"},
			DefaultForSet: "2.0",
			IsDefaultSet:  true,
		},
	}, jsonOut)

	// Test promote commands
	res = s.Execute(
		"task-queue", "update-build-ids", "promote-id-in-set",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--build-id", "1.0",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-ids", "promote-set",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--build-id", "1.0",
	)
	s.NoError(res.Err)

	// Text
	res = s.Execute(
		"task-queue", "get-build-ids",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "[2.0]", "2.0", "false")
	s.ContainsOnSameLine(res.Stdout.String(), "[1.1 1.0]", "1.0", "true")

	// json
	res = s.Execute(
		"task-queue", "get-build-ids",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal([]rowType{
		rowType{
			BuildIds:      []string{"2.0"},
			DefaultForSet: "2.0",
			IsDefaultSet:  false,
		},
		rowType{
			BuildIds:      []string{"1.1", "1.0"},
			DefaultForSet: "1.0",
			IsDefaultSet:  true,
		},
	}, jsonOut)

	// Test reachability
	res = s.Execute(
		"task-queue", "get-build-id-reachability",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--build-id", "1.0",
		"--build-id", "1.1",
		"--build-id", "2.0",
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "1.0", buildIdTaskQueue, "[NewWorkflows]")
	s.ContainsOnSameLine(res.Stdout.String(), "1.1", buildIdTaskQueue, "[]")
	// Not clear why 2.0 can create new workflows if it is not the default...
	s.ContainsOnSameLine(res.Stdout.String(), "2.0", buildIdTaskQueue, "[NewWorkflows]")

	res = s.Execute(
		"task-queue", "get-build-id-reachability",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--build-id", "1.0",
		"--build-id", "1.1",
		"--build-id", "2.0",
	)

	s.NoError(res.Err)
	type rowReachType struct {
		BuildId      string
		TaskQueue    string
		Reachability []string
	}

	var jsonReachOut []rowReachType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonReachOut))

	s.Equal([]rowReachType{
		rowReachType{
			BuildId:      "1.0",
			TaskQueue:    buildIdTaskQueue,
			Reachability: []string{"NewWorkflows"},
		},
		rowReachType{
			BuildId:      "1.1",
			TaskQueue:    buildIdTaskQueue,
			Reachability: []string{},
		},
		rowReachType{
			BuildId:      "2.0",
			TaskQueue:    buildIdTaskQueue,
			Reachability: []string{"NewWorkflows"},
		},
	}, jsonReachOut)
}
