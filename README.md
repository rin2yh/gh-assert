# gh-assert

Declarative environment assertions for GitHub Actions.

`gh-assert` moves common shell checks into a small YAML contract. It validates environment variables and inputs at runtime, and validates Reusable Workflow and Composite Action interfaces statically.

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

A contract declares `env`, `inputs`, or both. Supported constraints are `required`, string `enum` and `pattern`, integer `min` and `max`, and boolean type checking. Defaults, expressions, conditional rules, and step or job contracts are outside the supported contract format.

## Runtime assertion

Place the Action in a workflow step and pass the environment values that the contract names:

```yaml
- uses: rin2yh/gh-assert@<full-length-commit-sha> # v0.0.4
  with:
    contract: .github/workflows/deploy_assert.yml
  env:
    ENVIRONMENT: ${{ vars.DEPLOY_ENVIRONMENT }}
    RETRIES: ${{ vars.DEPLOY_RETRIES }}
    DRY_RUN: ${{ vars.DRY_RUN }}
```

Replace `<full-length-commit-sha>` with the full commit SHA for a published release, and keep the version comment so Dependabot can track updates. The Action embeds the matching release version, so no separate version input is needed. It runs on Linux, macOS and Windows runners on x64 and arm64, picks the release binary for `RUNNER_OS` and `RUNNER_ARCH`, and verifies it against the release's `checksums.txt` before execution. It exits non-zero when a required variable is missing or empty, a value has the wrong type, or a declared constraint fails. Values are not printed in diagnostics.

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
      - uses: rin2yh/gh-assert@<full-length-commit-sha> # v0.1.0
        with:
          contract: .github/workflows/deploy_assert.yml
          inputs: ${{ toJSON(inputs) }}
        env:
          ENVIRONMENT: ${{ inputs.environment }}
```

For a sibling pair such as `deploy.yml` and `deploy_assert.yml`, gh-assert compares the names, `required` settings and types in `workflow_call.inputs` with the contract. A contract `integer` corresponds to a GitHub Actions `number`. At runtime, the Action asserts the explicitly forwarded input values and the declared environment variables.

GitHub does not automatically pass a reusable workflow's `inputs` context to a Composite Action, so `inputs: ${{ toJSON(inputs) }}` is required when the contract declares inputs. gh-assert fails instead of silently skipping runtime input assertions when it is omitted. Values are still hidden from diagnostics.

## Composite Actions

Place a contract next to a Composite Action:

```text
.github/actions/deploy/
  action.yml
  action_assert.yml
```

The contract uses the same `inputs` and `env` sections as a workflow contract. Add gh-assert as the first step of the Composite Action and forward the Action's inputs:

```yaml
runs:
  using: composite
  steps:
    - uses: rin2yh/gh-assert@<full-length-commit-sha> # v0.1.0
      with:
        contract: .github/actions/deploy/action_assert.yml
        inputs: ${{ toJSON(inputs) }}
```

Static validation compares the input names and `required` settings in `action.yml` with `action_assert.yml`. Composite Action inputs are strings at the GitHub Actions interface, but a contract can apply semantic `string`, `integer`, or `boolean` validation to their runtime values. Environment variables passed to the Composite Action remain available to gh-assert. They have no declaration in Action metadata, so their contract is checked at runtime.

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

Validation checks YAML syntax, supported fields and types, regular expressions, and integer ranges. For Reusable Workflows, it also checks that `workflow_call.inputs` matches the sibling contract's input names, `required` settings and types. For Composite Actions, it checks that the sibling `action.yml` matches the contract's input names and `required` settings. It does not execute a workflow or Action.

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
