# Separate assert file

## Context

workflow定義へ検証ルールを混在させると、実行設定と契約の境界が曖昧になる。契約だけを確認できるファイルも必要である。

## Decision

workflowとは別に`*_assert.yml`を配置し、env contractを記述する。`validate`とruntime Actionはこのファイルを共通のsource of truthとして扱う。

## Consequences

workflowの実行設定を変更せずに契約をレビューできる。requiredなenvがActionの実行環境に存在しない場合、runtime assertionで不足を検出する。
