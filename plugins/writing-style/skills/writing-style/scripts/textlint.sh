#!/usr/bin/env bash
# textlint.sh — 日本語文書を textlint で機械検査する
#
# 使い方:
#   ./textlint.sh formal path/to/design-doc.md
#   ./textlint.sh casual path/to/article.md
#   ./textlint.sh formal --fix path/to/draft.md      # 3番目以降は textlint にそのまま渡す
#
# 仕様:
#   - スタイル別の設定は同ディレクトリ階層の ../textlint/<style>.json を使う
#   - textlint 本体と ルールパッケージは「検査対象プロジェクト側」に入れる。
#     このスキルは設定ファイルだけを配る（node_modules は同梱しない）
#   - textlint が見つからない場合は導入コマンドを案内し、終了コード0で抜ける。
#     機械検査は任意の工程であり、レビュー手順全体を止めない
#
# 検査対象プロジェクトでの導入:
#   mise use node@24    # node が無い場合。mise が nodejs を用意する
#   npm i -D textlint textlint-rule-preset-ja-technical-writing textlint-rule-no-kangxi-radicals
#
# 注意:
#   textlint はルールをカレントディレクトリの node_modules から解決する。
#   設定ファイルがプロジェクト外にあっても解決できる（検証済み）が、
#   実行はかならず検査対象プロジェクトのルートで行うこと。

set -euo pipefail

die() {
    echo "Error at line ${BASH_LINENO[0]}: $*" >&2
    exit 1
}

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly CONFIG_DIR="${SCRIPT_DIR}/../textlint"
readonly NODE_VERSION="24"
readonly NPM_PACKAGES="textlint textlint-rule-preset-ja-technical-writing textlint-rule-no-kangxi-radicals"

# textlint の実体を探す。プロジェクトローカルを優先する。
# textlint は node スクリプトなので、node が無ければ実体があっても動かない。
# 実際に --version を走らせて、起動できるものだけを採用する
runnable() {
    [ -x "$1" ] && "$1" --version >/dev/null 2>&1
}

find_textlint() {
    local local_bin="./node_modules/.bin/textlint" path_bin
    if runnable "$local_bin"; then
        echo "$local_bin"
        return
    fi
    path_bin="$(command -v textlint 2>/dev/null || true)"
    if [ -n "$path_bin" ] && runnable "$path_bin"; then
        echo "$path_bin"
    fi
}

usage() {
    echo "Usage: $0 <casual|formal> <file>... [textlint のオプション]" >&2
}

# 何が足りないかを見て、導入手順を提案する
suggest_install() {
    echo "textlint を実行できないため機械検査をスキップする。" >&2
    echo "" >&2
    echo "検査したい場合は、対象プロジェクトのルートで次を実行すること:" >&2

    if ! command -v node >/dev/null 2>&1; then
        if command -v mise >/dev/null 2>&1; then
            echo "  mise use node@${NODE_VERSION}" >&2
        else
            echo "  # node が無い。mise の導入を推奨する: https://mise.jdx.dev/" >&2
            echo "  mise use node@${NODE_VERSION}" >&2
        fi
    fi
    echo "  npm i -D ${NPM_PACKAGES}" >&2
}

main() {
    local style config textlint_bin

    style="${1:-}"
    case "$style" in
        casual | formal) shift ;;
        "")
            usage
            die "スタイルを指定すること"
            ;;
        *)
            usage
            die "未知のスタイル: ${style}（casual か formal）"
            ;;
    esac

    [ "$#" -ge 1 ] || { usage; die "検査対象のファイルを指定すること"; }

    config="${CONFIG_DIR}/${style}.json"
    [ -f "$config" ] || die "設定ファイルが見つからない: ${config}"

    textlint_bin="$(find_textlint)"
    if [ -z "$textlint_bin" ]; then
        suggest_install
        exit 0
    fi

    "$textlint_bin" --config "$config" "$@"
}

main "$@"
