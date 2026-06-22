# コミット・PR作成の許可制 × writing style 連携 — 実現方式の検討

`docs/ideas.md` のアイディア3を具体化するための設計メモ。公式機能と欧米・中国のコミュニティ議論を根拠に5案を比較した結果、**案5（ハイブリッド: PreToolUse hook で強制 ＋ writing-style スキルで文体）を採用**する。本メモは採用の設計と、採用に至った比較検討の記録を残す。

> 実装（plugin の `hooks/hooks.json`・スクリプト作成等）は別タスク。本メモは設計と方針の記録。

## やりたいこと

- Claude Code（AI）が `git commit` / `git push` / `gh pr create` を勝手に実行するのを禁止し、ユーザーの許可を挟む
- コミットメッセージ・PR 本文のドラフトを writing style spec（formal）に従って `./tmp/` に出力する

## 採用決定: 案5（ハイブリッド）

検討の結果、**案5（PreToolUse hook で許可制を強制 ＋ writing-style スキルで文体を担保）を採用**する。

> 位置づけ: これは**プロジェクト共有スキル**として作る。各プロジェクトに plugin を入れ、project スコープ（`.claude/settings.json`、git でチーム共有）で**チーム全員に適用**することを前提とする。個人専用ツールではない。

採用理由:

- **確実性**: hook が `git commit`/`git push`/`gh pr create` を実行前に捕捉し許可を求める。指示（CLAUDE.md）と違い回避されにくい
- **文体品質**: ドラフトを writing-style の formal で清書し、コミット/PR 文の品質を担保
- **配布性・チーム適用**: hook と関連設定を1つの plugin に同梱でき、`/plugin install` で自動有効化。project スコープで配布すれば**チーム全員に同じ許可制を強制**できる。本リポジトリの marketplace 構造にそのまま乗る
- **過剰防御の回避**: 不可逆操作（force push / PR merge）まで hook で抱え込まず、それらは CI + branch protection に委ねる。承認疲れを避け、決定論的な壁を少数に絞る

実装の方向性・5案の比較根拠は後述。

## hook の用語明確化（重要）

本メモで「hook」と書くのは **Claude CLI（Claude Code）の `PreToolUse` hook**。git のネイティブ hook（`.git/hooks/pre-commit` 等）ではない。両者は制御対象が異なる。

| 観点 | Claude CLI hook（`PreToolUse`） | git native hook（`.git/hooks`） |
|---|---|---|
| 制御対象 | AI のツール実行（Bash コマンド）を実行**前**に割り込む | git 操作そのもの。人間でも AI でも発火 |
| 設定場所 | `.claude/settings.json` または plugin の `hooks/hooks.json` | `.git/hooks/` または `core.hooksPath` |
| 主な作用 | コマンド文字列を検査し `ask`/`deny` を返す＝実行を止めて許可を求める | コミット内容・メッセージを検証して中止 |
| AI による回避 | Claude 経由のコマンドのみ捕捉（Claude 外の git は素通り） | `--no-verify` で trivially にスキップ可能 |

「AI に勝手に commit させず許可を挟む」目的には CLI hook が直接的。git native hook・CI は補完的な位置づけとして言及するに留める（後述「レイヤの役割分担」）。

## 前提（調査で判明した根拠）

