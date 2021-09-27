[![build](https://github.com/temporalio/tctl/actions/workflows/ci.yml/badge.svg)](https://github.com/temporalio/tctl/actions/workflows/ci.yml)

**NOT READY FOR PRODUCTION**

This is an ongoing work on tctl UX based on proposals:
 - https://github.com/temporalio/proposals/tree/master/cli
 - https://github.com/temporalio/proposals/pulls (marked with tctl)

Documentation for the Temporal command line interface is located at our [main site](https://docs.temporal.io/docs/system-tools/tctl).

## Quick Start
Run `make` from the project root. You should see an executable file called `tctl`. Try a few example commands to 
get started:   
`./tctl` for help on top level commands and global options   
`./tctl namespace` for help on namespace operations  
`./tctl workflow` for help on workflow operations  
`./tctl taskqueue` for help on tasklist operations  
(`./tctl help`, `./tctl help [namespace|workflow]` will also print help messages)

**Note:** Make sure you have a Temporal server running before using the CLI.

## License

MIT License, please see [LICENSE](https://github.com/temporalio/temporal-cli/blob/master/LICENSE) for details.
