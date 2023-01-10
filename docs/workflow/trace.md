
Trace progress of a Workflow Execution and its children

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--concurrency**="": Request concurrency (default: 10)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--depth**="": Number of child workflows to expand, -1 to expand all child workflows (default: -1)

**--env**="": Env name to read the client environment variables from (default: default)

**--fold**="": Statuses for which child workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if --no-fold supplied (default: completed,canceled,terminated)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-fold**: Disable folding. All child workflows within the set depth will be fetched and displayed

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

