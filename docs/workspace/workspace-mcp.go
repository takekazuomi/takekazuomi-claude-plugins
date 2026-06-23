// Command workspace-mcp は workspace.yaml を真実の源とするワークスペース MCP サーバ。
//
// 単一ワークスペースディレクトリを作らない(ghq がリポをフラットに配置する)設計のため、
// Claude Code のディレクトリツリー由来のワークスペース文脈が使えない。その穴を埋める
// location-independent な知識層がこのサーバ。iamraghuveer の repo-registry MCP に相当する。
//
// 公開ツール:
//   workspace_info      ワークスペースの概要(ロール語彙・リポ数・関係・スキル束縛)
//   list_repos          リポ一覧(URL・ロール・解決済みローカルパス・clone済みか)
//   resolve_paths       --add-dir に渡すローカルパス一覧(ロールで絞れる)
//   setup_workspace     ghq get で(ロールで絞った)リポを clone/更新する
//   repo_relationships  あるリポの依存関係(consumes/implements 等)を双方向で返す
//
// 所在(どこに clone するか)はマニフェストに持たせず、ユーザーごとの ghq(GHQ_ROOT / ghq.root)
// から決定論的に解決する。マニフェストは識別子(URL)とロールと関係だけを持つ。
//
// 起動と CLI:
//   workspace-mcp --manifest path/to/workspace.yaml          # MCP サーバ(stdio)として常駐
//   workspace-mcp --manifest ... --print-paths --roles a,b   # パスを1行ずつ出力して終了
//       → 起動前ブートストラップに: claude --add-dir $(workspace-mcp --print-paths --roles source,api-definition)
//
// 依存: github.com/modelcontextprotocol/go-sdk, gopkg.in/yaml.v3
//
// セキュリティ: 唯一の副作用は setup_workspace の ghq get(clone/更新)。マニフェストから
// 任意コマンドは実行しない(リポ識別子は exec 引数として渡し、シェルを介さない)。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

// ---- workspace.yaml の構造(validate-workspace と同じモデル) ----

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
	ID          string `yaml:"id" json:"id"`
	Description string `yaml:"description" json:"description,omitempty"`
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
	When  string   `yaml:"when" json:"when,omitempty"`
	Use   string   `yaml:"use" json:"use"`
	Roles []string `yaml:"roles" json:"roles,omitempty"`
}

// WS はロード済みワークスペースと解決済み ghq ルートを保持する。
type WS struct {
	ws   *Workspace
	root string
}

// ---- ghq 所在解決 ----

