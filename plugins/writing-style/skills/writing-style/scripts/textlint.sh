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
#   - このスキルは設定ファイルだけを配る。textlint 本体とルールは導入しない・導入を案内しない
#     （利用者の環境を変えないため。mise も前提にしない）
#   - 利用者の環境にある textlint を使う。プロジェクトローカル（./node_modules/.bin）を優先する
#   - 次の場合は警告を出し、同梱のルール設定を参考として案内する。機械検査は任意の工程であり、
#     レビュー手順全体を止めないため、検査できなかった場合も終了コード 0 で抜ける
#       - textlint が無い（または node が無く起動できない）: 検査をスキップする
#       - 本体のメジャー版が想定（TEXTLINT_MAJOR）と違う: 警告したうえで検査する
#       - ルールを読み込めない: 検査をスキップする
#
# 注意:
#   textlint はルールを本体の設置場所から解決する（カレントディレクトリではない）。
#   本体とルールが同じ node_modules に無いと "No rules found" になる。

set -euo pipefail

die() {
    echo "Error at line ${BASH_LINENO[0]}: $*" >&2
    exit 1
}

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
CONFIG_DIR="$(cd -- "${SCRIPT_DIR}/../textlint" && pwd)"
readonly CONFIG_DIR
# 開発側（リポジトリの package.json）で固定している textlint のメジャー版
readonly TEXTLINT_MAJOR="15"

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

# 同梱のルール設定の所在を案内する
show_bundled_config() {
    echo "このスキルには textlint のルール設定が組み込まれている。導入や設定の参考にすること:" >&2
    echo "  ${CONFIG_DIR}/casual.json" >&2
    echo "  ${CONFIG_DIR}/formal.json" >&2
    echo "  必要なルール: textlint-rule-preset-ja-technical-writing・textlint-rule-no-kangxi-radicals（textlint 本体と同じ node_modules に置く）" >&2
}

# 本体のメジャー版が想定と違えば警告する（検査は続ける）
warn_version() {
    local version major
    version="$("$1" --version 2>/dev/null || true)"
    version="${version#v}"
    major="${version%%.*}"
    if [ "$major" != "$TEXTLINT_MAJOR" ]; then
        echo "警告: textlint ${version:-不明} を検出した。想定はメジャー版 ${TEXTLINT_MAJOR}。結果が想定と異なる場合がある。" >&2
        show_bundled_config
        echo "" >&2
    fi
}

main() {
    local style config textlint_bin output rc

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
        echo "警告: textlint を実行できない（未導入、または node が無い）ため機械検査をスキップする。" >&2
        show_bundled_config
        exit 0
    fi

    warn_version "$textlint_bin"

    # 指摘ありとルール読み込み失敗はどちらも非 0 で返るため、出力で見分ける
    rc=0
    output="$("$textlint_bin" --config "$config" "$@" 2>&1)" || rc=$?
    if [ "$rc" -ne 0 ] && printf '%s\n' "$output" | grep -qE "No rules found|Failed to load textlint's module"; then
        echo "警告: textlint のルールを読み込めないため機械検査をスキップする。" >&2
        show_bundled_config
        exit 0
    fi
    [ -z "$output" ] || printf '%s\n' "$output"
    exit "$rc"
}

main "$@"
