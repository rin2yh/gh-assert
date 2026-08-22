# CLI contract path interface

## Context

`gh-assert` はworkflow、Composite Action、Reusable Workflowへ対象を広げる予定である。利用者に対象種別を選ばせると、contractの配置とCLI指定が二重になる。

## Decision

contract pathをCLIの位置引数として扱う。

- `gh assert`: `.github` 配下の全対応contractに対するruntime assertion。
- `gh assert <contract>`: 指定contractに対するruntime assertion。
- `gh assert validate`: `.github` 配下の全対応contractを静的検証する。
- `gh assert validate <contract>`: 指定contractだけを静的検証する。

対象種別はcontractの配置と内容から判定する。`validate` はruntime assertionと異なる操作であるため、サブコマンドとして扱う。

## Consequences

runtime assertionを既定動作として短いコマンドで実行できる。静的検証は`validate`サブコマンドで明確に区別できる。対象分類を追加してもCLIを変更せず、探索規則を拡張できる。
