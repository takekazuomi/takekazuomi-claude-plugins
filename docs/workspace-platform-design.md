# マルチリポ・ワークスペース定義基盤 — 設計ドキュメント

対象リポジトリ: `github.com/takekazuomi/takekazuomi-claude-plugins`
対象: Claude Code 専用のコーディングエージェント / 最小から始める
最終更新: 2026-06-23

一連の設計検討の結論をまとめた文書。**確認済みの事実**（URL付き）、**推論**、**決定/推奨**を区別する。反証・限界は脚注に逃さず本文に残す。

---

## 1. 目的

`takekazuomi-claude-plugins`（現状は Skill 単体型プラグイン6本のマーケットプレイス）を、**マルチリポにまたがるコーディングエージェントの「ワークスペース」を定義・共有する基盤**へ育てる。具体的には、複数リポ（マイクロサービス本体・proto/IDL・仕様・スマホアプリ等）の構造・役割・関係を一箇所で定義し、チームで再利用できるようにする。

---

## 2. ワークスペースとは（このプロジェクトでの定義）

このワークスペースは **作業セット型**（`go.work` 相当）である。「どのリポを一緒に編集するか」を束ねるのが役割で、ビルドの再現性は背負わない。再現性は各リポの **git tag**（ソースの版）と **go.sum / lockfile**（依存の版）が担う。west・Android repo がマニフェストに SHA を焼くのは、C/組み込み・多言語混在に言語横断の依存ロック機構が無く、マニフェストが唯一の固定点になるためで、Go エコシステムには当てはまらない。`go.work` をコミットしないのと同じ理由で、本マニフェストも再現性を持たない。

ワークスペースは3つの軸からなる。

| 軸 | 含むもの | 答える問い | 永続性 |
|---|---|---|---|
| **Definition（定義）** | ロール・リポ識別子・関係・スキル束縛 | 何があり、どう関係し、何にどのスキルを使うか | 共有・コミット |
| **Placement（所在）** | clone 先・ワークスペースルート | どこに置くか | ローカル解決（ghq） |
| **Overlay（一時上書き）** | revision のローカル差し替え | 今どのブランチを手元で見るか | ローカル・非コミット |

**決定**: 「Definition はマニフェストに持たせて共有・永続化し、Placement はユーザーごとに解決してマニフェストに焼かない。Overlay は個人のローカルにのみ置き、共有定義を汚さない」という分割線を採る。Definition/Placement の分離は west・Android repo・`go.work`・ghq がいずれも採用する設計、Overlay は `go.work` の `replace`（ローカルで一時的にモジュールを差し替える）に対応する（D8・§5）。

2つの利用シナリオを想定する。

- **シナリオ1（コード結合型）**: ワークスペース定義をリポにコミットし、コードと一緒に運ぶ。clone すれば同じ環境が立ち上がる。
- **シナリオ2（ワークスペース結合型）**: ワークスペース定義は特定コードに紐づかない第一級の存在。clone 後に「都度セットアップ」すると関連リポが芋づる式に取得される。各リポは「自分がどのプロジェクトの一部か」だけを持つ。

**決定**: シナリオ2を主軸に設計する（理由: 利用者の社内システムが既にこの形であり、「ワークスペース定義＝MCPの組み合わせ」という発想がここに直結する）。シナリオ1は各リポが自分の `.claude/` をコミットすれば併用できる。

---

## 3. 設計判断と根拠

決定は一本の論理の連鎖をなす。各決定が次を呼び、アーキテクチャ（§4）へ着地する。

