---
name: workspace
description: >-
  複数の Git リポにまたがるコーディング・ワークスペースを操作するときに使う。
  workspace.yaml（ロール・リポ・関係・スキル束縛の定義）を真実の源とし、
  ghq で所在を解決して複数リポを1セッションに取り込む。
  リポ一覧・所在解決・clone・依存関係の参照・定義の検証を MCP ツールで行う。
  新しいマシンでのセットアップ、関連リポの取得、影響範囲の調査、新 API ブランチの一時参照に使う。
---

# workspace

マルチリポ・ワークスペースを `workspace.yaml`（Definition）から操作する。所在（clone 先）はマニフェストに持たせず、ghq（`GHQ_ROOT` / `ghq.root`）から決定論的に解決する。再現性はワークスペースの責務ではなく、各リポの git tag・go.sum が担う（作業セット型）。

## 前提

- `workspace.yaml` をプロジェクトルートに置く（`.mcp.json` が `${CLAUDE_PROJECT_DIR}/workspace.yaml` を指す）
- バイナリ `workspace-mcp` は `SessionStart` フックが `${CLAUDE_PLUGIN_DATA}/bin` に用意する（GitHub Releases、無ければ `go install`）
- ghq がインストール済みであること

## MCP ツール

| ツール | 用途 |
|---|---|
| `workspace_info` | ワークスペースの概要（ロール語彙・リポ数・関係・スキル束縛・有効な Overlay 数）|
| `list_repos` | リポ一覧（URL・ロール・解決済みパス・clone済みか・Overlay 差し替えの有無）。role で絞れる |
| `resolve_paths` | ローカルパス一覧と、そのまま使える `--add-dir` 引数文字列。roles で絞れる |
| `setup_workspace` | `ghq get` でリポを clone/更新（唯一の副作用）。roles と optional で絞れる |
| `repo_relationships` | あるリポの依存関係を双方向で返す |
| `validate_workspace` | 定義の妥当性（参照整合・命名・ロール検証）を検証 |

## 典型フロー

1. **把握**: `workspace_info` で何のワークスペースかを知る
2. **セットアップ**: `setup_workspace` で必要なロールのリポを clone
3. **取り込み**: `resolve_paths` の `--add-dir` 文字列で複数リポを1セッションに入れる
   - エージェントは `--add-dir` を自分で実行できないため、起動時に `workspace-mcp --print-paths --roles ...` の出力を `claude --add-dir $(...)` へ渡す
4. **調査**: `repo_relationships` で変更の影響範囲を確認

## Overlay（一時差し替え）

新 API を試すなど、特定リポを別ブランチで参照したいときは、プロジェクトルートに `workspace.local.yaml`（`.gitignore` 管理）を置く。`go.work` の `replace` に相当し、共有定義は変えない。

```yaml
# workspace.local.yaml（非コミット）
overrides:
  - name: payment-proto
    revision: feature/api-v2     # 一時的にこのブランチを参照する
```

差し替えたリポは worktree（`WT_ROOT`、既定 `~/wt` の ghq 同型配置 ＋ ブランチ）に展開する。`resolve_paths` はその worktree パスを返す。worktree の作成は手動（`git worktree add`）。

## 検証

- セッション内: `validate_workspace`（編集の後に妥当性確認）
- CI: `workspace-mcp validate <workspace.yaml>`（参照整合をゲート）
