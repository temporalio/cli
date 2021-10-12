module github.com/temporalio/tctl

go 1.16

replace github.com/temporalio/tctl-core => /home/user0/tc

require (
	github.com/fatih/color v1.13.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-hclog v0.16.1
	github.com/hashicorp/go-plugin v1.4.1
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pborman/uuid v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/temporalio/tctl-core v0.1.0
	github.com/uber-go/tally v3.4.2+incompatible
	github.com/urfave/cli v1.22.5
	github.com/urfave/cli/v2 v2.3.0
	go.temporal.io/api v1.5.0
	go.temporal.io/sdk v1.10.0
	go.temporal.io/server v1.12.1
	google.golang.org/grpc v1.40.0
)
