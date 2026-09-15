# workspace-mcp

マルチリポ・ワークスペースの定義・所在解決・検証を行う実行体。1つのバイナリが3つのモードを兼ねる。

- **MCP サーバ**（stdio）— Claude Code から呼ぶ知識層
- **`validate` サブコマンド** — 参照整合の検証（CI ゲート）
- **`--print-paths`** — 解決済みパスを出力（起動時 `--add-dir` 用）

このドキュメントは、サーバを単独で動かして手で動作確認する手順をまとめる。

## 前提

- Go 1.26 以降、ghq
- マニフェストの例: `../workspaces/payments/workspace.yaml`
- 所在解決の環境変数（任意）
  - `GHQ_ROOT` — clone 先（既定は `ghq.root` / `~/ghq`）
  - `WT_ROOT` — Overlay の worktree ルート（既定 `~/wt`）

ビルド:

```bash
go build -o bin/workspace-mcp ./workspace
```

## 1. CLI モードの確認（最も手軽）

ネットワークも MCP クライアントも要らず、シェルだけで確認できる。

### validate サブコマンド

```bash
$ ./bin/workspace-mcp validate ../workspaces/payments/workspace.yaml
ok: workspace "payments" (4 repos, 4 roles)
```

参照整合エラー（未宣言ロール・関係端点の欠落など）があれば、`error:` を stderr に出して終了コード 1 で終わる。警告（未使用ロール等）は終了コードに影響しない。

### --print-paths

ロールで絞り、解決済みローカルパスを1行ずつ出力する。

```bash
$ GHQ_ROOT=/tmp/ghq ./bin/workspace-mcp \
    --manifest ../workspaces/payments/workspace.yaml \
    --print-paths --roles source,api-definition
/tmp/ghq/github.com/acme/payment-service
/tmp/ghq/github.com/acme/payment-proto
```

そのまま `--add-dir` に渡せる。

```bash
claude --add-dir $(./bin/workspace-mcp --manifest ../workspaces/payments/workspace.yaml \
                                       --print-paths --roles source,api-definition)
```

### Overlay の確認

マニフェストと同じディレクトリに `workspace.local.yaml`（非コミット）を置くと、そのリポだけ別ブランチの worktree パスに差し替わる。

```bash
$ cat > ../workspaces/payments/workspace.local.yaml <<'YAML'
overrides:
  - name: payment-proto
    revision: feature/api-v2
YAML

$ GHQ_ROOT=/tmp/ghq WT_ROOT=/tmp/wt ./bin/workspace-mcp \
    --manifest ../workspaces/payments/workspace.yaml \
    --print-paths --roles source,api-definition
/tmp/ghq/github.com/acme/payment-service
/tmp/wt/github.com/acme/payment-proto/feature/api-v2
```

payment-proto だけ `WT_ROOT` 配下の `.../feature/api-v2` に切り替わる。確認後は `workspace.local.yaml` を消す。

## 2. MCP サーバの確認

stdio の MCP サーバは、initialize ハンドシェイクを経た双方向セッションで話す。生の JSON-RPC を手で流すのは現実的でない（メッセージごとに応答待ちが要り、一括送信すると stdin の EOF でサーバが閉じる）。次の2つを使う。

### 方法A: MCP Inspector（推奨）

公式の検査ツール。サーバを起動し、ツール一覧と各ツールの呼び出しを GUI/CLI で試せる。

```bash
npx @modelcontextprotocol/inspector \
    ./bin/workspace-mcp --manifest ../workspaces/payments/workspace.yaml
```

環境変数を渡したいときは Inspector の起動環境に設定する。

```bash
GHQ_ROOT=/tmp/ghq WT_ROOT=/tmp/wt \
    npx @modelcontextprotocol/inspector \
    ./bin/workspace-mcp --manifest ../workspaces/payments/workspace.yaml
```

確認ポイント

- `workspace_info` — workspace=payments、repoCount=4、activeOverrides が Overlay 数と一致
- `list_repos`（role=source）— payment-service のみ。`cloned` と `path` を確認
- `resolve_paths`（roles=source,api-definition）— `paths` と `addDirCommand`。Overlay 時は worktree パス
- `setup_workspace` — `ghq get` を実行（唯一の副作用。試すなら捨ててよい GHQ_ROOT で）
- `repo_relationships`（name=payment-service）— outgoing/incoming の依存
- `validate_workspace` — `ok`・`errors`・`warnings`

### 方法B: Claude Code に登録して使う

実際の利用形態に近い。絶対パスで登録する。

```bash
claude mcp add workspace \
    -e GHQ_ROOT="$HOME/ghq" \
    -- "$(pwd)/bin/workspace-mcp" --manifest "$(pwd)/../workspaces/payments/workspace.yaml"
```

登録後、セッション内で各ツールを呼んで確認する。構文の詳細は `claude mcp add --help` を参照。

## 3. 自動テスト

手動確認の裏側は自動テストで担保している。

```bash
go test ./...          # ユニット＋シナリオ＋e2e（実バイナリを stdio 起動）
go test -short ./...    # e2e のビルドを省いて高速に
```

- `pkg/workspace` — 検証・所在解決のユニットテスト
- `workspace`（scenario）— インメモリで全ツールを呼ぶ統合テスト
- `workspace`（e2e）— ビルドしたバイナリを `CommandTransport` で起動し本番経路を検証

## トラブルシュート

- **バイナリが見つからない** — `go build -o bin/workspace-mcp ./workspace` を実行。プラグイン運用時は SessionStart フックが `${CLAUDE_PLUGIN_DATA}/bin` へ配置する
- **パスが想定と違う** — `GHQ_ROOT` / `ghq.root` の解決結果を確認。複数 root のときは最後の値が主ルート
- **Overlay が効かない** — `workspace.local.yaml` がマニフェストと同じディレクトリにあるか確認。`--local` で明示も可
