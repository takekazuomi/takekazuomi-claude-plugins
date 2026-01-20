# claude-skills

Claude Code用のパーソナルスキル集。

## スキル一覧

| スキル | 説明 |
|--------|------|
| [go-new](./go-new/SKILL.md) | 新規Goプロジェクトの初期化 |
| [mysql-local](./mysql-local/SKILL.md) | ローカルテスト用MySQLコンテナ追加 |
| [pr-workflow](./pr-workflow/SKILL.md) | Plan単位でPRを作成するワークフロー |

## セットアップ

```bash
# 1. リポジトリをclone
ghq get github.com/takekazuomi/claude-skills

# 2. パーソナルスキルディレクトリ作成
mkdir -p ~/.claude/skills

# 3. シンボリックリンク作成
ln -s ~/ghq/github.com/takekazuomi/claude-skills/go-new ~/.claude/skills/go-new
ln -s ~/ghq/github.com/takekazuomi/claude-skills/mysql-local ~/.claude/skills/mysql-local
ln -s ~/ghq/github.com/takekazuomi/claude-skills/pr-workflow ~/.claude/skills/pr-workflow
```

## 使い方

Claude Codeで `/go-new`、`/mysql-local`、`/pr-workflow` コマンドを実行。

## 構造

```text
claude-skills/
├── go-new/
│   └── SKILL.md
├── mysql-local/
│   └── SKILL.md
├── pr-workflow/
│   └── SKILL.md
├── docs/
│   └── template-management.md  # テンプレート管理方針
└── README.md
```

## ドキュメント

- [テンプレート管理方針](./docs/template-management.md) - スキル内のコードテンプレート管理方法
