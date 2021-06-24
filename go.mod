module github.com/temporalio/tctl

replace github.com/temporalio/shared-go => /home/user0/shared

replace go.temporal.io/server => /home/user0/server

go 1.16

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.10.0
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-hclog v0.16.1
	github.com/hashicorp/go-plugin v1.4.1
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pborman/uuid v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/temporalio/shared-go v1.0.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/valyala/fastjson v1.6.3
	go.temporal.io/api v1.4.1-0.20210429213054-a9a257b5cf16
	go.temporal.io/sdk v1.7.0
	go.temporal.io/server v1.9.2
	google.golang.org/grpc v1.38.0
)
