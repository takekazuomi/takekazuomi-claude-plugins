package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// e2eBinPath は TestMain で1回だけビルドした workspace-mcp バイナリのパス。
// 全 e2e テストで共有し、テストごとの go build プロセス起動を避ける。
var e2eBinPath string

// TestMain は e2e 用バイナリを1回ビルドしてから全テストを走らせる。
// -short ではビルドを省く（e2e は各テストの testing.Short() でスキップされる）。
func TestMain(m *testing.M) {
	flag.Parse() // testing.Short() を参照するために必要
	if !testing.Short() {
		dir, err := os.MkdirTemp("", "workspace-mcp-e2e")
		if err != nil {
			fmt.Fprintln(os.Stderr, "e2e setup:", err)
			os.Exit(1)
		}
		e2eBinPath = filepath.Join(dir, "workspace-mcp")
		if out, err := exec.Command("go", "build", "-o", e2eBinPath, ".").CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "e2e build: %v\n%s", err, out)
			os.Exit(1)
		}
	}
	// defer は os.Exit で実行されないため、終了コードを受けてから後始末する。
	code := m.Run()
	if e2eBinPath != "" {
		_ = os.RemoveAll(filepath.Dir(e2eBinPath))
	}
	os.Exit(code)
}

// TestE2EStdioRoundtrip は、実際にビルドした workspace-mcp バイナリを
// サブプロセスとして stdio 起動し、MCP クライアントから handshake → tools/list →
// tools/call を通す。インメモリではなく本番と同じ実行経路（バイナリ＋stdio）を検証する。
func TestE2EStdioRoundtrip(t *testing.T) {
	if testing.Short() {
		t.Skip("uses a built binary; skipped in -short")
	}
	bin := e2eBinPath // TestMain でビルド済み（plugins の SessionStart フックが配置するものと同じ実体）

	manifest, err := filepath.Abs("testdata/workspace.yaml") // 同ディレクトリの workspace.local.yaml が Overlay として自動ロードされる
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cmd := exec.Command(bin, "--manifest", manifest)
	cmd.Env = append(os.Environ(),
		"GHQ_ROOT="+filepath.FromSlash("/ghqroot"),
		"WT_ROOT="+filepath.FromSlash("/wtroot"),
	)

	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-probe", Version: "0"}, nil)
	cs, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect to subprocess: %v", err)
	}
	defer func() { _ = cs.Close() }()

	// tools/list: 6ツールが公開されているか。
	lt, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	got := map[string]bool{}
	for _, tool := range lt.Tools {
		got[tool.Name] = true
	}
	for _, want := range []string{
		"workspace_info", "list_repos", "resolve_paths",
		"setup_workspace", "repo_relationships", "validate_workspace",
	} {
		if !got[want] {
			t.Errorf("tool %q not advertised; got %v", want, lt.Tools)
		}
	}

	// tools/call workspace_info: 構造化出力が返るか。
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "workspace_info", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("call workspace_info: %v", err)
	}
	if res.IsError {
		t.Fatalf("workspace_info returned error: %+v", res.Content)
	}
	var info infoOut
	decodeStructured(t, res.StructuredContent, &info)
	if info.Workspace != "payments" || info.RepoCount != 4 || info.ActiveOverrides != 1 {
		t.Errorf("info = %+v, want workspace=payments repoCount=4 activeOverrides=1", info)
	}

	// tools/call resolve_paths: Overlay が効き、payment-proto が worktree パスになるか。
	res2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "resolve_paths",
		Arguments: map[string]any{"roles": []string{"api-definition"}},
	})
	if err != nil {
		t.Fatalf("call resolve_paths: %v", err)
	}
	var paths resolvePathsOut
	decodeStructured(t, res2.StructuredContent, &paths)
	wantProto := filepath.FromSlash("/wtroot/github.com/acme/payment-proto/feature/api-v2")
	if len(paths.Paths) != 1 || paths.Paths[0] != wantProto {
		t.Errorf("resolve_paths = %v, want [%s] (Overlay worktree)", paths.Paths, wantProto)
	}
}

// decodeStructured は StructuredContent(any) を JSON 経由で型に詰め直す。
func decodeStructured(t *testing.T, sc any, dst any) {
	t.Helper()
	raw, err := json.Marshal(sc)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatalf("unmarshal structured content %s: %v", raw, err)
	}
}
