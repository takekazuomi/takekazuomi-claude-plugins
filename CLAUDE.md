# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

Claude Code用のパーソナルスキル集。各スキルは `~/.claude/skills/` にシンボリックリンクを作成して使用。

## 構造

```
claude-skills/
├── go-new/SKILL.md       # 新規Goプロジェクト初期化スキル
├── mysql-local/SKILL.md  # ローカルテスト用MySQLコンテナ追加スキル
└── README.md
```

## スキルの追加方法

1. `<スキル名>/SKILL.md` を作成
2. frontmatterに `name` と `description` を記述
3. README.mdのスキル一覧テーブルに追加
4. シンボリックリンク: `ln -s ~/ghq/github.com/takekazuomi/claude-skills/<スキル名> ~/.claude/skills/<スキル名>`

## SKILL.mdフォーマット

```markdown
---
name: スキル名
description: スキルの説明
---

# スキル名

[詳細な手順]
```
