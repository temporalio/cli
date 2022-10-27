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

### Trying out the new `tctl next` with updated UX

**Note** Switching to `tctl next` is not recommended on production environments.

The package contains both `tctl v1` and the updated `tctl next`. Version `next` brings updated UX, new commands and flags semantics, new features ([see details](https://github.com/temporalio/proposals/tree/master/cli)). Please expect more of upcoming changes in `tctl next`

By default, executing tctl commands will execute commands from tctl v1. In order to switch to experimental `tctl next` run

```
tctl config set version next
```

This will create a configuration file (`~/.config/temporalio/tctl.yaml`) and set tctl to `next`.

To switch back to the stable v1, run

```
tctl config set version current
```

## License

MIT License, please see [LICENSE](https://github.com/temporalio/tctl/blob/master/LICENSE) for details.
