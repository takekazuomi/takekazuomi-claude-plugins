#!/usr/bin/env bash
# wordcount.sh — 日本語文章の分量を測り、原稿用紙・書籍ページに換算する
#
# 使い方:
#   ./wordcount.sh path/to/article.md
#   cat article.md | ./wordcount.sh
#
# 仕様:
#   - 入力は UTF-8 とみなす
#   - 空白を除いた文字数で数える。半角スペース・タブ・改行に加え、
#     全角スペース(U+3000)・NBSP(U+00A0) も除去する
#     （「小説家になろう」「カクヨム」など一般的な文字数カウントと同じ方針）
#   - 数える単位は Unicode 符号点。結合文字（か+結合濁点）や絵文字の ZWJ 連結は
#     見た目の1文字より多く数える。NFC 正規化済みの実用文では差は出ない
#   - 400字詰め原稿用紙の枚数を出す。ただし文字数÷400 の概算であり、
#     実際の組版では改行や字下げが空きマスを消費するため、枚数はこれより多くなる
#   - 書籍ページは基準が複数あるため、単行本/新書/文庫の3基準で併記する
#   - markdownの脚注・コードブロックも本文に含めて数える（必要なら事前に除去すること）
#
# 実装メモ:
#   `wc -m` はロケールが C / POSIX のときバイト数を返す。UTF-8 の日本語は
#   1文字3バイトのため、黙って約3倍に数えてしまう。誤った数値は無いより
#   有害なので、UTF-8 ロケールを確保し、確保できなければエラーで停止する。
#
#   換算は bash の整数演算で行い、bc に依存しない。bc は最小構成の
#   コンテナに無いことがあり、欠けると 0.0 という誤った数値を表示していた。
#   外部コマンドは sed・tr・wc・grep・locale のみ。

set -euo pipefail

die() {
    echo "Error at line ${BASH_LINENO[0]}: $*" >&2
    exit 1
}

# wc -m を正しく動かすため UTF-8 ロケールを確保する
ensure_utf8_locale() {
    local candidate
    if [ "$(locale charmap 2>/dev/null)" = "UTF-8" ]; then
        return
    fi
    for candidate in C.UTF-8 C.utf8 en_US.UTF-8 ja_JP.UTF-8; do
        if locale -a 2>/dev/null | grep -qxiF "$candidate"; then
            export LC_ALL="$candidate"
            return
        fi
    done
    die "UTF-8 ロケールが必要（現在: $(locale charmap 2>/dev/null)）。LC_ALL に UTF-8 ロケールを指定すること"
}

# 空白を除いた文字数を標準入力から数えて出力する
count_chars() {
    # tr はバイト単位で動くため [:space:] では ASCII 空白しか落とせない。
    # 全角スペースと NBSP は sed のリテラル置換で先に除く
    # （エスケープ表記にしてソース上に不可視文字を置かない）
    local zenkaku_space=$'\xe3\x80\x80' nbsp=$'\xc2\xa0'
    sed "s/${zenkaku_space}//g; s/${nbsp}//g" | tr -d '[:space:]' | wc -m | tr -d ' '
}

# 小数第1位まで四捨五入して出力する（bash の整数演算のみ）
div1() {
    local num="$1" den="$2" tenths
    tenths=$(( (num * 10 + den / 2) / den ))
    printf '%d.%d' "$(( tenths / 10 ))" "$(( tenths % 10 ))"
}

# 整数に四捨五入して出力する
div0() {
    local num="$1" den="$2"
    printf '%d' "$(( (num + den / 2) / den ))"
}

main() {
    local src text chars

    if [ "$#" -ge 1 ]; then
        [ -f "$1" ] || die "ファイルが見つからない: $1"
        src="$1"
        text="$(cat -- "$src")"
    else
        src="(stdin)"
        text="$(cat)"
    fi

    ensure_utf8_locale
    chars="$(printf '%s' "$text" | count_chars)"

    echo "対象: $src"
    echo "文字数（空白・改行除く）: ${chars}字"
    echo ""
    echo "=== 400字詰め原稿用紙 ==="
    echo "枚数: $(div1 "$chars" 400) 枚（概算。改行・字下げの空きマスは含まない）"
    echo ""
    echo "=== 書籍ページ換算（基準別） ==="
    echo "単行本相当(600字/頁): $(div1 "$chars" 600) 頁"
    echo "新書相当  (680字/頁): $(div1 "$chars" 680) 頁"
    echo "文庫相当  (700字/頁): $(div1 "$chars" 700) 頁"
    echo ""
    echo "=== 黙読の目安 ==="
    echo "分速500字で約 $(div0 "$chars" 500) 分"
}

main "$@"
