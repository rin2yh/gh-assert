# Share contract models independently from parsing

## Context

Contract、rule、type、診断位置を表す`Position`はparserだけでなく、contract validation、runtime assertion、診断、Reusable Workflow検証から参照する。これらをparserに置くと、YAMLをparseしない処理までparser packageへ依存する。

## Decision

gh-assert固有の共有構造を`internal/model`へ置く。`Contract`、`Rule`、`Type`、`Position`をmodelとし、読み込みやvalidationの処理は含めない。

Contract Parserは`model.Contract`を生成し、validationは`internal/contract`が担当する。

GitHub Actions由来のWorkflow構造はgh-assert固有のmodelではないため、`internal/model`には置かない。Workflowはactionlint ASTを直接利用する。

## Consequences

共有構造を利用するpackageがparserの実装詳細へ依存しなくなる。modelはデータ表現だけを持ち、parse、validation、assertionの責務はそれぞれのpackageに残る。
