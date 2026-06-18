# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

Claude Code用のパーソナルスキル集。各スキルは `~/.claude/skills/` にシンボリックリンクを作成して使用。

## 構造

ディレクトリ構成は [README.md](./README.md#構造) を参照。

## スキルの追加方法

1. `<スキル名>/SKILL.md` を作成
2. frontmatterに `name` と `description` を記述
3. README.mdのスキル一覧テーブルに追加
4. MakefileのSKILLS変数にスキル名を追加
5. `make install` でインストール

## スキル命名規則

- 形式: kebab-case（小文字+ハイフン）
- パターン: 名詞形式（`noun-noun`）
- 例: `bash-script-template`, `go-project-scaffold`, `mysql-container-setup`

## SKILL.mdフォーマット

```markdown
---
name: スキル名
description: スキルの説明
---

# スキル名

[詳細な手順]
```
