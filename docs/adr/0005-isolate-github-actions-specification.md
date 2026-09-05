# Isolate GitHub Actions specifications

## Context

Workflow、event、`workflow_call.inputs`、`GITHUB_EVENT_NAME`、`GITHUB_EVENT_PATH`はGitHub Actionsが定義する仕様であり、gh-assert独自の概念ではない。これらがparser、runtime assertion、Reusable Workflow検証へ分散すると、外部仕様への依存範囲が不明確になる。

## Decision

Workflow YAMLはactionlintでparseし、`actionlint.Workflow` ASTをそのまま利用する。gh-assert独自のWorkflow model、parser、adapter、別構造への変換は持たない。依存するactionlintのversionは`go.mod`で固定する。

Reusable Workflow検証とruntime判定はactionlint ASTを直接参照する。gh-assert独自のcontract YAMLは`internal/contract`、Action metadataは`internal/github`、event payloadの読み取りは`internal/runtime`が担当する。

## Consequences

Workflow構文の網羅的な解析をactionlintへ任せられ、gh-assert内でGitHub Actions仕様を重複実装せずに済む。gh-assertはactionlint ASTとcontractを直接比較するため、中間modelとの同期も不要になる。

## References

* [Workflow syntax for GitHub Actions](https://docs.github.com/actions/using-workflows/workflow-syntax-for-github-actions)
* [Variables reference](https://docs.github.com/en/actions/reference/workflows-and-actions/variables)
* [rhysd/actionlint](https://github.com/rhysd/actionlint)
