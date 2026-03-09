# Exposing the Temporal Dev Server on a Tailscale Network

### Introduction

When you run `temporal server start-dev`, the dev server binds to `localhost` by default.
This means only processes on your local machine can connect to it.
If you want teammates or other machines on your network to access the dev server, you typically need to configure port forwarding, firewall rules, or a VPN.

The `--tsnet` flag removes that friction.
It uses [tsnet](https://tailscale.com/kb/1244/tsnet), Tailscale's embedded networking library, to expose the dev server directly on your Tailscale network (your *tailnet*).
Once enabled, any machine on your tailnet can connect to the dev server using a stable hostname like `temporal-dev:7233` without any manual network configuration.

This is useful when you need to:

- Share a dev server with a teammate for debugging or demos.
- Connect to the dev server from a separate machine, such as a staging environment or a different laptop.
- Run Workers on remote machines against a local dev server during development.

In this guide, you'll learn how to configure and use the `--tsnet` flag with `temporal server start-dev`.

## Prerequisites

To use the `--tsnet` feature, you'll need:

- **Temporal CLI** built from this branch or a release that includes tsnet support.
- **A Tailscale account** with at least one other device on your tailnet. You can sign up at [tailscale.com](https://tailscale.com).
- **A Tailscale auth key** (recommended) for non-interactive authentication. You can generate one from the [Tailscale admin console](https://login.tailscale.com/admin/settings/keys). Reusable, ephemeral keys work well for development.

## Step 1 — Starting the Dev Server with Tailscale

To expose the dev server on your tailnet, add the `--tsnet` flag when starting the server.

The recommended approach is to set your auth key as an environment variable and then start the server:

```
export TS_AUTHKEY="tskey-auth-your-key-here"
temporal server start-dev --tsnet
```

The server starts as usual on `localhost`, and the tsnet integration creates a Tailscale node that proxies connections from your tailnet to the local server.
You'll see output similar to this:

```
Temporal CLI 0.0.0-DEV (Server 1.30.1, UI 2.45.3)

Temporal Server:  localhost:7233
Temporal UI:      http://localhost:8233
Temporal Metrics: http://localhost:49487/metrics
Tsnet gRPC:       temporal-dev:7233
Tsnet UI:         http://temporal-dev:8233
```

The last two lines confirm that the dev server is now accessible on your tailnet.
The node appears as **temporal-dev** in your Tailscale admin console.

You can also pass the auth key directly with the `--tsnet-authkey` flag:

```
temporal server start-dev --tsnet --tsnet-authkey "tskey-auth-your-key-here"
```

**Note:** If you omit both `--tsnet-authkey` and the `TS_AUTHKEY` environment variable, a Tailscale login URL is printed to the terminal for interactive authentication.

## Step 2 — Connecting from Another Machine

Now that the dev server is running on your tailnet, you can connect to it from any other machine on the same tailnet.

To verify connectivity, open a terminal on another machine and run:

```
temporal workflow list --address temporal-dev:7233
```

If the connection is successful, you'll see the list of Workflow Executions in the `default` Namespace (which may be empty on a fresh server).

You can also open the Web UI in a browser by navigating to `http://temporal-dev:8233` from any device on your tailnet.

When writing application code, point your Temporal Client at the tsnet address instead of `localhost`:

```
# Python example
client = await Client.connect("temporal-dev:7233", namespace="default")
```

```
// Go example
client, err := client.Dial(client.Options{HostPort: "temporal-dev:7233"})
```

This allows Workers and starter scripts on remote machines to interact with your dev server without any port forwarding.

## Step 3 — Customizing the Configuration

The `--tsnet` feature has several flags you can use to customize its behavior.

### Custom Hostname

By default, the Tailscale node is named **temporal-dev**.
To use a different hostname, pass the `--tsnet-hostname` flag:

```
temporal server start-dev --tsnet --tsnet-hostname my-temporal
```

The server will then be accessible at `my-temporal:7233` and `http://my-temporal:8233` on your tailnet.

### Headless Mode

If you don't need the Web UI, you can combine `--tsnet` with `--headless`.
In headless mode, only the gRPC port is proxied over the tailnet:

```
temporal server start-dev --tsnet --headless
```

### Custom State Directory

The tsnet integration stores its Tailscale node state (authentication tokens, keys) in a local directory.
By default, this is a `tsnet-temporal-dev` directory under your operating system's user config directory (for example, `~/Library/Application Support/tsnet-temporal-dev` on macOS).

To use a different location, pass the `--tsnet-state-dir` flag:

```
temporal server start-dev --tsnet --tsnet-state-dir /path/to/state
```

### Custom Ports

The tsnet listeners use the same port numbers as the local server.
To change the gRPC or UI ports, use the standard `--port` and `--ui-port` flags:

```
temporal server start-dev --tsnet --port 7234 --ui-port 8234
```

The tsnet node will then listen on `temporal-dev:7234` and `http://temporal-dev:8234`.

## Step 4 — Stopping the Server

When you press `CTRL+C`, the server shuts down cleanly.
The tsnet proxy listeners close first, which stops accepting new connections from the tailnet.
Then the Temporal dev server itself shuts down.

If you used an ephemeral auth key, the node is automatically removed from your tailnet after shutdown.
If you used a reusable key, the node remains registered but goes offline.
You can remove it manually from the [Tailscale admin console](https://login.tailscale.com/admin/machines) if needed.

## Flag Reference

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tsnet` | bool | `false` | Enable Tailscale network exposure for the dev server. |
| `--tsnet-hostname` | string | `temporal-dev` | Tailscale hostname. Appears as `<hostname>.<tailnet>.ts.net`. |
| `--tsnet-authkey` | string | (empty) | Tailscale auth key. Falls back to `TS_AUTHKEY` env var. |
| `--tsnet-state-dir` | string | (auto) | State directory for tsnet node data. |

## Conclusion

You configured the Temporal dev server to be accessible over a Tailscale network using the `--tsnet` flag.
Any machine on your tailnet can now connect to the dev server's gRPC and Web UI ports using the Tailscale hostname.

This is particularly useful for collaborative development, running distributed Workers, and testing from multiple machines without managing network infrastructure.

For more information about Tailscale and tsnet, see the [Tailscale documentation](https://tailscale.com/kb/1244/tsnet).
For more information about the Temporal CLI, see the [Temporal CLI documentation](https://docs.temporal.io/cli).
