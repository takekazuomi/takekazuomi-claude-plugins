# takekazuomi-claude-plugins

Claude Code用のスキル集。プロジェクトに導入し、チームで共有して使う。

## スキル一覧

| スキル                                                                                       | 説明                                     |
| -------------------------------------------------------------------------------------------- | ---------------------------------------- |
| [bash-script-template](./plugins/bash-script-template/skills/bash-script-template/SKILL.md)   | 新規bashスクリプトの作成                 |
| [go-project-scaffold](./plugins/go-project-scaffold/skills/go-project-scaffold/SKILL.md)      | 新規Goプロジェクトの初期化               |
| [mysql-container-setup](./plugins/mysql-container-setup/skills/mysql-container-setup/SKILL.md) | ローカルテスト用MySQLコンテナ追加        |
| [pr-workflow](./plugins/pr-workflow/skills/pr-workflow/SKILL.md)                              | worktree作成からPR作成までのワークフロー |
| [go-pr-review](./plugins/go-pr-review/skills/go-pr-review/SKILL.md)                           | Go専用PRレビュー                         |
| [workspace](./plugins/workspace/skills/workspace/SKILL.md)                                    | マルチリポ・ワークスペースの定義・所在解決・検証（MCP）|
| [writing-style](./plugins/writing-style/skills/writing-style/SKILL.md)                        | 日本語文書の文体適用（casual＝ブログ・記事、formal＝Issue/PR/Design Doc等の技術文書）|
| [memo-capture](./plugins/memo-capture/skills/memo-capture/SKILL.md)                           | 見つけた知見・アイディアを共通メモリポジトリに残す |

## セットアップ

利用方法は2通り。

### 方法A: /plugin で導入（プロジェクト単位・配布向け）

このリポジトリは marketplace 構造を持ち、各スキルを独立した plugin として配布。利用側プロジェクトの Claude Code で marketplace を登録し、必要なスキルを個別に導入。

```text
/plugin marketplace add takekazuomi/takekazuomi-claude-plugins
/plugin install <skill>@takekazuomi-claude-plugins
```

- プロジェクトのメンバー全員に共有する場合は `--scope project` を付与（例: `/plugin install writing-style@takekazuomi-claude-plugins --scope project`）。
- 一覧確認: `claude plugin list --json --available`、または `/plugin` の Discover タブ。
- `*@takekazuomi-claude-plugins` のようなワイルドカード一括導入は非対応。必要なスキルを個別に指定。

### 方法B: mise run install で導入（開発向け・ローカル symlink）

`mise run install` で全スキルを `~/.claude/skills/` にシンボリックリンク。symlink なのでスキル編集が即反映され、開発時に便利。

```bash
ghq get github.com/takekazuomi/takekazuomi-claude-plugins
cd ~/ghq/github.com/takekazuomi/takekazuomi-claude-plugins
mise run install
```

### mise タスク

| タスク | 説明 |
| ---------------- | ------------------------------ |
| `mise run install`   | 全スキルをインストール             |
| `mise run uninstall` | 全スキルをアンインストール         |
| `mise run list`      | インストール状態を表示             |
| `mise run lint`      | Markdown の lint（markdownlint + textlint） |
| `mise tasks`         | タスク一覧を表示                   |

## 使い方

Claude Code で各スキルをスラッシュコマンドとして実行する。`/bash-script-template`、`/go-project-scaffold`、`/mysql-container-setup`、`/pr-workflow`、`/go-pr-review`、`/workspace`、`/writing-style`、`/memo-capture` が使える。

## 構造

```text
takekazuomi-claude-plugins/
├── .claude-plugin/
│   └── marketplace.json                # marketplace カタログ（全 plugin を列挙）
├── plugins/                            # 各スキル = 独立 plugin
│   ├── bash-script-template/
│   │   ├── .claude-plugin/plugin.json
│   │   └── skills/bash-script-template/SKILL.md
│   ├── go-project-scaffold/
│   │   ├── .claude-plugin/plugin.json
│   │   └── skills/go-project-scaffold/SKILL.md
│   ├── go-pr-review/
│   │   ├── .claude-plugin/plugin.json
│   │   └── skills/go-pr-review/
│   │       ├── SKILL.md
│   │       └── docs/                   # Idiomatic Go等の補助ドキュメント
│   ├── memo-capture/
│   │   ├── .claude-plugin/plugin.json
│   │   └── skills/memo-capture/SKILL.md
│   ├── mysql-container-setup/
│   │   ├── .claude-plugin/plugin.json
│   │   └── skills/mysql-container-setup/SKILL.md
│   ├── pr-workflow/
│   │   ├── .claude-plugin/plugin.json
│   │   └── skills/pr-workflow/SKILL.md
│   ├── writing-style/
│   │   ├── .claude-plugin/plugin.json
│   │   └── skills/writing-style/
│   │       ├── SKILL.md                # ルーター（文書種別を判定しスタイル適用）
│   │       ├── styles/                 # common/casual/formal
│   │       ├── textlint/               # スタイル別 textlint 設定（casual/formal）
│   │       └── scripts/                # wordcount.sh（分量測定）・textlint.sh（機械検査）
│   └── workspace/                      # MCP プラグイン（Go バイナリは外部取得）
│       ├── .claude-plugin/plugin.json
│       ├── .mcp.json                   # MCP サーバ定義
│       ├── hooks/                      # SessionStart でバイナリを取得
│       └── skills/workspace/SKILL.md
├── mcp/                                # workspace-mcp のソース（Go モジュール）
│   ├── pkg/workspace/                  # モデル・検証・所在解決
│   └── workspace/                      # main（MCP サーバ＋validate＋--print-paths）
├── workspaces/                         # ワークスペース定義の例とスキーマ
│   ├── workspace.schema.json
│   └── payments/workspace.yaml
├── docs/                               # リポジトリ文書
│   ├── template-management.md          # テンプレート管理方針
│   ├── idea.md                         # アイデアメモ
│   └── skills/
│       ├── memo-capture.md             # memo-captureスキルの保守ガイド
│       └── writing-style.md            # writing-styleスキルの保守ガイド
├── mise.toml                           # node のバージョンと開発タスク
├── CLAUDE.md
└── README.md
```

## ドキュメント

- [テンプレート管理方針](./docs/template-management.md) - スキル内のコードテンプレート管理方法
