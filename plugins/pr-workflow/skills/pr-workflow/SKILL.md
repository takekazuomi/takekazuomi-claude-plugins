---
name: pr-workflow
description: worktree作成からPR作成まで一貫実行するワークフロー。ブランチ作成、実装、lint/test、PRサマリー作成を段階的に実施。機能追加やバグ修正でPRを作成する際に使用。
---

# pr-workflow

worktree作成からPR作成まで一貫して実行するワークフロー。

## いつ使うか

- 新機能の追加時（feature/ブランチ）
- バグ修正時（fix/ブランチ）
- リファクタリング時（refactor/ブランチ）

**注意**: 単純な1コミット修正では使用せず、直接mainブランチで作業可能。

## 使用方法

```
/pr-workflow [ブランチタイプ/ブランチ名] [説明]
```

例:
- `/pr-workflow feature/add-login ログイン機能の追加`
- `/pr-workflow fix/null-pointer`
- `/pr-workflow` （対話的に入力）

## ワークフロー

### Phase 1: 初期化

1. リポジトリ情報の取得
   - `git remote get-url origin` からユーザー/リポジトリを取得
   - ghq構造（github.com/ユーザー/リポジトリ）を想定

2. ブランチ名の決定
   - 引数で指定された場合はそれを使用
   - 未指定の場合はユーザーに確認

3. tmpディレクトリの確認・作成
   - worktreeパス配下に `tmp/` を作成
   - `.gitignore` に `tmp/` が含まれているか確認
   - 未設定の場合はユーザーに追加確認

### Phase 2: worktree作成

```bash
git worktree add -b <ブランチ名> ~/wt/github.com/<ユーザー>/<リポジトリ>/<ブランチタイプ>/<ブランチ名> main
```

worktreeパス: `~/wt/github.com/<ユーザー>/<リポジトリ>/<ブランチタイプ>/<ブランチ名>`

### Phase 3: 実装作業

- ファイル編集は絶対パスを使用
- コマンド実行は `cd <worktreeパス> && <コマンド>` 形式
- Plan modeで計画・実装を進める

### Phase 4: コミット（繰り返し可能）

1. 変更をステージング
2. `tmp/commit-msg.md` にコミットメッセージのドラフト作成
3. ユーザーにファイル編集の機会を提供
4. 確認後、コミット実行

```bash
# tmp/commit-msg.md の形式
コミットタイトル（1行目）

本文（詳細説明）
```

5. 追加コミットが必要か確認
   - 必要な場合はPhase 3に戻る
   - 完了の場合はPhase 5へ

### Phase 5: PR作成準備

1. lint/test実行
   ```bash
   cd <worktreeパス> && command make lint && command make test
   ```

2. 失敗時の対応
   - ユーザーに続行するか確認
   - 修正する場合はPhase 3に戻る

3. `tmp/pr-summary.md` にPRサマリー作成
   ```markdown
   ## Summary
   - 変更点1
   - 変更点2

   ## Test plan
   - [ ] テスト項目1
   - [ ] テスト項目2
   ```

4. ユーザーにファイル編集の機会を提供

### Phase 6: PR作成

1. リモートプッシュ
   ```bash
   cd <worktreeパス> && git push -u origin <ブランチ名>
   ```

2. PR作成
   ```bash
   cd <worktreeパス> && gh pr create --title "<タイトル>" --body "$(cat tmp/pr-summary.md)"
   ```

3. 完了通知
   - PR URLを表示
   - worktree削除手順を案内
     ```bash
     git worktree remove ~/wt/github.com/<ユーザー>/<リポジトリ>/<ブランチタイプ>/<ブランチ名>
     ```

## 作業ファイル構造

```
<worktreeパス>/tmp/
├── commit-msg.md       # コミットメッセージ（ユーザー編集可）
├── pr-summary.md       # PRサマリー（ユーザー編集可）
└── workflow-state.md   # ワークフロー状態（Claude管理）
```

## ユーザー確認ポイント

| タイミング | 確認内容 |
|-----------|---------|
| 初期化時 | ブランチ名（引数未指定時） |
| 初期化時 | tmpをgitignoreに追加（未設定時） |
| コミット前 | コミットメッセージの確認 |
| コミット後 | 追加コミットを行うか |
| PR作成前 | PRサマリーの確認 |
| PR作成前 | lint/test失敗時の続行 |

## 注意事項

- PRのマージはユーザーが手動で行う
- マージ後は `git worktree remove` でworktreeを削除
- `command make` を使用（Makefileのalias問題回避）
