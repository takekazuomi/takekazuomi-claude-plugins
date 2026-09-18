#!/usr/bin/env bash
#
# check-registration.sh はスキルの登録の整合を検査する。
# CLAUDE.md「スキルの追加方法」が求める登録箇所を、plugins/ 配下のディレクトリと突き合わせる。
#   - plugins/<name>/.claude-plugin/plugin.json（name・description・version）
#   - plugins/<name>/skills/<name>/SKILL.md（frontmatter の name・description）
#   - .claude-plugin/marketplace.json（name・source・description）
#   - README.md（スキル一覧の表と構造ツリー）
#   - mise.toml の SKILLS
#
# 使い方: mise run test:static（SKILLS は mise.toml の [env] から渡る）
set -euo pipefail

readonly MARKETPLACE=".claude-plugin/marketplace.json"
readonly README="README.md"

errors=0

die() {
    echo "Error at line ${BASH_LINENO[0]}: $*" >&2
    exit 1
}

# 不整合を記録し、検査は続ける
fail() {
    echo "NG: $*" >&2
    errors=$((errors + 1))
}

# frontmatter（1 行目の --- から次の --- まで）から key の値を取り出す
frontmatter_value() {
    local file="$1" key="$2"
    awk -v key="${key}" '
        NR == 1 && $0 != "---" { exit }
        NR > 1 && $0 == "---" { exit }
        NR > 1 && index($0, key ":") == 1 {
            sub("^" key ":[[:space:]]*", "")
            print
            exit
        }
    ' "${file}"
}

# SKILLS に name が含まれるか
in_skills() {
    local name="$1" skill
    for skill in ${SKILLS}; do
        [[ "${skill}" == "${name}" ]] && return 0
    done
    return 1
}

# README のスキル一覧の表（| で始まる行）にある SKILL.md へのリンクから名前を取り出す
readme_table_names() {
    awk '
        index($0, "|") == 1 {
            while (match($0, /\]\(\.\/plugins\/[^\/]+\//)) {
                name = substr($0, RSTART + 12, RLENGTH - 13)
                print name
                $0 = substr($0, RSTART + RLENGTH)
            }
        }
    ' "${README}" | sort -u
}

# README の「構造」節のツリーに plugins/<name>/ の行があるか
in_readme_tree() {
    local name="$1"
    awk -v name="${name}" '
        /^## / { in_section = ($0 == "## 構造") }
        in_section && index($0, "── " name "/") { found = 1 }
        END { exit !found }
    ' "${README}"
}

check_manifest() {
    local name="$1"
    local manifest="plugins/${name}/.claude-plugin/plugin.json"

    if [[ ! -f "${manifest}" ]]; then
        fail "${name}: ${manifest} がない"
        return
    fi
    if ! jq empty "${manifest}" 2>/dev/null; then
        fail "${name}: ${manifest} が JSON として読めない"
        return
    fi
    [[ "$(jq -r '.name // ""' "${manifest}")" == "${name}" ]] ||
        fail "${name}: plugin.json の name がディレクトリ名と異なる"
    [[ -n "$(jq -r '.description // ""' "${manifest}")" ]] ||
        fail "${name}: plugin.json に description がない"
    [[ -n "$(jq -r '.version // ""' "${manifest}")" ]] ||
        fail "${name}: plugin.json に version がない"
}

check_skill() {
    local name="$1"
    local skill="plugins/${name}/skills/${name}/SKILL.md"

    if [[ ! -f "${skill}" ]]; then
        fail "${name}: ${skill} がない"
        return
    fi
    [[ "$(frontmatter_value "${skill}" name)" == "${name}" ]] ||
        fail "${name}: SKILL.md の frontmatter の name がディレクトリ名と異なる"
    [[ -n "$(frontmatter_value "${skill}" description)" ]] ||
        fail "${name}: SKILL.md の frontmatter に description がない"
}

check_marketplace() {
    local name="$1" count
    count="$(jq --arg n "${name}" '[.plugins[] | select(.name == $n)] | length' "${MARKETPLACE}")"

    if [[ "${count}" -eq 0 ]]; then
        fail "${name}: marketplace.json に登録がない"
        return
    fi
    if [[ "${count}" -gt 1 ]]; then
        fail "${name}: marketplace.json に ${count} 件登録されている"
        return
    fi
    [[ "$(jq -r --arg n "${name}" '.plugins[] | select(.name == $n) | .source // ""' "${MARKETPLACE}")" == "./plugins/${name}" ]] ||
        fail "${name}: marketplace.json の source が ./plugins/${name} ではない"
    [[ -n "$(jq -r --arg n "${name}" '.plugins[] | select(.name == $n) | .description // ""' "${MARKETPLACE}")" ]] ||
        fail "${name}: marketplace.json に description がない"
}

check_readme() {
    local name="$1"

    readme_table_names | grep -qxF "${name}" ||
        fail "${name}: README.md のスキル一覧の表にない"
    in_readme_tree "${name}" ||
        fail "${name}: README.md の構造ツリーにない"
}

main() {
    command -v jq >/dev/null 2>&1 || die "jq が見つからない。mise run test:static から実行する"
    [[ -n "${SKILLS:-}" ]] || die "SKILLS が未設定。mise run test:static から実行する"

    cd "$(dirname "${BASH_SOURCE[0]}")/.."
    jq empty "${MARKETPLACE}" 2>/dev/null || die "${MARKETPLACE} が JSON として読めない"

    local -a names=()
    local dir name
    shopt -s nullglob
    for dir in plugins/*/; do
        name="$(basename "${dir}")"
        names+=("${name}")
        check_manifest "${name}"
        check_skill "${name}"
        check_marketplace "${name}"
        check_readme "${name}"
        in_skills "${name}" || fail "${name}: mise.toml の SKILLS にない"
    done
    shopt -u nullglob
    [[ ${#names[@]} -gt 0 ]] || die "plugins/ にプラグインがない"

    # 逆方向: ディレクトリのない名前が登録に残っていないか
    local registered
    while IFS= read -r registered; do
        [[ -d "plugins/${registered}" ]] || fail "${registered}: marketplace.json に登録があるが plugins/${registered} がない"
    done < <(jq -r '.plugins[].name' "${MARKETPLACE}")
    for registered in ${SKILLS}; do
        [[ -d "plugins/${registered}" ]] || fail "${registered}: mise.toml の SKILLS にあるが plugins/${registered} がない"
    done
    while IFS= read -r registered; do
        [[ -d "plugins/${registered}" ]] || fail "${registered}: README.md の表にあるが plugins/${registered} がない"
    done < <(readme_table_names)

    if [[ "${errors}" -gt 0 ]]; then
        die "登録の不整合が ${errors} 件"
    fi
}

main "$@"
