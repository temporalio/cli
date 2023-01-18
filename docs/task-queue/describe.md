### describe

Describes the Workers that have recently polled on this Task Queue

    The Server records the last time of each poll request.
    
    Poll requests can last up to a minute, so a LastAccessTime under a minute is normal.
    If it's over a minute, then likely either the Worker is at capacity (all Workflow and Activity slots are full) or it has shut down.
    Once it has been 5 minutes since the last poll request, the Worker is removed from the list.
    
    RatePerSecond is the maximum Activities per second the Worker will execute.

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

**--fields**
customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace**
Alias: **-n**
Temporal workflow namespace (default: default)

**--output**
Alias: **-o**
format output as: table, json, card. (default: table)

**--task-queue**
Alias: **-t**
Task Queue name.

**--task-queue-type**
Task Queue type [workflow|activity] (default: workflow)

**--time-format**
format time as: relative, iso, raw. (default: relative)

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