| # | 決定 | 根拠（事実・URL） | 反証/限界（本文に残す） |
|---|---|---|---|
| D1 | **マニフェスト/オーケストレータ方式**（専用ワークスペース定義＋都度clone）を採る | west: マニフェストリポが member 群を定義し `west init`/`west update` で clone（https://docs.zephyrproject.org/latest/develop/west/manifest.html ）。Claude Code issue #44656「env.json は Claude Code 版 docker-compose」（https://github.com/anthropics/claude-code/issues/44656 ） | Claude Code に複数リポを束ねるネイティブ構造はない＝これは合成 |
| D2 | **clone 先をマニフェストに持たせない**（ghq で解決） | Android repo はマニフェストの `path` に絶対パス禁止・ワークスペースルートは init 時にユーザーが決定（https://man.archlinux.org/man/extra/repo/repo-manifest.1.en ）。ghq は URL から `{ghq.root}/{host}/{user}/{repo}` を決定論的に算出、`ghq.root` はユーザーの gitconfig（https://github.com/x-motemen/ghq ）。`go.work` はコミットしないのが通例（https://dev.to/gophers/what-are-go-workspaces-and-how-do-i-use-them-1643 ） | ghq 多重 root 時の主ルート解決に注意（最後＝主） |
| D3 | **ロールは固定 enum にせず、作成時に宣言する自由語彙＋検証** | Android repo `groups`（自由・マニフェスト内・`repo init -g` で取得を絞る、https://gerrit.googlesource.com/git-repo/+/master/docs/manifest-format.md ）。Backstage は `spec.type` 自由定義（https://backstage.io/docs/features/software-catalog/descriptor-format/ ） | Backstage は「Type の Cambrian explosion（爆発的増加）→ 検証が必要」と警告（https://roadie.io/blog/understanding-the-backstage-system-model/ ）。だから宣言＋検証 |
| D4 | **単一ワークスペースディレクトリを作らない**（ghq のフラット配置のまま） | Git はネストしたリポを一級で扱わず embedded-repo 警告になる。コミュニティの virtual monorepo / bootstrap-repo はいずれも「ワークスペースディレクトリは Git リポではない」と明記（https://medium.com/devops-ai/the-virtual-monorepo-pattern-how-i-gave-claude-code-full-system-context-across-35-repos-43b310c97db8 , https://www.iamraghuveer.com/posts/multi-repo-workspace-claude-code/ ） | Claude Code の CLAUDE.md は cwd から `/` 直前まで親を遡って読む（https://github.com/anthropics/claude-code/issues/21138 ）。共通の親が無いと「ワークスペース知識」を置く場所が無くなる→D5へ |
| D5 | **その帰結として MCP が知識層として必須化** | iamraghuveer は3つの相補手段を挙げ、その一つが repo-registry MCP「全リポのパスとメタデータを公開し、Claude がクエリして所在を見つけ任意リポのファイルを読む」（https://www.iamraghuveer.com/posts/multi-repo-workspace-claude-code/ ）。MCP はディレクトリツリーに紐づかない＝場所非依存 | 散ったリポへのファイルアクセスは別途 `--add-dir` が必要（§4） |
| D6 | **revision は追跡ブランチの指定**（既定はデフォルトブランチ）。再現性は持たせない | 作業セット型のため再現性は責務外（§2）。ソースの版は各リポの git tag、ビルドの再現は go.sum/lockfile が担う。west の追跡ブランチモードに相当（https://docs.zephyrproject.org/latest/develop/west/manifest.html ） | 追跡ブランチは時点で内容が変わる。それで良い（再現は各リポの版管理へ委ねる） |
| D7 | **検証は2層**: 構造=JSON Schema、参照整合=Go | 標準 JSON Schema の `enum` は同一文書の別配列の値を参照できない。Ajv 非標準拡張 `$data` でも「配列1要素の代わりには使えない」と明記（https://ajv.js.org/json-schema.html ）。`$data` 自体が非標準（https://ajv.js.org/guide/combining-schemas.html ） | 参照整合（role∈roles 等）は Go バリデータで担保 |
| D8 | **Overlay 層**（revision のローカル一時上書き）を設け、共有定義と分離する | `go.work` はコミットせずローカルで `replace` 差し替え（https://go.dev/ref/mod#workspaces ）。ghq は1リポ1ディレクトリのため、別ブランチの同時保持に git worktree が要る | worktree 管理の手間が増える。最小実装では worktree 作成・削除は手動 |

