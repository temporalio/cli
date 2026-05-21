## Related issues

<!-- Closes #123 -->

## What changed?

<!-- Describe what this PR does at a high level. -->

## Checklist

<!-- Remove items that don't apply to this PR -->

**Stability**
- [ ] Breaking changes are marked with 💥 in the PR title and release notes
- [ ] Changes to JSON output (`-o json` / `-o jsonl`) are treated as breaking changes

**Design**
- [ ] This feature only uses APIs exposed by the OSS server (no Cloud-specific behavior)
- [ ] New commands follow noun→verb structure (`temporal <resource> <action>`)
- [ ] New flags are named after the API concept, not the implementation mechanism
- [ ] New flags don't duplicate an existing flag that serves the same purpose
- [ ] New flags do not have short aliases without strong justification
- [ ] Experimental features are marked with `(Experimental)` in `commands.yaml`

**Help text** (see style guide at the top of `commands.yaml`)
- [ ] All flags shown in help text and examples are implemented and functional
- [ ] Summaries use sentence case, don't reword the command name, and have no trailing period
- [ ] Long descriptions end with a period and include at least one example invocation
- [ ] Examples use long flags (`--namespace`, not `-n`), one flag per line
- [ ] Placeholder values use `YourXxx` form (`YourWorkflowId`, `YourNamespace`)

**Behavior**
- [ ] Results go to stdout; errors and diagnostics go to stderr
- [ ] Error messages are lowercase with no trailing punctuation

**Tests**
- [ ] Added functional test(s) (`SharedServerSuite`)
- [ ] Added unit test(s) (`func TestXxx`) where applicable

## Manual tests

**Setup**
```
temporal server start-dev --headless
temporal workflow start \
    --type YourWorkflowType \
    --task-queue YourTaskQueue \
    --workflow-id YourWorkflowId
```

**Happy path**
```
$ temporal <command> \
    --flag value
<expected output>
```

**Error case**
```
$ temporal <command> \
    --invalid-combination
Error: <expected error message>
$ echo $?
1
```

**Composition** <!-- How might a user combine this with existing commands? e.g. using the output of one command as input to another -->
```
$ temporal <command-one> ...
$ temporal <command-two> --flag <value-from-above>
<expected output>
```
