# Isolate GitHub Actions specifications

## Context

Workflow、event、`workflow_call.inputs`、`GITHUB_EVENT_NAME`、`GITHUB_EVENT_PATH`はGitHub Actionsが定義する仕様であり、gh-assert独自の概念ではない。これらがparser、runtime assertion、Reusable Workflow検証へ分散すると、外部仕様への依存範囲が不明確になる。

## Decision

GitHub Actionsの仕様に由来する構造と処理を`internal/github`へ置く。Workflowとinputの構造、event名、event payloadの読み取りをこのpackageに含める。

YAMLの読み込みと構造体への変換は`internal/parser`が担当するが、Workflow Parserの出力は`github.Workflow`とする。`internal/reusable`はGitHub Workflowとgh-assertのcontractを比較する接続部分とし、cmdへGitHub固有の判定結果を漏らさない。

## Consequences

GitHub Actionsの仕様変更へ追従する場所が明確になる。parserはvalidationやWorkflowとcontractの比較を持たず、Reusable Workflow固有の処理も`internal/reusable`に保たれる。

## References

* [Workflow syntax for GitHub Actions](https://docs.github.com/actions/using-workflows/workflow-syntax-for-github-actions)
* [Variables reference](https://docs.github.com/en/actions/reference/workflows-and-actions/variables)