**この連鎖の要点**: D4（ネスト回避でディレクトリツリーを捨てる）が、D5（MCP を知識層として必須化）を生んだ。失われた「共通ツリー由来のワークスペース文脈」を、場所非依存の MCP が肩代わりする。これは利用者が最初に述べた「ワークスペース定義＝MCPの組み合わせ」に帰着する。D8（Overlay 層）はこの連鎖とは独立に、作業セット型（§2）の帰結として加わる。手元で一時的にブランチを差し替える `go.work` 的な使い方を、共有定義を汚さず実現する。

---

## 4. アーキテクチャ

3つの層と1つの橋からなる。

```
┌─────────────────────────────────────────────────────────────┐
│ 真実の源 (Definition, コミット・共有・永続)                  │
│   workspace.yaml                                             │
│     roles[]         作成時に宣言する自由語彙                 │
│     repos[]         識別子(URL) + role + revision(ブランチ,任意)│  ← 場所/再現性は持たない
│     relationships[] 役割関係 (consumes/implements...)        │
│     skills[]        何をするとき、どのスキル                 │
└───────────────┬─────────────────────────────────────────────┘
                │ 検証: JSON Schema(構造) + Go(参照整合・ロール検証)
                ▼
┌─────────────────────────────────────────────────────────────┐
│ ローカル上書き (Overlay, 非コミット・一時)  ★作業セット型の帰結│
│   workspace.local.yaml  (.gitignore)                        │
│     overrides[]  name + revision(ブランチ)  → go.work replace 相当
│     差し替えたリポは git worktree に展開し、そのパスを指す   │
└───────────────┬─────────────────────────────────────────────┘
                │ override を反映 (差し替え時は worktree パス)
                ▼
┌─────────────────────────────────────────────────────────────┐
│ 物理配置 (Placement, ユーザーごと・非コミット)               │
│   ghq    {ghq.root}/{host}/{user}/{repo}  ← フラット, ネストしない
│   例: ~/ghq/github.com/acme/payment-service                  │
└───────────────┬─────────────────────────────────────────────┘
                │ 所在を解決
                ▼
┌─────────────────────────────────────────────────────────────┐
│ 知識層 (Knowledge, 場所非依存)  ★D4の帰結で必須化            │
│   workspace MCP (Go, 公式 go-sdk)                            │
│     workspace_info / list_repos / resolve_paths /           │
│     setup_workspace / repo_relationships / validate_workspace│
└───────────────┬─────────────────────────────────────────────┘
                │ パスを供給
                ▼
┌─────────────────────────────────────────────────────────────┐
│ アクセス (橋)                                                │
│   claude --add-dir <resolve_paths / --print-paths の出力>    │
│   → 散ったリポを1セッションに取り込み、編集可能にする        │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. 構成要素（成果物）

| ファイル | 役割 | 言語/形式 | 想定パス |
|---|---|---|---|
| `workspace.yaml` | ワークスペース定義（真実の源）の例 | YAML | `workspaces/<name>/workspace.yaml` |
| `workspace.local.yaml` | Overlay（revision のローカル一時上書き）の例 | YAML | 各自のローカル（`.gitignore`） |
| `workspace.schema.json` | 構造検証（型・必須・命名）＋エディタ補完 | JSON Schema 2020-12 | `workspaces/workspace.schema.json` |
| `workspace-mcp`（バイナリ） | **唯一の Go 成果物**。MCP サーバ（知識層）＋ `validate` サブコマンド（CIゲート）＋ `--print-paths`（起動ブートストラップ）を兼ねる | Go（公式 go-sdk + yaml.v3） | ソース: `mcp/`／配布: GitHub Releases |
| `internal/workspace`（パッケージ） | モデルと検証ロジックの共有実装。MCP ツール `validate_workspace` と `validate` サブコマンドが共用 | Go | `mcp/internal/workspace/` |

各要素は同じ `workspace.yaml` モデルを共有する。検証が通ったマニフェストを MCP が読み、ghq で所在を解決して Claude に供給する。

### workspace.yaml スキーマ（要約）

| フィールド | 必須 | 意味 |
|---|---|---|
| `schemaVersion` | ○ | スキーマ版（固定 `1`） |
| `workspace` | ○ | 識別子（DNSラベル） |
| `roles[].id` / `.description` | ○ | **宣言するロール語彙**（固定enumではない） |
| `repos[].name` | ○ | 短い一意エイリアス（関係/スキルで参照） |
| `repos[].repo` | ○ | **ghq解決可能な識別子(URL)**。場所は書かない |
| `repos[].role` | ○ | `roles[].id` のいずれか（参照整合を検証） |
| `repos[].revision` | △ | **追跡ブランチの指定**（既定はデフォルトブランチ）。無指定も正常。再現性は持たない |
| `repos[].optional` | △ | 既定setupでは clone しない（repo の notdefault 相当） |
| `relationships[]` | △ | `from`/`to` は `repos[].name`、`type` 自由形式 |
| `skills[]` | △ | `when`/`use`/`roles`（roles は宣言済みの部分集合） |

### 検証の2層

- **構造層（JSON Schema）**: 型・必須・命名パターン・`additionalProperties:false`。エディタ補完（`# yaml-language-server: $schema=...`）と汎用CI（`check-jsonschema` 等）。
- **参照整合層（Go）**: JSON Schema では原理的に書けない不変条件。
  - **R1（中核）**: 各 `repos[].role` ∈ `{roles[].id}` ← Type 揺れ/爆発を防ぐ要
  - R2: `roles[].id` 一意 / R3: `repos[].name` 一意
  - R4: `relationships` の端点 ∈ `{repos[].name}` / R5: `skills[].roles` ⊆ `{roles[].id}`
  - 警告（CIは止めない）: 未使用ロール。**revision の無指定・ブランチ名は正常**（作業セット型のため警告しない。再現性は各リポの git tag/go.sum が担う）

