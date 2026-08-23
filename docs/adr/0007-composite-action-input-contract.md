# Validate Composite Action inputs against sibling contracts

## Context

Composite Actionは`action.yml`の`inputs`で公開interfaceを定義する。runtimeでは`inputs` contextを利用できるが、Composite Actionから別のComposite Actionへ自動的には渡されない。また、Action metadataには呼び出し元から受け取るenvの公開interfaceを宣言する仕組みがない。

Action inputsはGitHub Actions上では文字列として扱われる。一方、gh-assertのcontractでは文字列表現に対して`integer`や`boolean`を含む意味上の型と制約を検証できる。

## Decision

`.github/actions/<name>/action.yml`の隣に`action_assert.yml`を置く。`validate`は`action.yml`がComposite Actionの場合に、input名と`required`をcontractと比較する。contractの型はAction metadataとは比較せず、runtime値へ適用する。

runtime assertionでは`action-inputs: ${{ toJSON(inputs) }}`によってAction inputsを明示的にgh-assertへ渡す。envはcontractに定義し、runtime assertionだけを行う。

GitHub Actions由来のAction構造とYAML parserは`internal/github`、Actionとcontractを比較する処理は`internal/composite`へ置く。

## Consequences

Composite Actionの公開inputに対するcontractの不足やずれを実行前に検出できる。値の制約とenvは実行時に検証できる。

Action metadataにenv interfaceがないため、env名の静的な完全性は検証しない。Action inputsの`integer`と`boolean`もmetadata上の型ではなく、runtime文字列に対する意味上の型となる。
