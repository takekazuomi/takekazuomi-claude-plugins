package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ws "github.com/takekazuomi/takekazuomi-claude-plugins/mcp/pkg/workspace"
)

// startServer は testdata のマニフェストと Overlay でサーバを立て、
// インメモリトランスポートで接続したクライアントセッションを返す。
// ghq ルートと worktree ルートは固定値に差し替え、解決パスを予測可能にする。
func startServer(t *testing.T) *mcp.ClientSession {
	t.Helper()
	t.Setenv("GHQ_ROOT", filepath.FromSlash("/ghqroot"))
	t.Setenv("WT_ROOT", filepath.FromSlash("/wtroot"))

	ctx := context.Background()
	wsObj, err := ws.Load("testdata/workspace.yaml")
	if err != nil {
		t.Fatalf("load workspace: %v", err)
	}
	overlay, err := ws.LoadOverlay("testdata/workspace.local.yaml")
	if err != nil {
		t.Fatalf("load overlay: %v", err)
	}
	s := &server{r: ws.NewResolver(wsObj, overlay), manifestPath: "testdata/workspace.yaml"}

	srv := mcp.NewServer(&mcp.Implementation{Name: "workspace", Version: "test"}, nil)
	s.register(srv)

	clientT, serverT := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverT, nil) // サーバを先に接続する
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// callTool はツールを呼び、構造化出力の JSON を型 T にデコードして返す。
func callTool[T any](t *testing.T, cs *mcp.ClientSession, name string, args any) T {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("tool %s returned error: %+v", name, res.Content)
	}
	if len(res.Content) == 0 {
		t.Fatalf("tool %s: empty content", name)
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("tool %s: content is %T, want *TextContent", name, res.Content[0])
	}
	var out T
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		t.Fatalf("tool %s: unmarshal %q: %v", name, tc.Text, err)
	}
	return out
}

func TestScenarioWorkspaceInfo(t *testing.T) {
	cs := startServer(t)
	got := callTool[infoOut](t, cs, "workspace_info", noInput{})
	if got.Workspace != "payments" {
		t.Errorf("workspace = %q, want payments", got.Workspace)
	}
	if len(got.Roles) != 4 {
		t.Errorf("roles = %d, want 4", len(got.Roles))
	}
	if got.RepoCount != 4 {
		t.Errorf("repoCount = %d, want 4", got.RepoCount)
	}
	if got.RelationshipCount != 2 {
		t.Errorf("relationshipCount = %d, want 2", got.RelationshipCount)
	}
	if got.ActiveOverrides != 1 {
		t.Errorf("activeOverrides = %d, want 1", got.ActiveOverrides)
	}
}

func TestScenarioListReposByRole(t *testing.T) {
	cs := startServer(t)
	got := callTool[listReposOut](t, cs, "list_repos", listReposIn{Role: "source"})
	if len(got.Repos) != 1 {
		t.Fatalf("repos = %d, want 1", len(got.Repos))
	}
	r := got.Repos[0]
	if r.Name != "payment-service" {
		t.Errorf("name = %q, want payment-service", r.Name)
	}
	if r.Overridden {
		t.Errorf("payment-service should not be overridden")
	}
	want := filepath.FromSlash("/ghqroot/github.com/acme/payment-service")
	if r.Path != want {
		t.Errorf("path = %q, want %q", r.Path, want)
	}
}

// Overlay が効くこと: payment-proto は worktree パスを返す。
func TestScenarioResolvePathsWithOverlay(t *testing.T) {
	cs := startServer(t)
	got := callTool[resolvePathsOut](t, cs, "resolve_paths", resolvePathsIn{Roles: []string{"source", "api-definition"}})
	if len(got.Paths) != 2 {
		t.Fatalf("paths = %v, want 2", got.Paths)
	}
	wantSvc := filepath.FromSlash("/ghqroot/github.com/acme/payment-service")
	wantProto := filepath.FromSlash("/wtroot/github.com/acme/payment-proto/feature/api-v2")
	if got.Paths[0] != wantSvc {
		t.Errorf("paths[0] = %q, want %q", got.Paths[0], wantSvc)
	}
	if got.Paths[1] != wantProto {
		t.Errorf("paths[1] = %q (overlay worktree), want %q", got.Paths[1], wantProto)
	}
	if got.AddDirCommand == "" {
		t.Error("addDirCommand should not be empty")
	}
}

func TestScenarioRelationships(t *testing.T) {
	cs := startServer(t)
	got := callTool[relOut](t, cs, "repo_relationships", relIn{Name: "payment-service"})
	if len(got.Outgoing) != 1 || got.Outgoing[0].Repo != "payment-proto" || got.Outgoing[0].Type != "implements" {
		t.Errorf("outgoing = %+v, want [{payment-proto implements}]", got.Outgoing)
	}
	if len(got.Incoming) != 1 || got.Incoming[0].Repo != "payment-app" || got.Incoming[0].Type != "consumes" {
		t.Errorf("incoming = %+v, want [{payment-app consumes}]", got.Incoming)
	}
}

func TestScenarioRelationshipsUnknownRepo(t *testing.T) {
	cs := startServer(t)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "repo_relationships", Arguments: relIn{Name: "ghost"}})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Error("expected IsError for unknown repo name")
	}
}

func TestScenarioValidateOK(t *testing.T) {
	cs := startServer(t)
	got := callTool[validateOut](t, cs, "validate_workspace", validateIn{})
	if !got.OK {
		t.Errorf("ok = false, errors = %v", got.Errors)
	}
}

func TestScenarioValidateInvalidManifest(t *testing.T) {
	cs := startServer(t)
	got := callTool[validateOut](t, cs, "validate_workspace", validateIn{Manifest: "testdata/invalid.yaml"})
	if got.OK {
		t.Error("ok = true, want false for invalid manifest")
	}
	if len(got.Errors) == 0 {
		t.Error("expected errors for invalid manifest")
	}
}

// setup_workspace は ghq get を呼ぶ。フェイク ghq を PATH に置いて副作用を検証する。
func TestScenarioSetupWorkspace(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake ghq script is POSIX shell")
	}
	bin := t.TempDir()
	fake := filepath.Join(bin, "ghq")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake ghq: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	cs := startServer(t)
	got := callTool[setupOut](t, cs, "setup_workspace", setupIn{Roles: []string{"source"}})
	if len(got.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Status != "ok" {
		t.Errorf("status = %q, want ok (detail: %s)", got.Results[0].Status, got.Results[0].Detail)
	}
	wantPath := filepath.FromSlash("/ghqroot/github.com/acme/payment-service")
	if got.Results[0].Path != wantPath {
		t.Errorf("path = %q, want %q (clone 先は ghq、Overlay を無視)", got.Results[0].Path, wantPath)
	}
}
