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

## セットアップ

```bash
# 1. リポジトリをclone
ghq get github.com/takekazuomi/claude-skills

# 2. パーソナルスキルディレクトリ作成
mkdir -p ~/.claude/skills

# 3. シンボリックリンク作成
ln -s ~/ghq/github.com/takekazuomi/claude-skills/bash-script-template ~/.claude/skills/bash-script-template
ln -s ~/ghq/github.com/takekazuomi/claude-skills/go-project-scaffold ~/.claude/skills/go-project-scaffold
ln -s ~/ghq/github.com/takekazuomi/claude-skills/mysql-container-setup ~/.claude/skills/mysql-container-setup
ln -s ~/ghq/github.com/takekazuomi/claude-skills/pr-workflow ~/.claude/skills/pr-workflow
ln -s ~/ghq/github.com/takekazuomi/claude-skills/go-pr-review ~/.claude/skills/go-pr-review
```

## 使い方

Claude Codeで `/bash-script-template`、`/go-project-scaffold`、`/mysql-container-setup`、`/pr-workflow`、`/go-pr-review`コマンドを実行。

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
│       ├── idiomatic-go.md
│       └── go-internal-package-debate.md
├── docs/
│   └── template-management.md  # テンプレート管理方針
└── README.md
```

## ドキュメント

- [テンプレート管理方針](./docs/template-management.md) - スキル内のコードテンプレート管理方法
