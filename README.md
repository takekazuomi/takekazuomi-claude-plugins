# claude-skills

Claude Code用のパーソナルスキル集。

## スキル一覧

| スキル                                                    | 説明                                     |
| --------------------------------------------------------- | ---------------------------------------- |
| [bash-script-template](./bash-script-template/SKILL.md)   | 新規bashスクリプトの作成                 |
| [go-project-scaffold](./go-project-scaffold/SKILL.md)     | 新規Goプロジェクトの初期化               |
| [mysql-container-setup](./mysql-container-setup/SKILL.md) | ローカルテスト用MySQLコンテナ追加        |
| [pr-workflow](./pr-workflow/SKILL.md)                     | worktree作成からPR作成までのワークフロー |
| [go-pr-review](./go-pr-review/SKILL.md)                   | Go専用PRレビュー                         |
| [writing-style](./writing-style/SKILL.md)                 | 日本語文書の文体適用（casual/formal切替）|

## セットアップ

`make install` で全スキルを `~/.claude/skills/` にシンボリックリンク。

```bash
# 1. リポジトリをclone
ghq get github.com/takekazuomi/claude-skills
cd ~/ghq/github.com/takekazuomi/claude-skills

# 2. 全スキルをインストール
make install
```

### Makefileターゲット

| ターゲット       | 説明                           |
| ---------------- | ------------------------------ |
| `make install`   | 全スキルをインストール         |
| `make uninstall` | 全スキルをアンインストール     |
| `make list`      | インストール状態を表示         |
| `make help`      | ヘルプ表示                     |

## 使い方

Claude Codeで `/bash-script-template`、`/go-project-scaffold`、`/mysql-container-setup`、`/pr-workflow`、`/go-pr-review`、`/writing-style`コマンドを実行。

## 構造

```text
claude-skills/
├── bash-script-template/
│   └── SKILL.md
├── go-project-scaffold/
│   └── SKILL.md
├── mysql-container-setup/
│   └── SKILL.md
├── pr-workflow/
│   └── SKILL.md
├── go-pr-review/
│   ├── SKILL.md
│   └── docs/
│       ├── idiomatic-go.md             # Idiomatic Goガイド
│       ├── go-internal-package-debate.md  # internal パッケージ議論
│       ├── doc.md                      # ドキュメント連動チェック観点（作成中）
│       └── sql.md                      # SQL変更時の注意点（作成中）
├── writing-style/
│   ├── SKILL.md                        # ルーター（文書種別を判定しスタイル適用）
│   ├── styles/
│   │   ├── common.md                   # 共通コア（全スタイル共通）
│   │   ├── casual.md                   # レポート/ブログ/Zenn向け
│   │   └── formal.md                   # 仕様書/技術文書向け
│   └── scripts/
│       └── wordcount.sh                # 文字数・原稿用紙換算
├── docs/
│   ├── template-management.md          # テンプレート管理方針
│   ├── idea.md                         # アイデアメモ
│   └── skills/
│       └── writing-style.md            # writing-styleスキルの保守ガイド
├── Makefile
├── CLAUDE.md
└── README.md
```

## ドキュメント

- [テンプレート管理方針](./docs/template-management.md) - スキル内のコードテンプレート管理方法
