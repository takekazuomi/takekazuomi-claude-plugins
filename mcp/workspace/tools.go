package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ws "github.com/takekazuomi/takekazuomi-claude-plugins/mcp/pkg/workspace"
)

// server は解決済みワークスペースと起動時マニフェストパスを保持する。
type server struct {
	r            *ws.Resolver
	manifestPath string
}

// register は全 MCP ツールをサーバに登録する。
func (s *server) register(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "workspace_info",
		Description: "このワークスペースの概要(名前・宣言されたロール語彙・リポ数・関係数・スキル束縛・有効な Overlay 数)を返す。何のワークスペースかを把握するときに最初に使う。",
	}, s.info)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_repos",
		Description: "リポ一覧を返す。各リポの URL・ロール・解決済みローカルパス・clone済みか・Overlay で差し替えられているかを含む。role で絞れる。",
	}, s.listRepos)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "resolve_paths",
		Description: "リポのローカルパス一覧と、そのまま使える --add-dir 引数文字列を返す。複数リポを1セッションで開くときに使う。roles で絞れる。Overlay 差し替え分は worktree パスを返す。",
	}, s.resolvePaths)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "setup_workspace",
		Description: "リポを ghq get で各自のローカル(ghq ルート)に clone/更新する。新しいマシンでのセットアップや、メンバーがリポを追加した後に使う。roles と optional で取得対象を絞れる。",
	}, s.setup)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "repo_relationships",
		Description: "あるリポ(name)の依存関係を双方向で返す(consumes/implements 等の出方向と、それに依存する入方向)。影響範囲を調べるときに使う。",
	}, s.relationships)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "validate_workspace",
		Description: "ワークスペース定義の妥当性(参照整合・命名・ロール検証)を検証し、ok・errors・warnings を返す。作成や編集の後に使う。manifest を省略すると起動時のマニフェストを検証する。",
	}, s.validate)
}

// ---- workspace_info ----

type noInput struct{}

type infoOut struct {
	Workspace         string            `json:"workspace"`
	Description       string            `json:"description,omitempty"`
	Roles             []ws.Role         `json:"roles"`
	RepoCount         int               `json:"repoCount"`
	RelationshipCount int               `json:"relationshipCount"`
	Skills            []ws.SkillBinding `json:"skills,omitempty"`
	ActiveOverrides   int               `json:"activeOverrides"`
}

func (s *server) info(_ context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, infoOut, error) {
	w := s.r.WS
	return nil, infoOut{
		Workspace:         w.Workspace,
		Description:       w.Description,
		Roles:             w.Roles,
		RepoCount:         len(w.Repos),
		RelationshipCount: len(w.Relationships),
		Skills:            w.Skills,
		ActiveOverrides:   len(s.r.Overlay.Overrides),
	}, nil
}

// ---- list_repos ----

type listReposIn struct {
	Role            string `json:"role,omitempty" jsonschema:"単一ロールidで絞る。空なら全部"`
	IncludeOptional bool   `json:"includeOptional,omitempty" jsonschema:"optional 指定のリポも含める"`
}

type repoView struct {
	Name       string `json:"name"`
	Repo       string `json:"repo"`
	Role       string `json:"role"`
	Revision   string `json:"revision,omitempty"`
	Path       string `json:"path"`
	Cloned     bool   `json:"cloned"`
	Optional   bool   `json:"optional,omitempty"`
	Overridden bool   `json:"overridden,omitempty"`
}

type listReposOut struct {
	Repos []repoView `json:"repos"`
}

func (s *server) listRepos(_ context.Context, _ *mcp.CallToolRequest, in listReposIn) (*mcp.CallToolResult, listReposOut, error) {
	var roles []string
	if in.Role != "" {
		roles = []string{in.Role}
	}
	var views []repoView
	for _, p := range s.r.SelectRepos(roles, in.IncludeOptional) {
		path := s.r.PathOf(p)
		rev := p.Revision
		ov, overridden := s.r.OverrideFor(p.Name)
		if overridden && ov.Revision != "" {
			rev = ov.Revision
		}
		views = append(views, repoView{
			Name: p.Name, Repo: p.Repo, Role: p.Role, Revision: rev,
			Path: path, Cloned: ws.IsCloned(path), Optional: p.Optional, Overridden: overridden,
		})
	}
	return nil, listReposOut{Repos: views}, nil
}

// ---- resolve_paths ----

type resolvePathsIn struct {
	Roles           []string `json:"roles,omitempty" jsonschema:"これらのロールidで絞る。空なら全部"`
	IncludeOptional bool     `json:"includeOptional,omitempty"`
	OnlyCloned      bool     `json:"onlyCloned,omitempty" jsonschema:"既に clone 済みのリポだけに限る"`
}

