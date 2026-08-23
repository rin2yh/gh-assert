# Validate and assert reusable workflow inputs

## Context

Reusable Workflowが受け取る値は`workflow_call.inputs`で宣言する。呼び出されたWorkflow内では`${{ inputs.<name> }}`として参照できるが、その中から呼び出すComposite Actionには自動で渡されない。また、`GITHUB_EVENT_PATH`には呼び出し元のevent payloadが入り、Reusable Workflowへ渡されたinput値は含まれない。

input名、`required`、`type`はWorkflow YAMLとcontractを比較できる。一方、`enum`、`pattern`、`min`、`max`は実際に渡された値がなければ検証できない。

## Decision

`<name>_assert.yml`と同じ場所にある`<name>.yml`が`workflow_call`を宣言している場合、gh-assertは`workflow_call.inputs`とcontractの`inputs`を静的に比較する。比較するのはinput名、`required`、`type`である。contractの`integer`はGitHub Actionsの`number`に対応させる。

runtime assertionでは、Reusable Workflow内からActionの`workflow-inputs`へ`${{ toJSON(inputs) }}`を渡す。Actionはこの値を`GH_ASSERT_INPUTS`としてCLIへ渡し、値のtypeと制約を検証する。Reusable Workflowのinput contractがあるのに値が渡されなかった場合は、検証をskipせずエラーにする。

通常の`workflow_dispatch`では、従来どおりevent payloadからinput値を取得する。

## Consequences

公開interfaceとcontractの不一致に加え、Reusable Workflowへ実際に渡された値の違反も検出できる。利用者はReusable Workflow内のgh-assert Actionへ、input contextを1項目明示的に渡す必要がある。