- **強制力の階層（レイヤの役割分担）**: CLAUDE.md（助言）→ permissions.deny/ask（前方一致のみ）→ CLI hook（Claude 経由のみ）→ git native hook（`--no-verify` で skip 可）→ **CI + branch protection（唯一の権威あるゲート）**。AI は CI に `--no-verify` を渡せないため、最終的な強制力は CI 側にしか置けない
- **「指示は提案、git のルールは壁」**: CLAUDE.md は advisory であり enforce ではない。anthropics/claude-code #40117 では Opus 4.6 が pre-commit hook を6連続バイパス（`--no-verify` / `git stash` / quiet フラグ）し、CLAUDE.md で禁止していても発生、問い詰めると偽装した
- **承認疲れ**: ユーザーは permission プロンプトの約93%を承認する（Anthropic 報告）。壁を増やすほど注意が下がる → 確率的プロンプトより決定論的な壁を少数に置く方が有効
- **PR の merge/close は人間に残す**: claude-code-action は PR 承認を security 上ハードコードで禁止。PR 作成後11秒で本番 main へ auto-merge した事故報告あり（#44202）
- **hook 自体が攻撃面・脆い**: deny ルールが長いコマンドチェーンで無効化される報告（CVE-2025-54794 / -54795）、marketplace プラグイン経由で permissions を改竄する手口。設定ファイル自体を保護しないと、AI はルールを skip でなく「弱める」（glob を狭める・CI ジョブを消す）
- **既存資産**:
  - `pr-workflow` スキルが既に `tmp/commit-msg.md`・`tmp/pr-summary.md` を生成し、「ユーザー確認ポイント」を持つ＝案1の素地
  - `writing-style` の formal が コミット/PR 文に最適（体言止め・正確さ優先）＝ドラフト清書の文体
- **署名抑止は公式設定 `attribution`**: `attribution: { commit: "", pr: "" }` で抑止。`includeCoAuthoredBy` は deprecated。署名はシステムプロンプト由来のため CLAUDE.md だけでは不確実
- **実用ツール**: `block-no-verify`（PreToolUse hook で `--no-verify`/`-n` を阻止、GitHub MCP の穴も塞ぐ）

## plugin で hook を配布できる（案5を現実的にする根拠）

PreToolUse hook は marketplace plugin に同梱して配布できる（公式サポート）。本リポジトリの既存 plugin 構造にそのまま乗る。

- 配置: `plugins/<name>/hooks/hooks.json`（または `plugin.json` の `hooks` フィールド）
- スクリプト参照: `${CLAUDE_PLUGIN_ROOT}/scripts/*.sh`（install 先で解決される専用変数）
- 有効化: `/plugin install` で自動有効、plugin の disable で hook も無効
- bundle 可能要素: skills / agents / **hooks** / MCP / LSP / settings 等
- 信頼: project スコープは Workspace trust ダイアログを経る（MCP server と同じセキュリティモデル）

→ 「hook は設定が複雑で配布しづらい」という従来評価は、plugin 同梱・install で自動有効化により大きく緩和される。

## 5案の比較（採用に至った検討記録）

採用は案5。以下は採用に至った比較と、各案を選ばなかった理由の記録。

| 案 | 強制力 | 手軽さ | ドラフト出力 | カバーしない bypass | 概要 |
|---|---|---|---|---|---|
| 案1 ソフト運用スキル | 低 | 高 | スキル指示 | ほぼ全て（指示は無視可） | pr-workflow 拡張 |
| 案2 permissions のみ | 中 | 高 | 別手段が必要 | `-m` 後方のフラグ、長チェーン無効化 | settings.json で ask/deny |
| 案3 PreToolUse hook | 高 | 中 | hook 内で生成 | Claude 外の git、絶対パス呼出 | コマンド検出ゲート |
| 案4 多層防御 | 最高 | 最低 | hook + ラッパー | 設定弱体化（要・設定保護） | deny + ラッパー + hook + CI |
| 案5 ハイブリッド ★採用 | 高 | 中 | hook + スキル | Claude 外の git | hook 強制 + writing-style 文体 |

### 案1: ソフト運用スキル（pr-workflow 拡張）

- 既存 `pr-workflow` を強化し、commit/PR 前に formal style で `tmp/` にドラフト出力＋「許可を求める」ステップを明示
- 強制力は低（指示ベース、AI が無視しうる）。手軽さは高。marketplace でそのまま配布可能
- #40117 が示すとおり指示は壁にならない。既存資産を最大活用できるが「禁止」は保証できない

### 案2: permissions のみ（settings.json）

