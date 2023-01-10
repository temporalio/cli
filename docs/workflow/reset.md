### reset

Reset a Workflow Execution by event Id or reset type

**--address**
host:port for Temporal frontend service

**--codec-auth**
Authorization header to set for requests to Codec Server

**--codec-endpoint**
Remote Codec Server Endpoint

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
Optional timeout for context of RPC call in seconds (default: 5)

**--env**
Env name to read the client environment variables from (default: default)

**--event-id**
The eventId of any event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace**
Alias: ** -n**
Temporal workflow namespace (default: default)

**--reapply-type**
Event types to reapply after the reset point: Signal, None, . (default: All)

**--reason**
Reason to reset

**--run-id**
Alias: ** -r**
Run Id

**--tls-ca-path**
Path to server CA certificate

**--tls-cert-path**
Path to x509 certificate

**--tls-disable-host-verification**
Disable tls host name verification (tls must be enabled)

**--tls-key-path**
Path to private key

**--tls-server-name**
Override for target server name

**--type**
Event type to which you want to reset: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

**--workflow-id**
Alias: ** -w**
Workflow Id

