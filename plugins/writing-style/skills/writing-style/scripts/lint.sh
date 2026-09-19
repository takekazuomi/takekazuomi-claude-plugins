#!/usr/bin/env bash
# lint.sh — 原稿からAI由来の型を機械的に検出する
#
# 使い方:
#   bash lint.sh path/to/draft.md
#
# 仕様:
#   - コードブロック（``` で囲まれた範囲）とインラインコード（`...`）は検査対象から外す
#   - <!-- lint-off --> から <!-- lint-on --> までの範囲も外す。
#     禁止語そのものを列挙している箇所（このスキル自身の説明など）で使う
#   - 行番号は元ファイルのものを保つ
#   - ERROR は落とすべきもの、WARN は誤検出があり得るので目視確認するもの
#   - TODO(...) の残存は数えるだけで、失敗にはしない
#   - 文字数を数える処理は awk ではなく bash で行う。awk 実装によってはマルチバイトを
#     文字単位で扱えず、日本語で誤動作するため
#
# 終了コード:
#   0 = ERROR/WARN なし
#   1 = ERROR または WARN あり
#   2 = 引数エラー

set -euo pipefail

UTF8_OK=0
# 一時ファイル。EXIT の trap は main を抜けたあとに動くため、local にすると set -u で落ちる
STRIPPED=""
FINDINGS=""

cleanup() {
    rm -f "${STRIPPED}" "${FINDINGS}"
}

die() {
    echo "Error at line ${BASH_LINENO[0]}: $*" >&2
    exit 2
}

# UTF-8 ロケールを探して設定する。名前の表記ゆれ（C.UTF-8 / C.utf8）を吸収する
setup_locale() {
    local avail loc norm
    avail="$(locale -a 2>/dev/null || true)"

    for norm in ja_jp.utf8 en_us.utf8 c.utf8; do
        while IFS= read -r loc; do
            [[ -z "${loc}" ]] && continue
            if [[ "$(printf '%s' "${loc}" | tr '[:upper:]' '[:lower:]' | tr -d '-')" == "${norm}" ]]; then
                export LC_ALL="${loc}"
                UTF8_OK=1
                return 0
            fi
        done <<< "${avail}"
    done

    while IFS= read -r loc; do
        [[ -z "${loc}" ]] && continue
        if [[ "$(printf '%s' "${loc}" | tr '[:upper:]' '[:lower:]' | tr -d '-')" == *utf8 ]]; then
            export LC_ALL="${loc}"
            UTF8_OK=1
            return 0
        fi
    done <<< "${avail}"

    echo "Warning: UTF-8 ロケールが見つからないため、文字数に依存する検査を省略します" >&2
    return 0
}

# コードブロック・インラインコード・除外区間を空にする。行数は変えない
strip_code() {
    awk '
        /<!--[[:space:]]*lint-off[[:space:]]*-->/ { inskip = 1; print ""; next }
        /<!--[[:space:]]*lint-on[[:space:]]*-->/  { inskip = 0; print ""; next }
        inskip { print ""; next }
        /^[[:space:]]*```/ { infence = !infence; print ""; next }
        infence            { print ""; next }
        {
            gsub(/`[^`]*`/, "")
            print
        }
    '
}

# 長い抜粋を安全に切り詰める。UTF-8 が使えない環境では切らない
clip() {
    local s="$1"
    if [[ "${UTF8_OK}" -eq 1 && "${#s}" -gt 60 ]]; then
        printf '%s...' "${s:0:60}"
    else
        printf '%s' "${s}"
    fi
}

# grep で拾って findings に追記する
scan() {
    local level="$1" label="$2" pattern="$3" stripped="$4" out="$5"
    local hit lineno body
    while IFS= read -r hit; do
        [[ -z "${hit}" ]] && continue
        lineno="${hit%%:*}"
        body="${hit#*:}"
        body="${body#"${body%%[![:space:]]*}"}"
        printf '%s:%s [%s] %s | %s\n' \
            "${SRC}" "${lineno}" "${level}" "${label}" "$(clip "${body}")" >> "${out}"
    done < <(grep -nE "${pattern}" "${stripped}" || true)
}

# 段落の最終文が、重みを演出する型に一致する場合に警告する。
# 長さだけで判定すると普通の短文を大量に拾うため、型のパターンと併用する
CLOSING_PATTERN='^(ここ|これ|それ|そこ)(が|こそ)[^。]{0,20}$|(一番|最も|本質|核心|決定的|肝|要)[^。]{0,15}(だ|である|になる)$|(に尽きる|しかない|そのものだ|そのものである)$'

# 段落として扱わない行（箇条書き・引用・表・見出し・番号付きリスト）。
# 数字は「1. 」の形だけを除外し、年号などで始まる段落は検査対象に残す
NON_PROSE_LINE='^[[:space:]]*([-*>|#]|[0-9]+\.[[:space:]])'

