### create

Create a new schedule

**--address**="": host:port for Temporal frontend service

**--calendar**="": Calendar specification in JSON, e.g. {"dayOfWeek":"Fri","hour":"17","minute":"5"}

**--catchup-window**="": Maximum allowed catch-up time if server is down

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--cron**="": Calendar specification as cron string, e.g. "30 2 * * 5" or "@daily"

**--end-time**="": Overall schedule end time

**--env**="": Env name to read the client environment variables from (default: default)

**--execution-timeout**="": Workflow Execution timeout, including retries and continue-as-new (seconds) (default: 0)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input, -i**="": Optional input for the Workflow in JSON format. Pass "null" for null values

**--input-file**="": Pass an optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line overwrites input from the file

**--interval**="": Interval duration, e.g. 90m, or 90m/13m to include phase offset

**--jitter**="": Jitter duration

**--max-field-length**="": Maximum length for each attribute field (default: 0)

**--memo**="": Set a memo on a schedule. Format: key=value. Use valid JSON formats for value

**--memo-file**="": Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--notes**="": Initial value of notes field

**--overlap-policy**="": Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll

**--pause**: Initial value of paused state

**--pause-on-failure**: Pause schedule after any workflow failure

**--remaining-actions**="": Total number of actions allowed (default: 0)

**--run-timeout**="": Single workflow run timeout (seconds) (default: 0)

**--schedule-id, -s**="": Schedule Id

**--search-attribute**="": Set Search Attribute on a schedule. Format: key=value. Use valid JSON formats for value

**--start-time**="": Overall schedule start time

**--task-queue, -t**="": Task queue

**--task-timeout**="": Workflow task start to close timeout (seconds) (default: 10)

**--time-zone**="": Time zone (IANA name)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

**--workflow-type**="": Workflow type name

