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

リンク先は **実行したディレクトリ**（`mise.toml` のある場所）。worktree で実行すればその worktree を指す。既存のリンクが別の場所を指していれば張り替え、変更があったスキルだけを表示する。symlink 以外の実体が置かれている場合は触らずにエラーとして報告する。

```bash
ghq get github.com/takekazuomi/takekazuomi-claude-plugins
cd ~/ghq/github.com/takekazuomi/takekazuomi-claude-plugins
mise run install
```

### mise タスク

| タスク | 説明 |
| ---------------- | ------------------------------ |
| `mise run install`   | 全スキルをインストール（実行した作業ツリーへ張り替え） |
| `mise run uninstall` | 全スキルをアンインストール         |
| `mise run list`      | インストール状態を表示             |
| `mise run lint`      | Markdown の lint（markdownlint） |
| `mise run lint-fix`  | Markdown の lint と自動修正（markdownlint） |
| `mise run lint:text` | 日本語文書の textlint 検査（writing-style の formal 設定。引数でファイルを絞れる） |
| `mise run setup:node` | textlint を npm で導入（`lint:text` が自動で実行。Windows では導入しない） |
| `mise run test:static` | 静的検査（`claude plugin validate --strict`・スキル登録の整合・shellcheck）。Claude Code の CLI が必要 |
| `mise tasks`         | タスク一覧を表示                   |

注意点:

- **前提は mise のみ。** node・markdownlint-cli2・jq・shellcheck は `mise.toml` の `[tools]` から、textlint は `setup:node`（`npm ci`、`package-lock.json` に固定）から自動で導入する。未信頼の作業ツリーでは、先に `mise trust` を実行する
- **管理方式が 2 つある。** markdownlint-cli2 は mise の tools、textlint は npm で管理する。textlint のルールパッケージは textlint 本体と同じ `node_modules` に置く必要があり、mise の npm backend では解決できないため
- **`lint:text` は `lint` に含めない。** 既存文書に未修正の textlint の指摘が残っているため、推敲支援として個別に実行する
- **Windows ネイティブ（cmd）では一部のタスクが動かない。** mise は Windows のタスクを `cmd /c` で実行する
  - `install`・`uninstall`・`list`・`test:static`：sh 前提のため動かない。WSL か Git Bash で実行する
  - `lint`・`lint-fix`：cmd でも動く定義にしている
  - `lint:text`：Unix 用の定義（`set -f`・`eval`）は cmd で動かないため、Windows 用の別定義（`run_windows`）を使う。Windows では textlint を導入せず、`node_modules\.bin\textlint.cmd` があれば実行し、なければスキップする
  - Windows での実行は未検証

### 開発と利用での mise の扱い

開発は mise を使い、利用（配布したスキルの実行）は mise を前提にしない。スキルは利用者の環境を変えない。

- **`mise.toml` は開発用で、配布先には届かない。** `/plugin install` で入るのは `plugins/<name>/` 配下だけ。スキルのスクリプトは `[tools]` のツールや `[env]` の値（`SKILLS` など）を前提にしない
- **スキルは依存を導入せず、導入手順も案内しない。** `mise use` は最寄りの `mise.toml` を、`npm i -D` は利用者の `package.json` を書き換えるため。依存が無い、または版が違うときは警告だけを出し、同梱の設定を参考として示す。任意の工程は exit 0 でスキップする（例: writing-style の `textlint.sh`）
- **開発時のスクリプトは `mise exec` か `mise run` で呼ぶ。** `mise activate` は利用者によって有効な範囲が違うため前提にしない。どちらも `mise.toml` の `[env]` を読み込み、呼び出し元で設定した同名の環境変数を上書きする
- **`mise trust` を自動実行しない。** 未信頼の `mise.toml` があると、非対話の実行（Claude Code の Bash など）は確認待ちで止まるか失敗する。信頼の判断は利用者に任せる
- **`mise.lock` と `.mise/locks/` をコミットする。** lockfile を有効にしていると `mise run` 時に生成される。npm backend のツール（markdownlint-cli2）の依存は `.mise/locks/` に置かれ、`mise.lock` が digest 付きで参照するため、両方そろえて管理する

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
├── scripts/                            # リポジトリの検査スクリプト
│   └── check-registration.sh           # スキル登録の整合チェック（test:static）
├── mise.toml                           # node のバージョンと開発タスク
├── CLAUDE.md
└── README.md
```

## ドキュメント

- [テンプレート管理方針](./docs/template-management.md) - スキル内のコードテンプレート管理方法
