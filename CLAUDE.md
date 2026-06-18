# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

Claude Code用のパーソナルスキル集。リポジトリ自体が marketplace 構造を持ち、各スキルを独立した plugin として配布（`/plugin install <skill>@takekazuomi-claude-plugins`）。あわせて開発・個人グローバル用に `~/.claude/skills/` へのシンボリックリンク（`make install`）も併用可能。

## 構造

ディレクトリ構成は [README.md](./README.md#構造) を参照。

## スキルの追加方法

1. `plugins/<スキル名>/skills/<スキル名>/SKILL.md` を作成
2. frontmatterに `name` と `description` を記述
3. `plugins/<スキル名>/.claude-plugin/plugin.json` を作成（`name`・`description`・`version`）
4. `.claude-plugin/marketplace.json` の `plugins` 配列にエントリ追加（`name`・`source`・`description`）
5. README.mdのスキル一覧テーブルと構造ツリーに追加
6. MakefileのSKILLS変数にスキル名を追加
7. 検証: `claude plugin validate .`、`make install` でインストール確認

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
