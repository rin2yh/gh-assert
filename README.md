# gh-assert

Declarative environment assertions for GitHub Actions.

`gh-assert` moves common shell checks into a small YAML contract. Version 0.1 validates environment variables at runtime inside a GitHub Action and validates the contract definition from the command line.

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
```

Supported constraints are `required`, string `enum` and `pattern`, integer `min` and `max`, and boolean type checking. Inputs, defaults, expressions, conditional rules, and step or job contracts are outside v0.1.

## Runtime assertion

Place the Action in a workflow step and pass the environment values that the contract names:

```yaml
- uses: rin2yh/gh-assert@v0.0.2
  with:
    contract: .github/workflows/deploy_assert.yml
    version: v0.0.2
  env:
    ENVIRONMENT: ${{ vars.DEPLOY_ENVIRONMENT }}
    RETRIES: ${{ vars.DEPLOY_RETRIES }}
    DRY_RUN: ${{ vars.DRY_RUN }}
```

Pick a tag that [Releases](https://github.com/rin2yh/gh-assert/releases) already publishes; the example above names the next one. The Action runs on Linux, macOS and Windows runners on x64 and arm64. It picks the release binary for `RUNNER_OS` and `RUNNER_ARCH` and verifies it against the release's `checksums.txt` before execution. It exits non-zero when a required variable is missing or empty, a value has the wrong type, or a declared constraint fails. Values are not printed in diagnostics.

When `contract` is omitted, the Action discovers every `*_assert.yml` under `.github`.

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

Validation checks YAML syntax, supported fields and types, regular expressions, and integer ranges. It does not execute a workflow.

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
