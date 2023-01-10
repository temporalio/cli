
Operation on Temporal server

### namespace

Operations on namespaces

#### describe

Describe a Namespace by name or Id

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--namespace-id**="": Namespace Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### list

List all Namespaces

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### register

Register a new Namespace

**--active-cluster**="": Active cluster name

**--address**="": host:port for Temporal frontend service

**--cluster**="": Cluster name

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--data**="": Namespace data in a format key=value

**--description**="": Namespace description

**--email**="": Owner email

**--env**="": Env name to read the client environment variables from (default: default)

**--global**="": Flag to indicate whether namespace is a global namespace

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--history-archival-state**="": Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**="": Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--retention**="": Workflow Execution retention

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--visibility-archival-state**="": Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**="": Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### update

Update a Namespace

**--active-cluster**="": Active cluster name

**--address**="": host:port for Temporal frontend service

**--cluster**="": Cluster name

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--data**="": Namespace data in a format key=value

**--description**="": Namespace description

**--email**="": Owner email

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--history-archival-state**="": Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**="": Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--promote-global**: Promote local namespace to global namespace

**--reason**="": Reason for the operation

**--retention**="": Workflow Execution retention

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--visibility-archival-state**="": Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**="": Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### delete

Delete existing Namespace

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--yes, -y**: Confirm all prompts

### search-attribute

Operations on search attributes

#### create

Add custom search attributes

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--name**="": Search attribute name

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--type**="": Search attribute type: [Text Keyword Int Double Bool Datetime KeywordList]

**--yes, -y**: Confirm all prompts

#### list

List search attributes that can be used in list workflow query

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### remove

Remove custom search attributes metadata only (Elasticsearch index schema is not modified)

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--name**="": Search attribute name

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--yes, -y**: Confirm all prompts

### cluster

Operations on a Temporal cluster

#### health

Check health of frontend service

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### describe

Show information about the cluster

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### system

Show information about the system and capabilities

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### upsert

Add or update a remote cluster

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--enable-connection**: Enable cross cluster connection

**--env**="": Env name to read the client environment variables from (default: default)

**--frontend-address**="": Frontend address of the remote cluster

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### list

List all remote clusters

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--limit**="": number of items to print (default: 0)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": pager to use: less, more, favoritePager..

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### remove

Remove a remote cluster

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--name**="": Frontend address of the remote cluster

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

