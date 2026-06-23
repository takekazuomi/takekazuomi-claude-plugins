// Command validate-workspace は workspace.yaml の妥当性を検証する。
//
// 標準 JSON Schema では表現できない「参照整合性」を中心に検証する:
//   - repos[].role は roles[].id のいずれかでなければならない（中核のロール検証）
//   - roles[].id / repos[].name は一意でなければならない
//   - relationships の from/to は repos[].name に存在しなければならない
//   - skills[].roles は roles[].id の部分集合でなければならない
// あわせて、構造の基本（必須・命名規約・識別子形式）と、再現性に関する警告も出す。
//
// エラーがあれば終了コード 1、使用法/IO/パース失敗は 2、正常は 0。
// 警告は stderr に出すが終了コードには影響しない（CI を止めない）。
//
// 依存: gopkg.in/yaml.v3
// 使い方: validate-workspace path/to/workspace.yaml
package main

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Workspace は workspace.yaml の最上位構造。
type Workspace struct {
	SchemaVersion int            `yaml:"schemaVersion"`
	Workspace     string         `yaml:"workspace"`
	Description   string         `yaml:"description"`
	Roles         []Role         `yaml:"roles"`
	Repos         []Repo         `yaml:"repos"`
	Relationships []Relationship `yaml:"relationships"`
	Skills        []SkillBinding `yaml:"skills"`
}

type Role struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
}

type Repo struct {
	Name     string `yaml:"name"`
	Repo     string `yaml:"repo"`
	Role     string `yaml:"role"`
	Revision string `yaml:"revision"`
	Optional bool   `yaml:"optional"`
}

type Relationship struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Type string `yaml:"type"`
}

type SkillBinding struct {
	When  string   `yaml:"when"`
	Use   string   `yaml:"use"`
	Roles []string `yaml:"roles"`
}

var (
	reLabel = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)        // DNS ラベル
	reSHA   = regexp.MustCompile(`^[0-9a-f]{7,40}$`)                       // 短縮〜完全SHA
	reIdent = regexp.MustCompile(`^[A-Za-z0-9.\-]+(/[A-Za-z0-9._\-]+)+$`)  // host/user/repo もしくは user/repo
	reURL   = regexp.MustCompile(`^(https?|git|ssh)://|^git@`)             // git URL
)

// validate は workspace を検証し、エラーと警告の文字列スライスを返す。
func validate(ws *Workspace) (errs, warns []string) {
	addErr := func(format string, a ...any) { errs = append(errs, fmt.Sprintf(format, a...)) }
	addWarn := func(format string, a ...any) { warns = append(warns, fmt.Sprintf(format, a...)) }

	if ws.SchemaVersion != 1 {
		addErr("schemaVersion: must be 1 (got %d)", ws.SchemaVersion)
	}
	if !reLabel.MatchString(ws.Workspace) {
		addErr("workspace: %q must be a DNS label (lowercase alphanumeric and -)", ws.Workspace)
	}

	// roles: 非空・id 一意・命名規約
	roleSet := make(map[string]bool)
	if len(ws.Roles) == 0 {
		addErr("roles: at least one role is required")
	}
	for i, r := range ws.Roles {
		if !reLabel.MatchString(r.ID) {
			addErr("roles[%d].id: %q is not a valid label", i, r.ID)
		}
		if roleSet[r.ID] {
			addErr("roles[%d].id: %q is duplicated", i, r.ID)
		}
		roleSet[r.ID] = true
	}

	// repos: name 一意・role∈roles・識別子形式・revision 固定
	nameSet := make(map[string]bool)
	usedRole := make(map[string]bool)
	if len(ws.Repos) == 0 {
		addErr("repos: at least one repo is required")
	}
	for i, p := range ws.Repos {
		label := p.Name
		if label == "" {
			label = fmt.Sprintf("#%d", i)
		}
		if !reLabel.MatchString(p.Name) {
			addErr("repos[%d].name: %q is not a valid label", i, p.Name)
		}
		if nameSet[p.Name] {
			addErr("repos[%d].name: %q is duplicated", i, p.Name)
		}
		nameSet[p.Name] = true

		// ★ 中核のロール検証: 宣言された roles のいずれかでなければならない
		if !roleSet[p.Role] {
			addErr("repos[%s]: role %q is not declared in roles", label, p.Role)
		}
		usedRole[p.Role] = true

		if !reIdent.MatchString(p.Repo) && !reURL.MatchString(p.Repo) {
			addErr("repos[%s]: repo %q is not a ghq-resolvable identity (host/user/repo) or a git URL", label, p.Repo)
		}

		switch {
		case p.Revision == "":
			addWarn("repos[%s]: no revision pin — setup is not reproducible", label)
		case !reSHA.MatchString(p.Revision):
			// タグや完全/短縮SHA以外（ブランチ名らしきもの）は west と同様に非再現的になりうる
			addWarn("repos[%s]: revision %q is not a pinned SHA — may be non-reproducible", label, p.Revision)
		}
	}

	// relationships: 端点は repos[].name に存在しなければならない
	for i, rel := range ws.Relationships {
		if !nameSet[rel.From] {
			addErr("relationships[%d].from: %q is not a known repo name", i, rel.From)
		}
		if !nameSet[rel.To] {
			addErr("relationships[%d].to: %q is not a known repo name", i, rel.To)
		}
	}

	// skills: roles は宣言された roles の部分集合でなければならない
	for i, s := range ws.Skills {
		for _, role := range s.Roles {
			if !roleSet[role] {
				addErr("skills[%d] (%s): role %q is not declared in roles", i, s.When, role)
			}
		}
	}

	// 宣言されたが未使用のロール → 警告（語彙の健全性を保つ）
	for id := range roleSet {
		if !usedRole[id] {
			addWarn("roles: %q is declared but used by no repo", id)
		}
	}

	return errs, warns
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: validate-workspace <workspace.yaml>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(2)
	}

	var ws Workspace
	if err := yaml.Unmarshal(data, &ws); err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(2)
	}

	errs, warns := validate(&ws)
	for _, w := range warns {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, "error:", e)
	}
	if len(errs) > 0 {
		os.Exit(1)
	}
	fmt.Printf("ok: workspace %q (%d repos, %d roles)\n", ws.Workspace, len(ws.Repos), len(ws.Roles))
}
