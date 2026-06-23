package workspace

import (
	"fmt"
	"regexp"
)

var (
	reLabel = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)       // DNS ラベル
	reIdent = regexp.MustCompile(`^[A-Za-z0-9.\-]+(/[A-Za-z0-9._\-]+)+$`) // host/user/repo もしくは user/repo
	reURL   = regexp.MustCompile(`^(https?|git|ssh)://|^git@`)            // git URL
)

// Validate は workspace を検証し、エラーと警告の文字列スライスを返す。
//
// 標準 JSON Schema では書けない参照整合が中心:
//   - repos[].role は roles[].id のいずれか（中核のロール検証）
//   - roles[].id / repos[].name は一意
//   - relationships の端点は repos[].name に存在
//   - skills[].roles は roles[].id の部分集合
//
// revision は追跡ブランチの指定であり、無指定・ブランチ名は正常（再現性は責務外）。
// このため revision に関する警告は出さない。
func Validate(ws *Workspace) (errs, warns []string) {
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

	// repos: name 一意・role∈roles・識別子形式
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

		// 中核のロール検証: 宣言された roles のいずれかでなければならない
		if !roleSet[p.Role] {
			addErr("repos[%s]: role %q is not declared in roles", label, p.Role)
		}
		usedRole[p.Role] = true

		if !reIdent.MatchString(p.Repo) && !reURL.MatchString(p.Repo) {
			addErr("repos[%s]: repo %q is not a ghq-resolvable identity (host/user/repo) or a git URL", label, p.Repo)
		}
	}

	// relationships: 端点は repos[].name に存在
	for i, rel := range ws.Relationships {
		if !nameSet[rel.From] {
			addErr("relationships[%d].from: %q is not a known repo name", i, rel.From)
		}
		if !nameSet[rel.To] {
			addErr("relationships[%d].to: %q is not a known repo name", i, rel.To)
		}
	}

	// skills: roles は宣言された roles の部分集合
	for i, s := range ws.Skills {
		for _, role := range s.Roles {
			if !roleSet[role] {
				addErr("skills[%d] (%s): role %q is not declared in roles", i, s.When, role)
			}
		}
	}

	// 宣言されたが未使用のロール → 警告（語彙の健全性）
	for id := range roleSet {
		if !usedRole[id] {
			addWarn("roles: %q is declared but used by no repo", id)
		}
	}

	return errs, warns
}
