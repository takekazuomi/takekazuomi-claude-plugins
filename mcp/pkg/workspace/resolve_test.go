package workspace

import (
	"path/filepath"
	"testing"
)

func TestParseIdentity(t *testing.T) {
	cases := []struct {
		in      string
		host    string
		subpath string
	}{
		{"github.com/acme/repo", "github.com", "acme/repo"},
		{"acme/repo", "github.com", "acme/repo"}, // host 省略
		{"https://github.com/acme/repo.git", "github.com", "acme/repo"},
		{"git@github.com:acme/repo.git", "github.com", "acme/repo"},
		{"ssh://git@example.com/acme/repo", "example.com", "acme/repo"},
		{"gitlab.com/group/sub/repo", "gitlab.com", "group/sub/repo"},
	}
	for _, c := range cases {
		host, sub := parseIdentity(c.in)
		if host != c.host || sub != c.subpath {
			t.Errorf("parseIdentity(%q) = (%q, %q), want (%q, %q)", c.in, host, sub, c.host, c.subpath)
		}
	}
}

func TestPathOfGhq(t *testing.T) {
	r := &Resolver{
		WS:      &Workspace{},
		Overlay: &Overlay{},
		GhqRoot: "/ghq",
		WtRoot:  "/wt",
	}
	p := Repo{Name: "proto", Repo: "github.com/acme/proto"}
	want := filepath.FromSlash("/ghq/github.com/acme/proto")
	if got := r.PathOf(p); got != want {
		t.Errorf("PathOf = %q, want %q", got, want)
	}
}

// Overlay で差し替えたリポは worktree パス（WtRoot/host/sub/<revision>）を返す。
func TestPathOfOverlayWorktree(t *testing.T) {
	r := &Resolver{
		WS:      &Workspace{},
		Overlay: &Overlay{Overrides: []Override{{Name: "proto", Revision: "feature/api-v2"}}},
		GhqRoot: "/ghq",
		WtRoot:  "/wt",
	}
	p := Repo{Name: "proto", Repo: "github.com/acme/proto"}
	want := filepath.FromSlash("/wt/github.com/acme/proto/feature/api-v2")
	if got := r.PathOf(p); got != want {
		t.Errorf("PathOf(overlay) = %q, want %q", got, want)
	}
	// GhqPathOf は Overlay を無視して clone 先を返す。
	wantGhq := filepath.FromSlash("/ghq/github.com/acme/proto")
	if got := r.GhqPathOf(p); got != wantGhq {
		t.Errorf("GhqPathOf = %q, want %q", got, wantGhq)
	}
}

func TestSelectReposFilters(t *testing.T) {
	r := &Resolver{
		WS: &Workspace{Repos: []Repo{
			{Name: "svc", Role: "source"},
			{Name: "proto", Role: "api-definition"},
			{Name: "docs", Role: "documentation", Optional: true},
		}},
		Overlay: &Overlay{},
	}
	// ロール絞り込み（空白付き入力も許容）
	got := r.SelectRepos([]string{" source "}, false)
	if len(got) != 1 || got[0].Name != "svc" {
		t.Fatalf("role filter failed: %+v", got)
	}
	// optional は既定で除外
	if all := r.SelectRepos(nil, false); len(all) != 2 {
		t.Fatalf("optional should be excluded, got %d", len(all))
	}
	// includeOptional で含める
	if all := r.SelectRepos(nil, true); len(all) != 3 {
		t.Fatalf("includeOptional failed, got %d", len(all))
	}
}