### MCP ツール面（6つ）

| ツール | 入力 | 返すもの |
|---|---|---|
| `workspace_info` | — | 名前・ロール語彙・リポ数・関係・スキル束縛 |
| `list_repos` | `role?` | URL・ロール・解決済みパス・clone済みか |
| `resolve_paths` | `roles?` | パス一覧＋そのまま使える `--add-dir` 文字列 |
| `setup_workspace` | `roles?`,`update?` | `ghq get` で clone/更新した結果 |
| `repo_relationships` | `name` | 依存の出方向／入方向 |
| `validate_workspace` | `manifest?`（省略時は起動時のマニフェスト） | `ok`・`errors[]`・`warnings[]`（参照整合＋構造の検証結果）|

`resolve_paths`・`setup_workspace` は `roles` と `optional` で取得対象を絞れる（repo の `-g` グループ相当）。`resolve_paths`・`list_repos` は `workspace.local.yaml` の override を反映し、差し替えたリポは worktree パスを返す。

### Overlay（ローカル一時上書き）

`workspace.local.yaml`（`.gitignore`）に `overrides[]`（`name` ＋ `revision`）を書くと、そのリポだけ手元で別ブランチに差し替えられる。`go.work` の `replace` に相当し、共有定義（`workspace.yaml`）は変えない。典型例は、新 API の proto をブランチで作り、それを参照しながら `source` を書く場面。手元でだけ api-definition を差し替え、生成された Go コードを使う。一時的・ローカル限定なら共有に影響しない。

ghq は1リポ1ディレクトリのため、別ブランチを同時に持つには git worktree が要る。差し替えたリポは worktree（`~/wt/...` に ghq 同型配置）へ展開し、`resolve_paths` がその worktree パスを返す。最小実装では worktree 作成は手動で、MCP は override を読んでパス解決に反映するだけ（副作用なし）。worktree の自動作成（`setup_overlay`）は段階2（§10）。

```yaml
# workspace.local.yaml（非コミット）。新 API を手元で試す例。
overrides:
  - name: payment-proto
    revision: feature/api-v2     # 一時的にこのブランチを参照する
```

### リポ構造（2層）

パストラバーサル禁止（プラグインは install 時に自身のサブツリーだけが cache される）のため、MCP コードをプラグイン外から参照できない。ソースの置き場（共有しやすさ）と配布物（自己完結）を分ける。

