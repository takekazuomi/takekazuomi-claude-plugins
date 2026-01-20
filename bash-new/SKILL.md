---
name: bash-new
description: 新規bashスクリプトの作成。bash best practiceに準拠したテンプレートを生成。
---

# bash-new

新規bashスクリプトを作成する。

## 使用方法

```
/bash-new <ファイル名>
```

## 引数

- `<ファイル名>`: 作成するスクリプトファイル名（例: `myscript.sh`）

## 生成テンプレート

```bash
#!/usr/bin/env bash
set -euo pipefail

die() {
    echo "Error at line ${BASH_LINENO[0]}: $*" >&2
    exit 1
}

main() {
    local arg="${1:-}"

    if [[ -z "${arg}" ]]; then
        die "Usage: $0 <argument>"
    fi

    # TODO: 処理を記述
}

main "$@"
```

## テンプレート解説

### シェバン

```bash
#!/usr/bin/env bash
```

- `/usr/bin/env` 経由でbashを起動し、環境による違いを吸収

### エラー設定

```bash
set -euo pipefail
```

- `-e`: コマンドエラー時に即座に終了
- `-u`: 未定義変数の参照でエラー
- `-o pipefail`: パイプライン中のエラーを検知

### die関数

```bash
die() {
    echo "Error at line ${BASH_LINENO[0]}: $*" >&2
    exit 1
}
```

- エラー発生時に行番号付きでメッセージを出力
- 標準エラー出力へ出力
- 終了コード1で終了

### main関数パターン

```bash
main() {
    # 処理
}

main "$@"
```

- スクリプト全体を関数で囲む
- グローバル変数の汚染を防止
- 関数内でlocal変数を使用可能

## 設計方針

- echoは失敗時のみ使用（行番号とエラーメッセージ）
- 絵文字・エスケープシーケンス不使用
- シンプルで可読性の高い構成
