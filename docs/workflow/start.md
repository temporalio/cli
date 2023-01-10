### start

Start a new Workflow Execution

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

**--cron**
Optional cron schedule for the Workflow. Cron spec is as following:
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--env**
Env name to read the client environment variables from (default: default)

**--execution-timeout**
Workflow Execution timeout, including retries and continue-as-new (seconds) (default: 0)

**--fields**
customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--id-reuse-policy**
Configure if the same Workflow Id is allowed for use in new Workflow Execution. Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning

**--input**
Alias: ** -i**
Optional input for the Workflow in JSON format. Pass "null" for null values

**--input-file**
Pass an optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line overwrites input from the file

**--limit**
number of items to print (default: 0)

**--max-field-length**
Maximum length for each attribute field (default: 0)

**--memo**
Pass a memo in a format key=value. Use valid JSON formats for value

**--memo-file**
Pass a memo from a file, where each line follows the format key=value. Use valid JSON formats for value

**--namespace**
Alias: ** -n**
Temporal workflow namespace (default: default)

**--no-pager**
Alias: ** -P**
disable interactive pager

**--output**
Alias: ** -o**
format output as: table, json, card. (default: table)

**--pager**
pager to use: less, more, favoritePager..

**--run-timeout**
Single workflow run timeout (seconds) (default: 0)

**--search-attribute**
Pass Search Attribute in a format key=value. Use valid JSON formats for value

**--task-queue**
Alias: ** -t**
Task queue

**--task-timeout**
Workflow task start to close timeout (seconds) (default: 10)

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

**--type**
Workflow type name

**--workflow-id**
Alias: ** -w**
Workflow Id