```
mcp/                          # ソースの真実の源（Go モジュール）。go install の対象
  go.mod
  workspace/                  # MCP サーバ本体（main）
  internal/workspace/         # モデル・検証ロジック（MCP ツールと validate が共用）
plugins/workspace/            # marketplace 配布物。軽量テキストのみ（バイナリを同梱せず clone を軽く保つ）
  .claude-plugin/plugin.json
  .mcp.json                   # command: ${CLAUDE_PLUGIN_DATA}/bin/workspace-mcp
  hooks/hooks.json            # SessionStart → バイナリ取得
  hooks/ensure-binary.sh
  skills/workspace/SKILL.md
```

### ビルドと配布

配布物は **バイナリ1つ（`workspace-mcp`）**。MCP サーバ・`validate` サブコマンド・`--print-paths` を兼ねる。リポにバイナリをコミットせず外部から取得する。`SessionStart` フックが MCP 起動前にバイナリを揃えるため、起動タイムアウト（既定30秒）には当たらない。

| 経路 | 取得元 | Go 要否 | 備考 |
|---|---|---|---|
| **主経路** | GitHub Releases のクロスコンパイル済みバイナリ | 不要 | フックが OS/arch を判定し `${CLAUDE_PLUGIN_DATA}/bin` へ DL・checksum 検証・キャッシュ（更新をまたいで再利用）|
| **フォールバック** | `go install .../mcp/workspace@<ver>`（モジュールプロキシ）| 必要 | Releases に該当 OS/arch が無い・DL 失敗時にフックが試す。Go 開発者向け |

不適だった手段も本文に残す。

- **docker（ghcr / GitHub Packages）**: コンテナ内ではホストの ghq/worktree が見えず、`setup_workspace` の clone がホストに残らず、`resolve_paths` がホストパスを返せない。ローカル FS 操作を本質とするこの MCP と構造的に不適
- **リモート MCP（`type: http`）**: リモートにホストの FS が無く、ローカルパスを供給できない
- **GitHub Packages**: 生バイナリ用レジストリを持たない（npm/maven/container 等のみ）。バイナリは GitHub Releases が自然

CI ゲートは `workspace-mcp validate <file>` サブコマンドを使う（MCP ツール `validate_workspace` はセッション内専用で CI から呼べない）。

---

## 6. ワークフロー

1. **作成**: `workspace.yaml` を作り、ロール語彙を宣言し、リポを識別子(URL)＋role（＋任意で追跡ブランチ）で列挙して **commit**（チームで再利用）。CI で `workspace-mcp validate` を通す。
2. **セットアップ（都度・各自）**: `setup_workspace`（または `ghq get`）で、ロールで絞ったリポを各自の ghq ルートへ clone。メンバーリポには痕跡を残さない。
3. **起動**: 散ったリポを1セッションに入れるため、`--print-paths` でパスを取り `--add-dir` に流す。
   ```bash
   claude --add-dir $(workspace-mcp --manifest workspaces/payments/workspace.yaml \
                                    --print-paths --roles source,api-definition)
   ```
   一時的に別ブランチを参照したいとき（新 API の proto を試す等）は、先に `workspace.local.yaml` に override を書き、そのブランチを worktree に展開しておく。`--print-paths` がその worktree パスを返すので、同じ手順で `--add-dir` に乗る。`go.work` と同じく一時的・ローカル限定で、共有定義は変わらない。
4. **セッション内**: MCP がトポロジ・所在・clone を担い、エージェントは追加済みディレクトリのファイルを読み書きする。

---

## 7. ガバナンス / セキュリティ

- `workspace.yaml`・プラグイン・MCP は **PR レビュー必須**。`workspace.local.yaml` は `.gitignore` 管理のローカル限定で、共有定義には影響しない。
- MCP・ghq はユーザー権限で動く。本 MCP の**唯一の副作用は `setup_workspace` の `ghq get`**（clone/更新）。マニフェストから任意コマンドは実行しない（識別子は exec 引数として渡し、シェルを介さない）。残りは読み取り専用。
- 配布バイナリ（`workspace-mcp`）は `SessionStart` フックで GitHub Releases から取得する。実質的に外部コード取得のため、**バージョン固定＋checksum（できれば署名）検証を必須**とする。
- CI ゲート: `workspace-mcp validate`（参照整合）＋ `check-jsonschema`（構造、任意）。
- 第三者の Skill/MCP を取り込む場合は権限・秘密情報・最小権限を審査（前回調査の Snyk ToxicSkills〈Skill の36.82%に脆弱性〉・CVE-2025-59536 を踏まえる）。

