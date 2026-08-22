# gh-assert

Declarative environment assertions for GitHub Actions.

`gh-assert` moves common shell checks into a small YAML contract. Version 0.2 validates environment variables and `workflow_dispatch` inputs at runtime inside a GitHub Action, and validates the contract definition from the command line.

## Contract

Create a file such as `.github/workflows/deploy_assert.yml`:

```yaml
env:
  ENVIRONMENT:
    required: true
    type:
      string:
        enum: [staging, production]
        pattern: '^[a-z]+$'
  RETRIES:
    type:
      integer:
        min: 0
        max: 5
  DRY_RUN:
    type:
      boolean: {}
inputs:
  environment:
    required: true
    type:
      string:
        enum: [staging, production]
  retries:
    type:
      integer:
        min: 0
        max: 5
```

A contract declares `env`, `inputs`, or both. Supported constraints are `required`, string `enum` and `pattern`, integer `min` and `max`, and boolean type checking. Defaults, expressions, conditional rules, and step or job contracts are outside v0.2. Reusable Workflow and Composite Action contracts are outside v0.2.

## Runtime assertion

Place the Action in a workflow step and pass the environment values that the contract names:

```yaml
- uses: rin2yh/gh-assert@v1
  with:
    contract: .github/workflows/deploy_assert.yml
    version: v0.2.0
  env:
    ENVIRONMENT: ${{ vars.DEPLOY_ENVIRONMENT }}
    RETRIES: ${{ vars.DEPLOY_RETRIES }}
    DRY_RUN: ${{ vars.DRY_RUN }}
```

Inputs need no wiring. `gh-assert` reads them from the event payload of the running workflow, so the contract is the only place that names them:

```yaml
on:
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [staging, production]
      retries:
        type: number
```

The Action targets `ubuntu-latest` in v0.2. It downloads the Linux amd64 release binary and verifies its checksum before execution. It exits non-zero when a required variable or input is missing or empty, a value has the wrong type, or a declared constraint fails. Values are not printed in diagnostics.

A contract that declares one or more `inputs` rules asserts only `workflow_dispatch` runs. On any other event the command fails with an exit code of 2 rather than reporting a passing assertion. An empty `inputs: {}` section declares no rule and asserts nothing, the same as an empty `env: {}` section, so it does not restrict the event. `workflow_call` inputs are v0.3.

When `contract` is omitted, the Action discovers every `*_assert.yml` under `.github`.

## Validate a contract

Install the extension:

```bash
gh extension install rin2yh/gh-assert
```

Validate all contracts under `.github`:

```bash
gh assert validate
```

To validate one contract, pass its path as a positional argument.

```bash
gh assert validate .github/workflows/deploy_assert.yml
```

The same command is available as `gh-assert validate ...` after building locally. Validation checks YAML syntax, supported fields and types, regular expressions, and integer ranges. It does not execute a workflow.

Runtime assertion follows the same path rule:

```bash
gh assert
gh assert .github/workflows/deploy_assert.yml
```

## Development

```bash
go test ./...
go vet ./...
```

The project is released as a precompiled GitHub CLI Extension. Releases are tagged and published by tagpr. The repository is available at `github.com/rin2yh/gh-assert`.

## License

MIT.
