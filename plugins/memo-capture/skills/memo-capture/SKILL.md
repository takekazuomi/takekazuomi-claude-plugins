---
name: memo-capture
description: >-
  作業中に見つけた知見・手順・アイディアを、共通のメモリポジトリに 1 メモ 1 ファイルで残す。
  「メモしておいて」「メモに残す」「あとで調べたいので控える」と指示されたとき、
  および蓄積したメモを検索・参照したいときに使用する。
  保存先は環境変数・個人設定・git config の順に解決するため、利用者ごとに変えられる。
  明示的に指示されたときだけ動作し、会話中に自動で提案はしない。
---

<!-- 保守者向け: このスキルの設計意図・構造は docs/skills/memo-capture.md を参照（実行時は読まなくてよい） -->

# memo-capture

作業中に得た知見やアイディアを、プロジェクトの外にある共通メモリポジトリへ残すスキル。1 メモ 1 ファイルで書き出し、`INDEX.md` に 1 行を登録する。

## いつ使うか

| 使う | 使わない |
| --- | --- |
| ツールの仕様・設定の落とし穴など、後で調べ直したくなる知見 | そのプロジェクトのコードやドキュメントに書くべき内容 |
| その場で出たアイディア・今後試したいこと | 会話の中で完結し、残す価値のない一時的な事実 |
| 外部資料の所在（URL・ドキュメント・ツール） | API キー・トークン・パスワードなどの秘匿値（保存しない） |
| 蓄積したメモの検索・参照（`--find`） | 特定プロジェクトの機密情報 |

## 原則: 既存メモへの不干渉

メモリポジトリは、別セッションや他メンバーが並行して書き込む共有領域である。

- 別セッションが書いたメモは、**明示指示がない限り内容を読まない・変更しない・削除しない**
- 既定で行う書き込みは 2 つだけ — 新規メモファイルの作成と、`INDEX.md` への 1 行追記
- 既存メモの参照が必要になったら、対象パスを提示して指示を仰ぐ
- 例外は `--find`（利用者が検索を明示的に指示した状態）。この場合もヒットしたファイルだけを読み、書き換えない

## 保存先の解決

先に見つかったものを採用する。

| 優先 | 参照元 | 備考 |
| --- | --- | --- |
| 1 | 環境変数 `MEMO_ROOT` | `~` は展開する |
| 2 | `~/.claude/memo-capture.yaml` の `root:` | 個人設定。リポジトリにはコミットしない |
| 3 | `git config --path --get memo.root` | リポジトリごとに切り替える場合 |
| 4 | `$GHQ_ROOT/github.com/*/memo` の探索 | 1 件に定まるときだけ採用。`GHQ_ROOT` 未設定時は `git config --path --get-all ghq.root` の最後の値 → `~/ghq` の順 |

```bash
resolve_memo_root() {
  local root ghq_root
  root="${MEMO_ROOT:-}"
  if [ -z "$root" ] && [ -f "$HOME/.claude/memo-capture.yaml" ]; then
    root=$(sed -n 's/^root:[[:space:]]*//p' "$HOME/.claude/memo-capture.yaml" | head -1 | tr -d "\"'")
  fi
  if [ -z "$root" ]; then
    root=$(git config --path --get memo.root 2>/dev/null || true)
  fi
  if [ -z "$root" ]; then
    ghq_root="${GHQ_ROOT:-$(git config --path --get-all ghq.root 2>/dev/null | tail -1)}"
    [ -n "$ghq_root" ] || ghq_root="$HOME/ghq"
    root=$(find "${ghq_root/#\~/$HOME}/github.com" -mindepth 2 -maxdepth 2 -type d -name memo 2>/dev/null || true)
  fi
  echo "${root/#\~/$HOME}"
}
```

解決結果が空、複数行、または存在しないディレクトリの場合は、**作成せずに停止**する。候補パスを提示し、`MEMO_ROOT` の設定か `~/.claude/memo-capture.yaml` の作成を案内して指示を待つ。

## 使用方法

```text
/memo-capture <メモしたい内容>   # 内容を指定して記録
/memo-capture                    # 直前の会話から対象を拾い、内容を確認してから記録
/memo-capture --find <キーワード>  # 蓄積したメモを検索
```

## 手順

1. **保存先の解決** — 上記の探索順で決める。未解決なら設定方法を案内して停止。
2. **カテゴリの判定** — 既定は `tips`。

   | 内容 | カテゴリ |
   | --- | --- |
   | 手順・設定・仕様・落とし穴といった確定した知見 | `tips` |
   | 思いつき・今後やりたいこと・未検証の案 | `ideas` |
   | 外部資料・URL・ツールの所在 | `refs` |

3. **衝突の検知** — `ls <root>/<category>/` で**ファイル名だけ**を確認する。既存メモの内容は読まない。同名 slug があれば、そのパスを提示して「別名で新規作成するか、既存メモを読んで追記するか」を確認する。無断での上書き・追記はしない。
4. **メモの作成** — 内容から slug（英小文字の kebab-case）を決め、`<root>/<category>/<slug>.md` を新規作成する。書き込み前に**フルパスを提示して確認を取る**（プロジェクト外への書き込みのため）。
5. **索引の更新と報告** — `INDEX.md` の該当カテゴリ節に 1 行を追記する。既存行は読み込みも書き換えもしない。`INDEX.md` が無ければ、カテゴリ節を持つ雛形を新規作成する。最後に作成したパスを報告する。`git commit` は明示指示があるまで実行しない。

## メモの書式

```markdown
---
title: mise の設定ファイル名
date: 2026-09-06
category: tips
tags: [mise, config]
source: https://github.com/jdx/mise/discussions/2206
project: takekazuomi/takekazuomi-claude-plugins
---

- `mise.toml` を使用
- `.mise.toml` は後方互換のために残る旧形式
```

- `title`・`date`・`category`・`tags` は必須。`source`（出典 URL）と `project`（発生元リポジトリ）は任意
- `date` は記録日を `YYYY-MM-DD` で書く
- 本文は体言止めの箇条書き。後から grep で拾えるよう、固有名詞（ツール名・オプション名・ファイル名）を省略しない

`INDEX.md` は「リンク＋ 1 行フック」を該当カテゴリ節に追記する。ファイルが無いときだけ、この雛形で新規作成する。

```markdown
# メモ索引

## tips

- [mise の設定ファイル名](tips/mise-config-filename.md) — `.mise.toml` は旧形式

## ideas

## refs
```

## 検索・参照

`--find` が指定されたときだけ、既存メモを読む。

1. 保存先を解決する
2. `INDEX.md` をキーワードで絞り込む（タイトルと 1 行フックで大半は当たる）
3. 足りなければ `grep -ril <キーワード> <root>` でファイルを特定する
4. **ヒットしたファイルだけ**を読み、内容と該当パスを提示する。読んだメモは書き換えない

## 注意事項

- 秘匿値（API キー・トークン・パスワード・個人情報）は記録しない。出典 URL と手順だけを残す
- メモリポジトリはプロジェクトの外にあるため、書き込み前に必ずパスを提示して確認を取る
- 1 メモは 1 話題に保つ。話題が 2 つ以上あるときはファイルを分ける
- `git add` / `git commit` / `git push` は明示指示があるまで実行しない
