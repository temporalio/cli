[![build](https://github.com/temporalio/tctl/actions/workflows/test.yml/badge.svg)](https://github.com/temporalio/tctl/actions/workflows/test.yml)

The Temporal CLI is a command-line tool you can use to perform various tasks on a Temporal Server. It can perform namespace operations such as register, update, and describe as well as Workflow operations like start Workflow, show Workflow history, and signal Workflow.

Documentation for the Temporal command line interface is located at our [main site](https://docs.temporal.io/docs/system-tools/tctl).

## Quick Start
Run `make` from the project root. You should see an executable file called `tctl`. Try a few example commands to 
get started:   
`./tctl` for help on top level commands and global options   
`./tctl namespace` for help on namespace operations  
`./tctl workflow` for help on workflow operations  
`./tctl task-queue` for help on tasklist operations  
(`./tctl help`, `./tctl help [namespace|workflow]` will also print help messages)

**Note:** Make sure you have a Temporal server running before using the CLI.

### Trying out tctl v1.14.0-alpha with the updated UX

**Note** Switching to `tctl v1.14.0-alpha` is not recommended on production environments.

The package contains both `tctl v1.13.1` and the updated `tctl v1.14.0-alpha`. Version v1.14.0-alpha brings updated UX, commands and flags semantics, new features ([see details](https://github.com/temporalio/proposals/tree/master/cli)). Please expect more of upcoming changes in v1.14.0

By default, executing tctl commands will execute commands from v1.13.1. In order to switch to experimental v1.14.0-alpha run

```
tctl config set version next
```

This will create a configuration file (`~/.config/temporalio/tctl.yaml`) and set tctl to v1.14.0-alpha.

To switch back to the stable v1.13.1, run

```
tctl config set --version current
```

## License

MIT License, please see [LICENSE](https://github.com/temporalio/tctl/blob/master/LICENSE) for details.