---

## 8. ネイティブ vs 合成 / 反証・未解決

- **ネイティブに存在**: 単一ツリーのモノレポ対応（nested CLAUDE.md、ツリー遡上）、プラグイン/マーケットプレイス、MCP（`--scope project` で `.mcp.json` 共有）、`--add-dir`、フック、設定の階層
- **合成して作る**: 「複数の独立 Git リポを束ねるワークスペース」という単位そのもの、所在解決（ghq）、関連リポの clone、ワークスペース知識の供給（MCP）
- **反証・限界（重要）**:
  - **Claude Code にマルチリポ・ワークスペースのネイティブ構造はない**（issue #44656 は未実装、別要望 #35362「`claude --workspace repo-a repo-b`」は Closed・未実装、https://github.com/anthropics/claude-code/issues/35362 ）。設定のワークスペース継承もない（各リポ独立、https://www.iamraghuveer.com/posts/shared-claude-settings-across-repos/ ）。
  - **公式 Desktop の「workspace」（2026-04-14 再設計）は別概念**。単一リポ内の N 並列セッション＋セッションごと git worktree であり、複数リポを束ねる単位ではない（名前は衝突するが対象が異なる）。
  - **エージェントは `--add-dir` を自分で実行できない**（スラッシュコマンドはエージェントから呼べない）。だから起動時 `--print-paths` 経路が要になる。これは設計上の制約として残る。
  - `--add-dir` でディレクトリを増やすほど検索範囲が広がり効率が落ちる（https://blog.vincentqiao.com/en/posts/claude-code-add-dir/ ）。ロールで絞って必要なリポだけ足す。
  - ghq 多重 root 時の主ルート解決は本基盤では解決しない。リポごとのブランチ管理は Overlay（§5）で一時差し替えのみ扱い、恒常的なブランチ運用は各リポに委ねる。
  - Overlay の worktree は最小実装では手動作成・手動クリーンアップ。MCP 自動化（`setup_overlay`）は未実装（§10）。
- **流動的**: `env.json` 的なネイティブ・マルチリポ構造の採否（#44656）は追跡対象。採用されれば橋渡しの一部は不要になりうる。

### 8.1 既存の類似実装との対比

マルチリポを束ねる既存物との位置づけ。「近さ」は一軸では測れず、**層（アーキ）で近い実装と形態（出荷物）で近い実装が異なる**。各実装は「実行」「共通ツリー」「知識層」のどれか1つに寄り、本基盤の3軸分離（Definition/Placement/Overlay）＋意味モデル（roles/relationships/skills）を同時に扱う設計は確認できなかった。

| 実装 | 形態 | マニフェスト | 所在解決 | アーキ層 | 供給単位 | 本基盤との差 |
|---|---|---|---|---|---|---|
| **Black Dog Labs MCP**（https://blackdoglabs.io/blog/claude-code-decoded-multi-repo-context ） | TS 参考実装（リリース物なし） | `~/.multi-repo-config.json`（name + **path** + `type`） | **path を直書き** | **知識層（MCP）★同層** | **symbol**（load_symbol / trace_dependency） | `type`＝service/library/contracts は**ロールの固定enum版**（D3 が退ける側の実例）。供給がパスでなく symbol（`--add-dir` と逆思想）。共有・clone・overlay・relationship宣言なし、個人の home 設定でクエリ時最適化に閉じる |
| **ttal**（tta-lab/ttal-cli, https://github.com/tta-lab/ttal-cli ） | **単一バイナリ★同形態** | `~/.config/ttal/projects.toml`（name + **path**） | **path を直書き** | 実行オーケストレータ（別カテゴリ） | worktree | D2 と逆（Placement を焼く）。意味層（role/relationship/skill）なし。本基盤が責務外とする実行自動化（manager/worker 2面・Telegram）が主眼 |
| **repo-registry MCP**（iamraghuveer） | MCP | レジストリ | MCP がクエリ | 知識層（MCP） | パス/メタデータ | 本基盤 MCP 層の祖型。意味モデルと Overlay は持たない |
| **bootstrap-repo**（karun.me） | 共通親リポ | リポ内マニフェスト＋context＋tasks | 親ツリー配下 | 共通ツリー | ツリー文脈 | D4 と逆（ネストを作る派）。共通親が無い前提の本基盤とは出発点が逆 |

