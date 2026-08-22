# Workflow inputs from event payload

## Context

v0.2では `workflow_dispatch` inputsをcontractの対象に追加する。envはstepの `env:` で渡せるが、inputsは同じ方法では渡せない。利用者に `${{ toJSON(github.event.inputs) }}` の受け渡しを要求すると、contractとworkflowの二重管理になり、記述漏れがそのままassertionの抜けになる。

## Decision

runtime assertionはinputsをGitHub Actionsのevent payload (`GITHUB_EVENT_PATH`) から読み取る。値はすべて文字列として扱い、env contractと同じ制約で検証する。

inputs ruleを1つ以上持つcontractは `workflow_dispatch` の実行のみを対象とする。`GITHUB_EVENT_NAME` が `workflow_dispatch` でない場合、assertionは成功も違反も報告せず、設定エラーとして終了コード2で失敗する。ruleが無い `inputs: {}` は検証対象を持たないため、`env: {}` と同様にeventを制限しない。`workflow_call.inputs` はv0.3で扱う。

## Consequences

利用者はcontractにinputsを書くだけでよく、Actionのstepへ追加の受け渡しを書く必要がない。event payloadを読むため、workflow側の記述漏れによるassertionの抜けが起きない。inputs contractを持つcontractは `workflow_dispatch` 以外の実行では検証できず、その場合は明示的なエラーになる。
