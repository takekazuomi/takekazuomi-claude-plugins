---
name: go-pr-review
description: Go専用PRレビュー。PR情報・コメント・レビュー確認、Issue紐付け・目的記載チェック、Idiomatic Go検証を実施。結果は./tmpにmarkdown出力。
---

# go-pr-review

Go専用のPRレビューを実施するスキル。

## いつ使うか

- GoプロジェクトのPRをレビューする際
- Idiomatic Goに準拠しているか確認したい際
- PRのメタデータ（Issue紐付け、目的記載）を検証したい際

## 使用方法

```text
/go-pr-review <PR番号 or URL>
```

例:

- `/go-pr-review 123`
- `/go-pr-review https://github.com/owner/repo/pull/123`

## レビューフェーズ

### Phase 1: PR情報取得

1. PR番号/URLのパース
   - 番号のみ: 現在のリポジトリのPRとして処理
   - フルURL: owner/repoも抽出

2. PR情報取得

   ```bash
   gh pr view <番号>
   gh pr view <番号> --comments
   gh api repos/{owner}/{repo}/pulls/{番号}/reviews
   ```

### Phase 2: メタデータ検証

以下の項目を検証:

1. **Issue紐付け確認**
   - `Fixes #xxx`、`Closes #xxx`、`Resolves #xxx` 等の参照
   - 紐付けがない場合は警告

2. **変更目的の記載確認**
   - PR本文に目的の説明があるか
   - ドキュメントへのリンクがあるか

3. **目的未記載時の対応**
   - ユーザーに確認メッセージを表示
   - 続行/中断を選択可能

### Phase 3: コード品質検証

1. 変更差分取得

   ```bash
   gh pr diff <番号>
   ```

2. Idiomatic Go検証
   - [docs/idiomatic-go.md](./docs/idiomatic-go.md) に基づくレビュー
   - 以下の観点でチェック:
     - 命名規則（パッケージ名、変数名、頭字語）
     - エラー処理（早期リターン、エラーラップ）
     - インターフェース設計（小さく、消費側で定義）
     - ゼロ値の活用
     - goroutineのライフサイクル管理
     - gofmtの適用

### Phase 4: レビュー結果出力

1. tmpディレクトリ確認
   - `./tmp` が存在しない場合は作成
   - `.gitignore` に `tmp/` がない場合は警告

2. 結果ファイル出力
   - ファイル名: `./tmp/pr-review-{番号}.md`

## 出力フォーマット

```markdown
# PR #<番号> レビュー結果

## 基本情報
- タイトル:
- 作成者:
- ブランチ:
- 状態:

## メタデータ検証
### Issue紐付け
- [x/空] Issue参照あり: #xxx

### 変更目的
- [x/空] 目的が明記されている
- [x/空] ドキュメントへのリンクあり

## コードレビュー
### Idiomatic Go検証
[検証結果と指摘事項]

### コメント・レビュー状況
[既存コメント・レビューのサマリー]

## 総合評価
[総合的な評価とアクション提案]
```

## Idiomatic Goチェックリスト

詳細は [docs/idiomatic-go.md](./docs/idiomatic-go.md) を参照。

主要なチェック項目:

| 観点 | 確認内容 |
| ------ | --------- |
| gofmt | フォーマット適用済みか |
| 命名 | パッケージ名は小文字単語、stuttering回避 |
| エラー処理 | 早期リターン、適切なラップ |
| インターフェース | 小さく、消費側で定義 |
| ゼロ値 | 有効活用されているか |
| Doc Comments | 公開名は名詞で始まるコメント |
| Import | 3グループ分離（標準/外部/内部） |

## 注意事項

- レビュー結果は機械的なチェックのみ
- 最終判断は人間のレビュアーが行う
- `command make lint` / `command make test` の実行は含まない（pr-workflowスキルの責務）
