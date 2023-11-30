(under development)

Known incompatibilities:

NOTE: All of these incompatibilities are intentional and almost all decisions can be reverted if decided.

* Removed `--memo-file` from workflow args
* `--color` not currently implemented everywhere (like for logs)
* Removed paging by default (i.e. basically `--no-pager` behavior)
* Duration arguments require trailing unit (i.e. `--workflow-timeout 5` is now `--workflow-timeout 5s`)
* `--output table` and `--output card` blended to `--output text` (the default), but we may let table options be applied
  as separate params
* `TEMPORAL_CLI_SHOW_STACKS` - no stack trace-based errors
* `--tls-ca-path` cannot be a URL
* Not explicitly setting TLS SNI name from host of URL
* JSON output for things like workflow start use more JSON-like field names
* Workflow history JSON not dumped as part of `workflow execute` when JSON set
* Concept of `--fields long` is gone, now whether some more verbose fields are emitted is controlled more specifically
* To get accurate workflow result, workflow follows runs for `workflow execute`
* Removed the `-f` alias for `--follow` on `workflow show`
* `server start-dev` will reuse the root logger which means:
  * Default is text (or "pretty") instead JSON
  * No way to set level to "fatal" only
  * All panic and fatal logs are just error logs
  * Goes to stderr instead of stdout
* `server start-dev --db-filename` no longer auto-creates directory path of 0777 dirs if not present

Known improvements:

* Cobra (any arg at any place)
* Customize path to env file
* Global log-level customization
* Global json vs text data output customization
* Markdown-based code generation
* Solid test framework
* Added `--input-encoding` to support more payload types (e.g. can support proto JSON and even proto binary)
* Library available for docs team to write doc generator with
* Only log or data content written, so disabling log means all data could be consumed with JSON tooling
* Properly gives failing status code if workflow fails on "execute" but JSON output is set
* `--color` is available to disable any coloring
* Dev server reuses logger meaning it is on stderr by default

Notes about approach taken:

* Did not spend time trying to improve documentation, so all of the inconsistent documentation remains and should be
  cleaned up separately
* Did not spend (much) time trying to completely change behavior or commands
* Compatibility intentionally retained in most reasonable ways
* File-level copyright notices retained on places with DataDog

Contribution rules:

* Follow rules in commands.md
* Refactoring and reuse is welcome
* Avoid package sprawl and complication
* Try to only use logger and printer for output
* Command testing (does not apply to unit tests that are not testing commands)
  * Use the command harness (create a new one for each different option if needed)
  * Name command tests as `Test<CamelCaseCommand>_<Qualifier>`, e.g. a simple "temporal server start dev" test may be
    named `TestServerStartDev_Simple`. Can test multiple subcommands at once and `CamelCaseCommand` can just be the
    parent, e.g. a simple test of different "temporal env" commands may be named `TestEnv_Simple`.

TODO:

* Version via goreleaser
* "card" output?
* Env variables
* Workflow show max-field-length?
* Workflow start delay: https://github.com/temporalio/cli/pull/402
* Enhance task queue describe: https://github.com/temporalio/cli/pull/399