type resolvePathsOut struct {
	Paths         []string `json:"paths"`
	AddDirCommand string   `json:"addDirCommand" jsonschema:"そのまま使える --add-dir 引数文字列"`
}

func (s *server) resolvePaths(_ context.Context, _ *mcp.CallToolRequest, in resolvePathsIn) (*mcp.CallToolResult, resolvePathsOut, error) {
	var paths []string
	for _, p := range s.r.SelectRepos(in.Roles, in.IncludeOptional) {
		path := s.r.PathOf(p)
		if in.OnlyCloned && !ws.IsCloned(path) {
			continue
		}
		paths = append(paths, path)
	}
	cmd := ""
	if len(paths) > 0 {
		cmd = "--add-dir " + strings.Join(quoteAll(paths), " ")
	}
	return nil, resolvePathsOut{Paths: paths, AddDirCommand: cmd}, nil
}

// ---- setup_workspace ----

type setupIn struct {
	Roles           []string `json:"roles,omitempty" jsonschema:"これらのロールidで絞る。空なら全部"`
	IncludeOptional bool     `json:"includeOptional,omitempty"`
	Update          bool     `json:"update,omitempty" jsonschema:"既存 clone を ghq get -u で更新する"`
}

type repoStatus struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Status string `json:"status"` // ok | error
	Detail string `json:"detail,omitempty"`
}

type setupOut struct {
	Results []repoStatus `json:"results"`
}

func (s *server) setup(ctx context.Context, _ *mcp.CallToolRequest, in setupIn) (*mcp.CallToolResult, setupOut, error) {
	var res []repoStatus
	for _, p := range s.r.SelectRepos(in.Roles, in.IncludeOptional) {
		args := []string{"get"}
		if in.Update {
			args = append(args, "-u")
		}
		args = append(args, p.Repo) // 識別子は exec 引数として渡す（シェルを介さない）
		out, err := exec.CommandContext(ctx, "ghq", args...).CombinedOutput()
		st := repoStatus{Name: p.Name, Path: s.r.GhqPathOf(p), Detail: strings.TrimSpace(string(out))}
		if err != nil {
			st.Status = "error"
		} else {
			st.Status = "ok"
		}
		res = append(res, st)
	}
	return nil, setupOut{Results: res}, nil
}

// ---- repo_relationships ----

type relIn struct {
	Name string `json:"name" jsonschema:"調べるリポの name(エイリアス)"`
}

type relEdge struct {
	Repo string `json:"repo"`
	Type string `json:"type"`
}

type relOut struct {
	Name     string    `json:"name"`
	Outgoing []relEdge `json:"outgoing"`
	Incoming []relEdge `json:"incoming"`
}

func (s *server) relationships(_ context.Context, _ *mcp.CallToolRequest, in relIn) (*mcp.CallToolResult, relOut, error) {
	known := false
	for _, p := range s.r.WS.Repos {
		if p.Name == in.Name {
			known = true
			break
		}
	}
	if !known {
		return nil, relOut{}, fmt.Errorf("unknown repo name: %q", in.Name)
	}
	out := relOut{Name: in.Name}
	for _, rel := range s.r.WS.Relationships {
		if rel.From == in.Name {
			out.Outgoing = append(out.Outgoing, relEdge{Repo: rel.To, Type: rel.Type})
		}
		if rel.To == in.Name {
			out.Incoming = append(out.Incoming, relEdge{Repo: rel.From, Type: rel.Type})
		}
	}
	return nil, out, nil
}

// ---- validate_workspace ----

type validateIn struct {
	Manifest string `json:"manifest,omitempty" jsonschema:"検証するマニフェストのパス。省略時は起動時のマニフェスト"`
}

type validateOut struct {
	OK       bool     `json:"ok"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

func (s *server) validate(_ context.Context, _ *mcp.CallToolRequest, in validateIn) (*mcp.CallToolResult, validateOut, error) {
	target := s.r.WS
	if in.Manifest != "" {
		loaded, err := ws.Load(in.Manifest)
		if err != nil {
			return nil, validateOut{}, err
		}
		target = loaded
	}
	errs, warns := ws.Validate(target)
	return nil, validateOut{OK: len(errs) == 0, Errors: errs, Warnings: warns}, nil
}

// quoteAll は空白を含むパスを二重引用符で囲む。
func quoteAll(ps []string) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		if strings.ContainsAny(p, " \t") {
			out[i] = `"` + p + `"`
		} else {
			out[i] = p
		}
	}
	return out
}
