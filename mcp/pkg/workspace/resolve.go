package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Resolver は Definition + Overlay を、ユーザーごとのローカル所在（ghq / worktree）へ解決する。
// 所在はマニフェストに持たせず、GHQ_ROOT / ghq.root と WT_ROOT から決定論的に算出する。
type Resolver struct {
	WS      *Workspace
	Overlay *Overlay
	GhqRoot string // ghq のルート（clone 先）
	WtRoot  string // worktree のルート（Overlay の差し替え先）
}

// NewResolver は ghq ルートと worktree ルートを解決して Resolver を作る。
func NewResolver(ws *Workspace, ov *Overlay) *Resolver {
	if ov == nil {
		ov = &Overlay{}
	}
	return &Resolver{WS: ws, Overlay: ov, GhqRoot: ghqRoot(), WtRoot: worktreeRoot()}
}

func ghqRoot() string {
	if r := os.Getenv("GHQ_ROOT"); r != "" { // GHQ_ROOT は他設定を上書きする
		return expandHome(r)
	}
	// --global を付けないのは、リポローカルの ghq.root も尊重するため。
	if out, err := exec.Command("git", "config", "--path", "--get-all", "ghq.root").Output(); err == nil {
		s := strings.TrimRight(string(out), "\n")
		if s != "" {
			lines := strings.Split(s, "\n")
			return expandHome(lines[len(lines)-1]) // 複数あるとき新規 clone の主ルートは最後（x-motemen/ghq）
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "ghq")
}

func worktreeRoot() string {
	if r := os.Getenv("WT_ROOT"); r != "" {
		return expandHome(r)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "wt")
}

// parseIdentity は ghq 識別子（host/user/repo、user/repo、各種 URL）から host とサブパスを取り出す。
func parseIdentity(id string) (host, subpath string) {
	s := strings.TrimSuffix(id, ".git")
	switch {
	case strings.HasPrefix(s, "git@"): // git@github.com:acme/repo
		s = strings.TrimPrefix(s, "git@")
		s = strings.Replace(s, ":", "/", 1)
	case strings.Contains(s, "://"): // https://github.com/acme/repo, ssh://git@host/...
		s = s[strings.Index(s, "://")+3:]
		if slash := strings.Index(s, "/"); slash >= 0 {
			if at := strings.Index(s[:slash], "@"); at >= 0 {
				s = s[at+1:]
			}
		}
	}
	parts := strings.Split(s, "/")
	if len(parts) >= 2 && strings.Contains(parts[0], ".") {
		return parts[0], strings.Join(parts[1:], "/") // 先頭がホスト名
	}
	return "github.com", strings.Join(parts, "/") // host 省略時は github.com
}

// OverrideFor は name に対する Overlay の上書きを返す。
func (r *Resolver) OverrideFor(name string) (Override, bool) {
	for _, o := range r.Overlay.Overrides {
		if o.Name == name {
			return o, true
		}
	}
	return Override{}, false
}

// GhqPathOf はリポの ghq 上の決定論的パスを返す（Overlay を無視。clone 先）。
func (r *Resolver) GhqPathOf(p Repo) string {
	host, subpath := parseIdentity(p.Repo)
	return filepath.Join(r.GhqRoot, host, filepath.FromSlash(subpath))
}

// PathOf はリポの参照先パスを返す。Overlay で差し替えられていれば worktree パス。
// worktree は WtRoot/host/subpath/<revision> に展開される想定（作成は手動＝段階1）。
func (r *Resolver) PathOf(p Repo) string {
	if ov, ok := r.OverrideFor(p.Name); ok && ov.Revision != "" {
		host, subpath := parseIdentity(p.Repo)
		return filepath.Join(r.WtRoot, host, filepath.FromSlash(subpath), filepath.FromSlash(ov.Revision))
	}
	return r.GhqPathOf(p)
}

// IsCloned は path に .git があるかを返す。
func IsCloned(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// SelectRepos はロールと optional でリポを絞る（repo の -g グループ相当）。
func (r *Resolver) SelectRepos(roles []string, includeOptional bool) []Repo {
	want := make(map[string]bool, len(roles))
	for _, x := range roles {
		if t := strings.TrimSpace(x); t != "" { // "a, b" のような空白付き入力も吸収
			want[t] = true
		}
	}
	var out []Repo
	for _, p := range r.WS.Repos {
		if p.Optional && !includeOptional {
			continue
		}
		if len(want) > 0 && !want[p.Role] {
			continue
		}
		out = append(out, p)
	}
	return out
}

func expandHome(p string) string {
	if p == "~" {
		h, _ := os.UserHomeDir()
		return h
	}
	if strings.HasPrefix(p, "~/") {
		h, _ := os.UserHomeDir()
		return filepath.Join(h, p[2:])
	}
	return p
}
