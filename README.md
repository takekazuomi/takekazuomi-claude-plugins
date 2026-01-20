# claude-skills

Claude Code用のパーソナルスキル集。

## スキル一覧

| スキル | 説明 |
|--------|------|
| [go-new](./go-new/SKILL.md) | 新規Goプロジェクトの初期化 |
| [mysql-local](./mysql-local/SKILL.md) | ローカルテスト用MySQLコンテナ追加 |

## セットアップ

```bash
# 1. リポジトリをclone
ghq get github.com/takekazuomi/claude-skills

# 2. パーソナルスキルディレクトリ作成
mkdir -p ~/.claude/skills

# 3. シンボリックリンク作成
ln -s ~/ghq/github.com/takekazuomi/claude-skills/go-new ~/.claude/skills/go-new
ln -s ~/ghq/github.com/takekazuomi/claude-skills/mysql-local ~/.claude/skills/mysql-local
```

## 使い方

Claude Codeで `/go-new` や `/mysql-local` コマンドを実行。

## 構造

```
claude-skills/
├── go-new/
│   └── SKILL.md
├── mysql-local/
│   └── SKILL.md
└── README.md
```