**要点（二分して読む）**:

- **層・問題意識で最も近いのは Black Dog Labs MCP**。本基盤の中核「ワークスペース定義＝MCP の組み合わせ」と同じ知識層（§4）に同居し、`type`（service/library/contracts）というロール相当を持つ。ただし供給単位が symbol であってパスでない点、共有・clone・overlay・relationship宣言を欠く点で機構は逆向き。その固定 `type` enum は、D3（ロールを自由語彙＋検証にする）が退ける側の具体例として引ける。
- **出荷形態で最も近いのは ttal**。実際に出荷された単一バイナリ＋リポ集合を登録するマニフェスト、という成果物の形が一致。ただしアーキ層は本基盤が責務外とする実行オーケストレータで、知識層ではない。
- 両者とも所在を **path 直書き**（D2 と逆）で、本基盤は ghq 解決で焼かない。この対比が「なぜ Definition/Placement を分離するか（D2）」「なぜ MCP が知識層として必須化するか（D5）」の正当化を補強する。

---

## 9. 参照（信頼階層: 公式 → 公式issue/当事者 → 実装 → 解説）

- Claude Code: plugins-reference / settings / memory / large-codebases（https://code.claude.com/docs/en/ ）、issue #44656・#35362・#21138・#23404・#45323、go-sdk（https://github.com/modelcontextprotocol/go-sdk ）
- マニフェスト系: Zephyr west（https://docs.zephyrproject.org/latest/develop/west/ ）、Android repo（https://gerrit.googlesource.com/git-repo/+/master/docs/manifest-format.md ）、`go.work`（https://go.dev/ref/mod#workspaces ）、VS Code multi-root（https://code.visualstudio.com/docs/editing/workspaces/ ）
- 知識層: Backstage software catalog（https://backstage.io/docs/features/software-catalog/ ）、iamraghuveer repo-registry MCP（https://www.iamraghuveer.com/posts/multi-repo-workspace-claude-code/ ）、karun.me bootstrap-repo（https://karun.me/blog/2026/03/26/structuring-claude-code-for-multi-repo-workspaces/ ）、Black Dog Labs Multi-Repo Context Loading（https://blackdoglabs.io/blog/claude-code-decoded-multi-repo-context ）
- 類似実装: ttal / tta-lab（https://github.com/tta-lab/ttal-cli 、解説 https://dev.to/neil_agentic/how-i-manage-15-repos-with-claude-code-without-losing-my-mind-2ood ）
- ツール: ghq（https://github.com/x-motemen/ghq ）、Ajv/JSON Schema（https://ajv.js.org/json-schema.html ）
- 背景レポート: `workspace-definition-3plans.md`（3案の比較、本ドキュメントの前段）

---

## 10. 次の一手（候補）

1. この MCP を `takekazuomi-claude-plugins` に **`workspace` プラグインとして同梱**（`.mcp.json` ＋ ビルド手順 ＋ 既存スキルとの結線）。
2. `workspace init` 相当の作成支援（対話でロール宣言→`workspace.yaml` 生成→commit）。
3. MCP 拡張: Overlay の worktree 自動作成（`setup_overlay`）と `resolve_paths` の worktree 対応（§5）、`find_skill`（ロール/タスクからスキルを引く）。
4. CI 整備: `workspace-mcp validate`（参照整合）＋ `check-jsonschema`（構造）を PR で実行。リリース CI（goreleaser でクロスコンパイル → GitHub Releases）を組み、`SessionStart` フックの取得経路を確立する。

---

*事実には確認できた URL を付した。反証・限界は脚注ではなく本文に残した。本ドキュメントは設計の現時点の結論であり、Claude Code 側の進展（#44656 等）により更新されうる。*
