# gh-assert

Declarative environment assertions for GitHub Actions.

`gh-assert` moves common shell checks into a small YAML contract. Version 0.0.4 validates environment variables and workflow inputs at runtime, and validates reusable workflow interfaces statically.

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

A contract declares `env`, `inputs`, or both. Supported constraints are `required`, string `enum` and `pattern`, integer `min` and `max`, and boolean type checking. Defaults, expressions, conditional rules, and step or job contracts are outside v0.0.4. Composite Action contracts are outside v0.0.4.

## Runtime assertion

Place the Action in a workflow step and pass the environment values that the contract names:

```yaml
- uses: rin2yh/gh-assert@v0.0.4
  with:
    contract: .github/workflows/deploy_assert.yml
    version: v0.0.4
  env:
    ENVIRONMENT: ${{ vars.DEPLOY_ENVIRONMENT }}
    RETRIES: ${{ vars.DEPLOY_RETRIES }}
    DRY_RUN: ${{ vars.DRY_RUN }}
```

Pick a tag that [Releases](https://github.com/rin2yh/gh-assert/releases) already publishes; the example above names the next one. The Action runs on Linux, macOS and Windows runners on x64 and arm64. It picks the release binary for `RUNNER_OS` and `RUNNER_ARCH` and verifies it against the release's `checksums.txt` before execution. It exits non-zero when a required variable is missing or empty, a value has the wrong type, or a declared constraint fails. Values are not printed in diagnostics.

When `contract` is omitted, the Action discovers every `*_assert.yml` under `.github`.

In a reusable workflow, use the Action normally:

```yaml
on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: rin2yh/gh-assert@v0.0.4
        with:
          contract: .github/workflows/deploy_assert.yml
          version: v0.0.4
          workflow-inputs: ${{ toJSON(inputs) }}
        env:
          ENVIRONMENT: ${{ inputs.environment }}
```

For a sibling pair such as `deploy.yml` and `deploy_assert.yml`, gh-assert compares the names, `required` settings and types in `workflow_call.inputs` with the contract. A contract `integer` corresponds to a GitHub Actions `number`. At runtime, the Action asserts the explicitly forwarded input values and the declared environment variables.

GitHub does not automatically pass a reusable workflow's `inputs` context to a Composite Action, so `workflow-inputs: ${{ toJSON(inputs) }}` is required when the contract declares inputs. gh-assert fails instead of silently skipping runtime input assertions when it is omitted. Values are still hidden from diagnostics.

## Validate a contract

Install the command:

```bash
go install github.com/rin2yh/gh-assert/cmd/gh-assert@latest
```

Validate all contracts under `.github`:

```bash
gh-assert validate
```

To validate one contract, pass its path as a positional argument.

```bash
gh-assert validate .github/workflows/deploy_assert.yml
```

Validation checks YAML syntax, supported fields and types, regular expressions, and integer ranges. For reusable workflows, it also checks that `workflow_call.inputs` matches the sibling contract's input names, `required` settings and types. It does not execute a workflow.

Runtime assertion follows the same path rule:

```bash
gh-assert
gh-assert .github/workflows/deploy_assert.yml
```

## Development

```bash
go test ./...
go vet ./...
```

Releases are tagged and published by tagpr, and the same run builds the binaries with GoReleaser and uploads them. A release carries `gh-assert-<os>-<arch>` for linux, darwin and windows on amd64 and arm64, plus `checksums.txt`. The repository is available at `github.com/rin2yh/gh-assert`.

## License

MIT.