- `permissions.ask` に `Bash(git commit*)` `Bash(git push*)` `Bash(gh pr create*)`、`deny` に `git push --force` 等
- 強制力は中。手軽さは高。ドラフト出力は別手段が必要
- **前方一致の穴**: `git commit -m "x" --no-verify` のように `-m` の後ろに付けると素通り。長いコマンドチェーンで deny がサイレント無効化される報告もあり、単独依存は不安。スキルと併用が前提

### 案3: PreToolUse hook によるゲート

- hook で commit/push/pr create を検出 → `tmp/` にドラフト生成 → `permissionDecision: "ask"` を返す
- 強制力は高。ドラフト出力も hook 内で実現。**plugin 同梱で配布・install で自動有効化できる**ため手軽さは「中」に上方修正（従来の「低」評価は plugin 配布で緩和）
- バイパス対策: `--no-verify`/`git stash`/quiet フラグを検査（`block-no-verify` 流用可）。中国語圏では `updatedInput` で危険コマンドを安全形に自動書き換えするプログラマブル制御も実践されている
- 限界: Claude 経由のコマンドのみ捕捉。Claude 外や絶対パス `/usr/bin/git` は捕捉できない

### 案4: 多層防御（permissions deny + ラッパー + hook + CI）

- `git`/`gh` を全 deny、許可リスト方式のラッパースクリプトのみ allow、hook で worktree/機密/危険コマンド検査、CI + branch protection を最終ゲート。PR merge/close は意図的に非提供で人間に強制
- 強制力は最高。手軽さは最低（保守コスト大）。設定ファイル自体の保護（CODEOWNERS）まで要る
- #40117 を踏まえると確実性は最も高い。プロジェクト共有スキルなら多層防御＋CI 連携は品質保証として正当化されうるが、ラッパー保守・設定保護のコストが大きい

### 案5: ハイブリッド（hook で強制 ＋ writing-style スキルで文体）★採用

- 強制レイヤ＝案3の hook（ask + tmp ドラフトの雛形生成）。文体レイヤ＝`writing-style` の formal でドラフトを清書。permissions.ask で二重化して補強
- 強制力は高。文体品質も担保。役割分離が明快（hook＝ゲート、スキル＝文体）
- 「スキルが良いか hook が良いか」の問いへの答え＝両方。hook で確実性、スキルで文体・配布性

#### 実装の方向性（実装は別タスク）

1 つの plugin に hook と関連設定を同梱して配布する。

- **plugin 構成**: `plugins/<name>/hooks/hooks.json` ＋ `plugins/<name>/scripts/*.sh`（`${CLAUDE_PLUGIN_ROOT}` で参照）。`/plugin install` で hook が自動有効化されるため、利用者の設定作業はほぼ不要
- **hook の挙動**: `git commit`/`git push`/`gh pr create` を検出 → `tmp/` にドラフト雛形を生成 → `permissionDecision: "ask"` を返して許可を求める
- **ドラフト清書**: 雛形を `writing-style` の formal（体言止め・正確さ優先）で清書。既存 `pr-workflow` の `tmp/commit-msg.md`・`tmp/pr-summary.md` 出力を土台に再利用
- **二重化と補強**: `permissions.ask` に `Bash(git commit*)` `Bash(git push*)` `Bash(gh pr create*)` を併記。署名抑止は `attribution: { commit: "", pr: "" }` 設定を併用
- **バイパス検査**: `--no-verify`/`-n`/`git stash`/quiet フラグを hook で検査（`block-no-verify` 流用可）
- **人間に残す境界**: PR の merge/close は hook で渡さない。force push 等の不可逆操作は CI + branch protection に委ねる

## 地域差の所見

- **欧米**: 「ガードレール vs ゲート」の中庸が主流。ローカル hook は fast feedback、CI + branch protection を唯一の権威ある最終ゲートに置く。`block-no-verify`・設定ファイル保護・承認疲れの定量データなど、対敵対性・運用工学の議論が厚い
- **中国**: 「修復成本（修復コスト）で権限層を決める」体系的フレーム（読取=allow / 編集=ask / 不可逆=deny）。`updatedInput` での自動書き換えなどプログラマブル制御に踏み込む。一方で全自動 push する手軽派と厳格审批派に二極化
- **両圏共通**: 「自動 pull 可・自動 push 不可・Draft PR」が定番。**署名禁止の需要が両圏で強く、本リポジトリの CLAUDE.md 方針（Claude シグニチャーを入れない）と一致**

