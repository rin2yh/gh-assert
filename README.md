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
- uses: rin2yh/gh-assert@v1
  with:
    contract: .github/workflows/deploy_assert.yml
  env:
    ENVIRONMENT: ${{ vars.DEPLOY_ENVIRONMENT }}
    RETRIES: ${{ vars.DEPLOY_RETRIES }}
    DRY_RUN: ${{ vars.DRY_RUN }}
```

The Action installs the matching release binary through GitHub CLI, then runs it. It exits non-zero when a required variable is missing or empty, a value has the wrong type, or a declared constraint fails. Values are not printed in diagnostics.

## Validate a contract

Install the extension:

```bash
gh extension install rin2yh/gh-assert
```

Validate a contract definition:

```bash
gh assert validate .github/workflows/deploy_assert.yml
```

The same command is available as `gh-assert validate ...` after building locally. Validation checks YAML syntax, supported fields and types, regular expressions, and integer ranges. It does not execute a workflow.

## Development

```bash
go test ./...
go vet ./...
```

The project is released as a precompiled GitHub CLI Extension. Releases are tagged and published by tagpr. The repository is available at `github.com/rin2yh/gh-assert`.

## License

MIT.
