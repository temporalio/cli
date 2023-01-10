
Query a Workflow Execution with __stack_trace as the query type

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input, -i**="": Optional input for the query, in JSON format. If there are multiple parameters, concatenate them and separate by space

**--input-file**="": Optional input for the query from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. Input from file will be overwrite by input from command line

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--reject-condition**="": Optional flag to reject queries based on Workflow state. Valid values are "not_open" and "not_completed_cleanly"

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

