# claude-skills

Claude Code用のパーソナルスキル集。

## スキル一覧

| スキル | 説明 |
|--------|------|
| [go-new](./go-new/SKILL.md) | 新規Goプロジェクトの初期化 |

## セットアップ

```bash
# 1. リポジトリをclone
ghq get github.com/takekazuomi/claude-skills

# 2. パーソナルスキルディレクトリ作成
mkdir -p ~/.claude/skills

# 3. シンボリックリンク作成
ln -s ~/ghq/github.com/takekazuomi/claude-skills/go-new ~/.claude/skills/go-new
```

## 使い方

Claude Codeで `/go-new` コマンドを実行。

## 構造

```
claude-skills/
├── go-new/
│   └── SKILL.md
└── README.md
```