func ghqRoot() string {
	if r := os.Getenv("GHQ_ROOT"); r != "" { // GHQ_ROOT は他設定を上書きする
		return expandHome(r)
	}
	if out, err := exec.Command("git", "config", "--path", "--get-all", "ghq.root").Output(); err == nil {
		s := strings.TrimRight(string(out), "\n")
		if s != "" {
			lines := strings.Split(s, "\n")
			return expandHome(lines[len(lines)-1]) // 複数あるとき新規 clone の主ルートは最後(x-motemen/ghq)
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "ghq")
}

// parseIdentity は ghq の識別子(host/user/repo、user/repo、各種 URL)から host とサブパスを取り出す。
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

// pathOf はリポ識別子の ghq 上の決定論的ローカルパスを返す(未 clone でも算出できる)。
func (w *WS) pathOf(id string) string {
	host, subpath := parseIdentity(id)
	return filepath.Join(w.root, host, filepath.FromSlash(subpath))
}

func isCloned(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// selectRepos はロールと optional でリポを絞る(repo の -g グループ相当)。
func (w *WS) selectRepos(roles []string, includeOptional bool) []Repo {
	want := make(map[string]bool, len(roles))
	for _, r := range roles {
		if r != "" {
			want[r] = true
		}
	}
	var out []Repo
	for _, p := range w.ws.Repos {
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

// ---- ツール入出力 ----

type NoInput struct{}

type InfoOut struct {
	Workspace         string         `json:"workspace"`
	Description       string         `json:"description,omitempty"`
	Roles             []Role         `json:"roles"`
	RepoCount         int            `json:"repoCount"`
	RelationshipCount int            `json:"relationshipCount"`
	Skills            []SkillBinding `json:"skills,omitempty"`
}

func (w *WS) Info(_ context.Context, _ *mcp.CallToolRequest, _ NoInput) (*mcp.CallToolResult, InfoOut, error) {
	return nil, InfoOut{
		Workspace:         w.ws.Workspace,
		Description:       w.ws.Description,
		Roles:             w.ws.Roles,
		RepoCount:         len(w.ws.Repos),
		RelationshipCount: len(w.ws.Relationships),
		Skills:            w.ws.Skills,
	}, nil
}

type ListReposIn struct {
	Role            string `json:"role,omitempty" jsonschema:"単一ロールidで絞る。空なら全部"`
	IncludeOptional bool   `json:"includeOptional,omitempty" jsonschema:"optional 指定のリポも含める"`
}

type RepoView struct {
	Name     string `json:"name"`
	Repo     string `json:"repo"`
	Role     string `json:"role"`
	Revision string `json:"revision,omitempty"`
	Path     string `json:"path"`
	Cloned   bool   `json:"cloned"`
	Optional bool   `json:"optional,omitempty"`
}

type ListReposOut struct {
	Repos []RepoView `json:"repos"`
}

func (w *WS) ListRepos(_ context.Context, _ *mcp.CallToolRequest, in ListReposIn) (*mcp.CallToolResult, ListReposOut, error) {
	var roles []string
	if in.Role != "" {
		roles = []string{in.Role}
	}
	var views []RepoView
	for _, p := range w.selectRepos(roles, in.IncludeOptional) {
		path := w.pathOf(p.Repo)
		views = append(views, RepoView{
			Name: p.Name, Repo: p.Repo, Role: p.Role, Revision: p.Revision,
			Path: path, Cloned: isCloned(path), Optional: p.Optional,
		})
	}
	return nil, ListReposOut{Repos: views}, nil
}

type ResolvePathsIn struct {
	Roles           []string `json:"roles,omitempty" jsonschema:"これらのロールidで絞る。空なら全部"`
	IncludeOptional bool     `json:"includeOptional,omitempty"`
	OnlyCloned      bool     `json:"onlyCloned,omitempty" jsonschema:"既に clone 済みのリポだけに限る"`
}

type ResolvePathsOut struct {
	Paths         []string `json:"paths"`
	AddDirCommand string   `json:"addDirCommand" jsonschema:"そのまま使える --add-dir 引数文字列"`
}

func (w *WS) ResolvePaths(_ context.Context, _ *mcp.CallToolRequest, in ResolvePathsIn) (*mcp.CallToolResult, ResolvePathsOut, error) {
	var paths []string
	for _, p := range w.selectRepos(in.Roles, in.IncludeOptional) {
		path := w.pathOf(p.Repo)
		if in.OnlyCloned && !isCloned(path) {
			continue
		}
		paths = append(paths, path)
	}
	cmd := ""
	if len(paths) > 0 {
		cmd = "--add-dir " + strings.Join(quoteAll(paths), " ")
	}
	return nil, ResolvePathsOut{Paths: paths, AddDirCommand: cmd}, nil
}

type SetupIn struct {
	Roles           []string `json:"roles,omitempty" jsonschema:"これらのロールidで絞る。空なら全部"`
	IncludeOptional bool     `json:"includeOptional,omitempty"`
	Update          bool     `json:"update,omitempty" jsonschema:"既存 clone を ghq get -u で更新する"`
}

type RepoStatus struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Status string `json:"status"` // ok | error
	Detail string `json:"detail,omitempty"`
}

type SetupOut struct {
	Results []RepoStatus `json:"results"`
}

func (w *WS) Setup(ctx context.Context, _ *mcp.CallToolRequest, in SetupIn) (*mcp.CallToolResult, SetupOut, error) {
	var res []RepoStatus
	for _, p := range w.selectRepos(in.Roles, in.IncludeOptional) {
		args := []string{"get"}
		if in.Update {
			args = append(args, "-u")
		}
		args = append(args, p.Repo)
		out, err := exec.CommandContext(ctx, "ghq", args...).CombinedOutput()
		st := RepoStatus{Name: p.Name, Path: w.pathOf(p.Repo)}
		if err != nil {
			st.Status = "error"
			st.Detail = strings.TrimSpace(string(out))
		} else {
			st.Status = "ok"
			st.Detail = strings.TrimSpace(string(out))
		}
		res = append(res, st)
	}
	return nil, SetupOut{Results: res}, nil
}

type RelIn struct {
	Name string `json:"name" jsonschema:"調べるリポの name(エイリアス)"`
}

type RelEdge struct {
	Repo string `json:"repo"`
	Type string `json:"type"`
}

type RelOut struct {
	Name     string    `json:"name"`
	Outgoing []RelEdge `json:"outgoing"` // name -> 他 (consumes/implements...)
	Incoming []RelEdge `json:"incoming"` // 他 -> name
}

func (w *WS) Relationships(_ context.Context, _ *mcp.CallToolRequest, in RelIn) (*mcp.CallToolResult, RelOut, error) {
	known := false
	for _, p := range w.ws.Repos {
		if p.Name == in.Name {
			known = true
			break
		}
	}
	if !known {
		return nil, RelOut{}, fmt.Errorf("unknown repo name: %q", in.Name)
	}
	out := RelOut{Name: in.Name}
	for _, r := range w.ws.Relationships {
		if r.From == in.Name {
			out.Outgoing = append(out.Outgoing, RelEdge{Repo: r.To, Type: r.Type})
		}
		if r.To == in.Name {
			out.Incoming = append(out.Incoming, RelEdge{Repo: r.From, Type: r.Type})
		}
	}
	return nil, out, nil
}

// ---- main ----

func main() {
	manifest := flag.String("manifest", envOr("WORKSPACE_MANIFEST", "workspace.yaml"), "path to workspace.yaml")
	printPaths := flag.Bool("print-paths", false, "resolve paths and print one per line, then exit (for --add-dir $(...))")
	rolesFlag := flag.String("roles", "", "comma-separated role ids to filter (with --print-paths)")
	includeOptional := flag.Bool("include-optional", false, "include optional repos (with --print-paths)")
	flag.Parse()

	data, err := os.ReadFile(*manifest)
	if err != nil {
		log.Fatalf("read manifest: %v", err)
	}
	var ws Workspace
	if err := yaml.Unmarshal(data, &ws); err != nil {
		log.Fatalf("parse manifest: %v", err)
	}
	w := &WS{ws: &ws, root: ghqRoot()}

	// CLI モード: 起動前ブートストラップ用にパスを出力して終了。
	if *printPaths {
		var roles []string
		if *rolesFlag != "" {
			roles = strings.Split(*rolesFlag, ",")
		}
		for _, p := range w.selectRepos(roles, *includeOptional) {
			fmt.Println(w.pathOf(p.Repo))
		}
		return
	}

	// MCP サーバモード(stdio)。
	server := mcp.NewServer(&mcp.Implementation{Name: "workspace", Version: "v0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "workspace_info",
		Description: "このワークスペースの概要(名前・宣言されたロール語彙・リポ数・関係数・スキル束縛)を返す。何のワークスペースかを把握するときに最初に使う。",
	}, w.Info)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_repos",
		Description: "ワークスペースのリポ一覧を返す。各リポの URL・ロール・解決済みローカルパス・clone済みかを含む。role で絞れる。どのリポがどこにあるかを知るときに使う。",
	}, w.ListRepos)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "resolve_paths",
		Description: "ワークスペースのリポのローカルパス一覧と、そのまま使える --add-dir 引数文字列を返す。複数リポを1セッションで開く(claude --add-dir に渡す)ときに使う。roles で絞れる。",
	}, w.ResolvePaths)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "setup_workspace",
		Description: "このワークスペースのリポを ghq get で各自のローカル(ghq ルート)に clone/更新する。新しいマシンでのセットアップや、メンバーがリポを追加した後に使う。roles と optional で取得対象を絞れる。",
	}, w.Setup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "repo_relationships",
		Description: "あるリポ(name)の依存関係を双方向で返す(consumes/implements 等の出方向と、それに依存する入方向)。リポ間の影響範囲を調べるときに使う。",
	}, w.Relationships)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

// ---- helpers ----

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
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
