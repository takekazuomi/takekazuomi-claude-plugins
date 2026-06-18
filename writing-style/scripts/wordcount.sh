#!/usr/bin/env bash
# wordcount.sh — 日本語文章の分量を測り、原稿用紙・書籍ページに換算する
#
# 使い方:
#   ./wordcount.sh path/to/article.md
#   cat article.md | ./wordcount.sh
#
# 仕様:
#   - 空白・改行を除いた純粋な文字数で数える
#   - 400字詰め原稿用紙の枚数を出す
#   - 書籍ページは基準が複数あるため、単行本/新書/文庫の3基準で併記する
#   - markdownの脚注・コードブロックも本文に含めて数える（必要なら事前に除去すること）

set -euo pipefail

if [ "$#" -ge 1 ] && [ -f "$1" ]; then
  src="$1"
  text="$(cat "$src")"
else
  text="$(cat)"   # 標準入力
  src="(stdin)"
fi

# 空白・改行・タブをすべて除いた文字数
chars=$(printf '%s' "$text" | tr -d '[:space:]' | wc -m | tr -d ' ')

echo "対象: $src"
echo "文字数（空白・改行除く）: ${chars}字"
echo ""
echo "=== 400字詰め原稿用紙 ==="
printf "枚数: %.1f 枚\n" "$(echo "scale=4; $chars / 400" | bc)"
echo ""
echo "=== 書籍ページ換算（基準別） ==="
printf "単行本相当(600字/頁): %.1f 頁\n" "$(echo "scale=4; $chars / 600" | bc)"
printf "新書相当  (680字/頁): %.1f 頁\n" "$(echo "scale=4; $chars / 680" | bc)"
printf "文庫相当  (700字/頁): %.1f 頁\n" "$(echo "scale=4; $chars / 700" | bc)"
echo ""
echo "=== 黙読の目安 ==="
printf "分速500字で約 %.0f 分\n" "$(echo "scale=4; $chars / 500" | bc)"