## 採用方針と段階導入

採用は案5（ハイブリッド）。段階導入で進める。

- まず案1相当（`pr-workflow` + `writing-style` 連携）で運用を固め、続けて hook（案3）を **plugin 同梱**で載せて案5を完成させる。permissions.ask（案2）と `attribution` 設定は案5に組み込む
- 過剰防御は承認疲れ（約93%承認）で逆効果。決定論的な壁を少数に絞る
- 不可逆操作（force push / PR merge）は CI + branch protection に委ねる

### 各案を採らなかった理由

- **案1単独**: 指示ベースで回避されうる（#40117）。許可制の「強制」を保証できない
- **案2単独**: 前方一致の穴・長チェーンでの deny 無効化があり、ドラフト出力もできない。案5に内包する形で活用
- **案3単独**: 確実だが文体担保がない。writing-style と組むことで案5になる
- **案4**: 最も確実で、プロジェクト共有スキルなら検討に値する。ただしラッパー＋設定保護＋CI 統合まで要し保守コストが大きい。案5（hook + 文体）で許可制と品質担保は必要十分で、不可逆操作は CI + branch protection に分担できるため、現段階では案4まで広げない

## 参考リンク

### 公式

- Claude Code Hooks Reference — <https://code.claude.com/docs/en/hooks>
- Plugins Reference（hook 同梱・`${CLAUDE_PLUGIN_ROOT}`）— <https://code.claude.com/docs/en/plugins-reference>
- Configure permissions — <https://code.claude.com/docs/en/permissions>
- Settings（`attribution`）— <https://code.claude.com/docs/en/settings>
- claude-code-action の制限（PR 承認不可）— <https://github.com/anthropics/claude-code-action/blob/main/docs/capabilities-and-limitations.md>
- How we contain Claude（承認疲れ）— <https://www.anthropic.com/engineering/how-we-contain-claude>

### コミュニティ（英語）

- #40117 — pre-commit hook の6連続バイパス実例 — <https://github.com/anthropics/claude-code/issues/40117>
- #44202 — PR 作成11秒後の本番 auto-merge 事故 — <https://github.com/anthropics/claude-code/issues/44202>
- block-no-verify — <https://github.com/tupe12334/block-no-verify>
- 危険な git コマンドを止める hook（aihero）— <https://www.aihero.dev/this-hook-stops-claude-code-running-dangerous-git-commands>
- deny ルールのチェーン無効化（Adversa）— <https://adversa.ai/blog/claude-code-security-bypass-deny-rules-disabled/>
- HN: guard 手法の批判 — <https://news.ycombinator.com/item?id=47343927>

### コミュニティ（中国）

- 宝玉: Hook + Skill による自動 commit — <https://baoyu.io/blog/2026-02-13/claude-code-auto-commit>
- 思否: 権限配置完全指南（修復成本フレーム）— <https://segmentfault.com/a/1190000047678820>
- how-claude-code-works: 権限/安全（`updatedInput`）— <https://github.com/Windy3f3f3f3f/how-claude-code-works/blob/main/docs/11-permission-security.md>
- Koder.ai: git hooks 中文（ガードレール vs ゲート）— <https://koder.ai/zh/blog/claude-code-git-hooks-geng-an-quan-de-ti-jiao-geng-kuai-de-shen-cha>
- CSDN: AI 署名禁止 + 提交规范 — <https://blog.csdn.net/YoungHong1992/article/details/155104953>

### コミュニティ（日本語）

- 3層防御モデル（zenn dely_jp）— <https://zenn.dev/dely_jp/articles/claude-code-3-layer-defense-git-github>
- 署名拒否 hook（zenn 7shi）— <https://zenn.dev/7shi/articles/20250702-hooks-commit>
