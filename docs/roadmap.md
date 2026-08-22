# Roadmap

## v0.1 — Env contract and runtime assertion

`*_assert.yml` にenv contractを定義し、contractのvalidationとGitHub Actions実行時のruntime assertionまで対応する。

* `required`
* `string`

  * `enum`
  * `pattern`
* `integer`

  * `min`
  * `max`
* `boolean`
* `gh assert validate`
* runtime env validation
* GitHub Actionとして実行可能にする
* `*_assert.yml` をvalidateとruntime assertionで共有する

v0.1ではenvだけをcontractの対象とする。inputs、job/step env、Reusable Workflow、Composite Action contractは後続versionで扱う。

`validate` はcontract定義のみを検証する。

* YAML syntax
* unknown fields
* type definition
* invalid regex
* `min > max`

## v0.2 — Workflow inputs

Workflowが受け取るinputsをcontractの対象に追加する。

* `workflow_dispatch` inputs
* required / type / value constraints
* runtime assertion

## v0.3 — Reusable Workflows

Reusable Workflowをcontractの対象に追加する。

* `workflow_call.inputs`
* reusable workflowの公開interfaceに対するcontract
* runtime assertion

## v0.4 — Composite Actions

`.github/actions/` 配下のComposite Actionをcontractの対象に追加する。

```text
.github/actions/<name>/
  action.yml
  action_assert.yml
```

* Action inputs
* Actionで利用するenv
* runtime assertion

## Future

* GitHub annotation

## Non-goals

`gh-assert` は値そのものに対するassertionに集中する。

以下は対象外とする。

* job単位のcontract
* step単位のcontract
* 特定の `uses:` 呼び出しに対するcontract
* workflowの実行順序や制御フローの検証
* conditional requirements
* cross-field conditions
* 複数値間の業務ルール
* secretsの値に対するcontract / assertion
* GitHub expressionの静的解析

`gh-assert` はWorkflow / Reusable Workflow / Actionが外部から受け取る値や環境に対するcontractを扱う。
