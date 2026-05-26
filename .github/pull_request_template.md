## Related issues

<!-- Closes #123 -->

## What changed?

<!-- Describe what this PR does at a high level. -->

## Checklist

<!-- Your PR should satisfy all these requirements. However, feel free to remove items that don't apply to the PR. Consider giving this checklist to an AI agent before opening your PR. -->

**Stability**
- [ ] Breaking changes are marked with 💥 in the PR title and release notes
- [ ] Changes to JSON output (`-o json` / `-o jsonl`) are treated as breaking changes

**Design**
- [ ] This feature does not depend on Cloud-only APIs or behavior (it works against an OSS server)
- [ ] New commands follow `temporal <noun> <verb>` structure (e.g. `temporal workflow start`)
- [ ] New flags are named after the API concept, not the implementation mechanism (good: `--search-attribute`, bad: `--index-field`)
- [ ] New flags don't duplicate an existing flag that serves the same purpose
- [ ] New flags do not have short aliases without strong justification
- [ ] Experimental features are marked with `(Experimental)` in `commands.yaml`

**Help text** (see style guide at the top of `commands.yaml`)
- [ ] All flags shown in help text and examples are implemented and functional
- [ ] Summaries use sentence case and have no trailing period
- [ ] Long descriptions end with a period and include at least one example invocation
- [ ] Examples use long flags (`--namespace`, not `-n`), one flag per line
- [ ] Placeholder values use `YourXxx` form (`YourWorkflowId`, `YourNamespace`)

**Behavior**
- [ ] Results go to stdout; errors and warnings go to stderr
- [ ] Error messages are lowercase with no trailing punctuation

**Tests**
- [ ] Added functional test(s) (`SharedServerSuite`)
- [ ] Added unit test(s) (`func TestXxx`) where applicable

## Manual tests

<!-- Edit the code samples below to provide setup and happy-path and error-path testing instructions. -->

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
