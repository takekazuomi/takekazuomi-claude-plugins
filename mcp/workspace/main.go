// Command workspace-mcp は workspace.yaml を真実の源とするワークスペースの実行体。
// 1つのバイナリで3つの役割を兼ねる:
//
//	workspace-mcp validate <file>            参照整合を検証する（CI ゲート）
//	workspace-mcp --manifest ... --print-paths --roles a,b
//	                                         解決済みパスを1行ずつ出力して終了（--add-dir 用）
//	workspace-mcp --manifest ...             MCP サーバ（stdio）として常駐
//
// 所在(どこに clone するか)はマニフェストに持たせず、ユーザーごとの ghq(GHQ_ROOT / ghq.root)
// と worktree(WT_ROOT)から決定論的に解決する。Overlay(workspace.local.yaml)があれば
// 差し替えたリポは worktree パスを返す。
//
// 唯一の副作用は setup_workspace の ghq get(clone/更新)。残りは読み取り専用。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ws "github.com/takekazuomi/takekazuomi-claude-plugins/mcp/pkg/workspace"
)

func main() {
	// validate サブコマンド（CI ゲート。MCP ツールは CI から呼べないため必要）。
	if len(os.Args) >= 2 && os.Args[1] == "validate" {
		os.Exit(runValidate(os.Args[2:]))
	}

	manifest := flag.String("manifest", envOr("WORKSPACE_MANIFEST", "workspace.yaml"), "path to workspace.yaml")
	local := flag.String("local", "", "path to workspace.local.yaml (overlay). default: <manifest dir>/workspace.local.yaml")
	printPaths := flag.Bool("print-paths", false, "resolve paths and print one per line, then exit (for --add-dir $(...))")
	rolesFlag := flag.String("roles", "", "comma-separated role ids to filter (with --print-paths)")
	includeOptional := flag.Bool("include-optional", false, "include optional repos (with --print-paths)")
	flag.Parse()

	workspace, err := ws.Load(*manifest)
	if err != nil {
		log.Fatalf("load manifest: %v", err)
	}
	overlay, err := ws.LoadOverlay(overlayPath(*manifest, *local))
	if err != nil {
		log.Fatalf("load overlay: %v", err)
	}
	s := &server{r: ws.NewResolver(workspace, overlay), manifestPath: *manifest}

	// CLI モード: 起動前ブートストラップ用にパスを出力して終了。
	if *printPaths {
		var roles []string
		if *rolesFlag != "" {
			roles = strings.Split(*rolesFlag, ",")
		}
		for _, p := range s.r.SelectRepos(roles, *includeOptional) {
			fmt.Println(s.r.PathOf(p))
		}
		return
	}

	// MCP サーバモード(stdio)。
	srv := mcp.NewServer(&mcp.Implementation{Name: "workspace", Version: "v0.1.0"}, nil)
	s.register(srv)
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

// runValidate は validate サブコマンドの本体。エラーなら 1、使用法/IO/パース失敗は 2、正常は 0。
func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: workspace-mcp validate <workspace.yaml>")
		return 2
	}
	workspace, err := ws.Load(fs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	errs, warns := ws.Validate(workspace)
	for _, w := range warns {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, "error:", e)
	}
	if len(errs) > 0 {
		return 1
	}
	fmt.Printf("ok: workspace %q (%d repos, %d roles)\n", workspace.Workspace, len(workspace.Repos), len(workspace.Roles))
	return 0
}

// overlayPath は Overlay マニフェストのパスを決める。--local 優先、無ければマニフェストと同じディレクトリの workspace.local.yaml。
func overlayPath(manifest, local string) string {
	if local != "" {
		return local
	}
	return filepath.Join(filepath.Dir(manifest), "workspace.local.yaml")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
