# Static contract validation over workflow execution

## Context

GitHub Actionsの実行環境をローカルで再現すると、runner、event、権限、secretなどの扱いが複雑になる。v0.1ではenv contractの定義検証とruntime assertionに集中する。

## Decision

`gh-assert` はworkflow実行エンジンを実装しない。`*_assert.yml` を読み込み、CLIではcontractを検証し、Actionでは実行時envを検証する。

## Consequences

同じcontractを静的検証とruntime assertionで共有できる。workflow全体の実行結果や制御フローは検証対象外となる。