scan_closing_sentence() {
    local stripped="$1" out="$2"
    local lineno=0 lasttext="" lastline=0 line tail trimmed

    [[ "${UTF8_OK}" -eq 1 ]] || return 0

    flush_paragraph() {
        [[ "${lastline}" -eq 0 ]] && return 0
        trimmed="${lasttext%。}"
        tail="${trimmed##*。}"
        tail="${tail#"${tail%%[![:space:]]*}"}"
        tail="${tail%"${tail##*[![:space:]]}"}"
        if [[ -n "${tail}" && "${#tail}" -lt 30 ]] \
            && printf '%s' "${tail}" | grep -qE "${CLOSING_PATTERN}"; then
            printf '%s:%s [WARN] 段落末の一行断定 | %s\n' \
                "${SRC}" "${lastline}" "${tail}" >> "${out}"
        fi
        lastline=0
        lasttext=""
    }

    while IFS= read -r line; do
        lineno=$((lineno + 1))
        if [[ -z "${line//[[:space:]]/}" ]]; then
            flush_paragraph
            continue
        fi
        if [[ "${line}" =~ ${NON_PROSE_LINE} ]]; then
            lastline=0
            lasttext=""
            continue
        fi
        lastline="${lineno}"
        lasttext="${line}"
    done < "${stripped}"
    flush_paragraph
}

# 箇条書きの比率を出す（判定はせず数値だけ報告する）
report_bullet_ratio() {
    local stripped="$1"
    awk '
        /^[[:space:]]*$/ { next }
        { total++ }
        /^[[:space:]]*([-*+]|[0-9]+\.)[[:space:]]/ { bullet++ }
        END {
            if (total == 0) { print "  本文行なし"; exit }
            printf "  箇条書き %d行 / 本文 %d行 = %.0f%%\n", bullet, total, bullet * 100 / total
        }
    ' "${stripped}"
}

count_matches() {
    local pattern="$1" file="$2" n
    n="$(grep -cE "${pattern}" "${file}" 2>/dev/null || true)"
    printf '%s' "${n:-0}"
}

main() {
    local src="${1:-}"

    if [[ -z "${src}" ]]; then
        die "Usage: $0 <draft.md>"
    fi
    if [[ ! -f "${src}" ]]; then
        die "File not found: ${src}"
    fi

    setup_locale
    SRC="${src}"

    local stripped findings
    trap 'cleanup' EXIT
    STRIPPED="$(mktemp)"
    FINDINGS="$(mktemp)"
    stripped="${STRIPPED}"
    findings="${FINDINGS}"

    strip_code < "${src}" > "${stripped}"
    : > "${findings}"

    # --- ERROR: 落とすべき型 ---
    scan ERROR "ダッシュ記法" \
        '—|–' "${stripped}" "${findings}"
    scan ERROR "根拠のない強調語" \
        'まさに|非常に|極めて|きわめて|シームレス|革新的|画期的|を実現する|圧倒的に' \
        "${stripped}" "${findings}"
    scan ERROR "定型の結び" \
        'と言えるでしょう|といえるでしょう|と言えるだろう|ではないでしょうか|に他ならない|にほかならない' \
        "${stripped}" "${findings}"
    scan ERROR "メタ説明" \
        '重要なのは.*という点|ポイントは.*という点|大切なのは.*という点' \
        "${stripped}" "${findings}"
    scan ERROR "種明かし構文" \
        'の正体は|とは要するに|の本質は' \
        "${stripped}" "${findings}"

    # --- WARN: 目視確認するもの ---
    scan WARN "対比構文" \
        'ではなく、.{1,30}(だ|である|になる|にある)' "${stripped}" "${findings}"
    scan WARN "空語" \
        'することが可能|を行うことができ|といったことが|という点において' \
        "${stripped}" "${findings}"
    # 箇条書きの行末に句読点を付けない規定は formal にしかないため WARN にする
    scan WARN "箇条書きの行末の句読点" \
        '^[[:space:]]*([-*+]|[0-9]+\.)[[:space:]].*[。、]$' "${stripped}" "${findings}"
    scan_closing_sentence "${stripped}" "${findings}"

    local errors warnings todo_count
    errors="$(count_matches '\[ERROR\]' "${findings}")"
    warnings="$(count_matches '\[WARN\]' "${findings}")"
    todo_count="$(count_matches 'TODO\(' "${src}")"

    if [[ "${errors}" -gt 0 || "${warnings}" -gt 0 ]]; then
        sort -t: -k2 -n "${findings}"
        echo ""
    fi

    echo "=== 集計 ==="
    echo "  ERROR: ${errors} 件（落とす）"
    echo "  WARN : ${warnings} 件（目視確認）"
    echo "  TODO : ${todo_count} 件（未確認箇所。残っていてよい）"
    echo ""
    echo "=== 参考 ==="
    report_bullet_ratio "${stripped}"
    echo "  比率の目安は styles/casual.md を参照。モデル世代で適正値が動くため判定はしない"

    if [[ "${errors}" -gt 0 || "${warnings}" -gt 0 ]]; then
        return 1
    fi
    return 0
}

main "$@"
