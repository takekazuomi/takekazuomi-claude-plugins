# writing-style スキル 保守ガイド

このドキュメントは **writing-style スキルを更新・改善する保守者向け**。スキルを*使う*ためのものではない（使い方は各 `SKILL.md` / `styles/*.md` が持つ）。設計の構造と「なぜそうしたか」を残し、将来の変更判断を助ける。

> 配置の意図: このファイルはリポジトリの `docs/` 配下に置き、**スキルディレクトリには含めない**。`make install` の symlink はスキルディレクトリ単位のため、`docs/` は `~/.claude/skills/` に露出せず、スキル実行時にロードされない。詳細は末尾「なぜスキル外に置くか」を参照。

## 構造

```text
writing-style/
├── SKILL.md            # ルーター（薄い）：文書種別を判定し、共通＋該当スタイルを読ませ、レビュー手順を走らせる
├── styles/
│   ├── common.md       # 全スタイル共通コア（約80%）
│   ├── casual.md       # casual 固有デルタ（レポート/ブログ/Zenn）
│   └── formal.md       # formal 固有デルタ（仕様書/技術文書）
└── scripts/
    └── wordcount.sh    # 文字数・原稿用紙換算（実行のみ、読み込まない）
```

### 各ファイルの責務

| ファイル | 責務 | ロードされる条件 |
| --- | --- | --- |
| `SKILL.md` | ルーティング（種別判定）＋共有レビュー手順＋frontmatter（name/description） | description は常時、本体はスキル起動時 |
| `styles/common.md` | 段落構成・接続詞・文末多様化・技術記述・禁止事項・共通チェックリスト | 起動後、両スタイルとも必ず読む |
| `styles/casual.md` | 人間味・体験談・3部構成・脚注設計・casual固有チェック | casual 判定時のみ |
| `styles/formal.md` | 正確さ優先・箇条書き体言止め・固有比喩・formal固有チェック | formal 判定時のみ |
| `scripts/wordcount.sh` | 分量測定 | レビュー手順で実行（内容は読まない） |

## なぜこの構造か

### なぜ「ルーター型単一スキル」か（別スキル2本にしない理由）

文書種別で文体を切り替えたいが、casual と formal の**実体は約80%が同一**だった（段落・接続詞・文末・技術記述・禁止事項・チェックリストが共通）。差分は「基本方針の力点」「casual=感情表現/3部構成 ⇔ formal=体言止め箇条書き/固有比喩」「記事構成の有無」に限られる。

- 別スキル2本にすると、この共通80%が二重管理になり、`wordcount.sh` も複製または脆いパス結合が必要になる（`make install` は「1ディレクトリ=1スキル=symlink」で、バンドルはスキルディレクトリ内に置く必要があるため）。
- 単一スキル＋共通コア＋デルタなら、共有資産を1か所に集約でき、DRY を保てる。

公式の裏付け: single responsibility は「スキル間」の原則であり、**相互排他・併用しない文脈は別ファイルへ分割してトークンを節約せよ**と公式が明言している（[best-practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)）。ルーター型（必要なスタイルだけ読む）は許容・推奨される構成。

### なぜ「common ＋ デルタ」か（各 style 自己完結にしない理由）

重複が大きい（80%）ため、各 style を自己完結させると共通ルール改修時に2ファイルを直す必要があり、片方の直し忘れが起きる。共通コアを `common.md` に一本化し、各 style は差分のみに保つことで保守点を1か所に絞った。

### なぜ voice bleed が起きないか

casual と formal は文体が正反対（砕け↔正確さ優先）。progressive disclosure により、**使用時は `common.md` ＋ 片方の style だけがコンテキストに載る**。両スタイルが同時にロードされないため、casual ルールが formal 文書に混入する事故を構造的に防げる。

### なぜ `wordcount.sh` は scripts/ か

スクリプトは「実行するだけで内容はロードしない」層（progressive disclosure の第3段）。文字数測定ロジックを本文に書かず scripts/ に逃がすことで、SKILL.md を薄く保ちトークンを節約する。

## 更新の指針

- **共通ルールの変更** → `styles/common.md` のみ。casual/formal 両方に効く。
- **片方の文体だけ変更** → 該当する `styles/casual.md` か `styles/formal.md`。
- **新しい文書種別の追加**（例: メール、プレゼン） → `styles/<種別>.md` を追加し、SKILL.md の「文書種別の判定」表に1行追加。common は触らない。
- **description の変更** → 自動ディスパッチの精度に直結する。種別キーワード（レポート/ブログ/Zenn/仕様書/技術文書）を維持し、「会話的短文・コード生成は除外」の但し書きを残す。
- **チェックリストの追加** → 共通なら common、種別固有なら該当 style の「固有チェックリスト」へ。
- 変更後は `npx markdownlint-cli2 "writing-style/**/*.md"` でlintを通す（リポジトリ規約）。

## なぜスキル外（docs/）に置くか

このガイドは「スキルを実行するときには不要、更新するときだけ要る」情報。スキル実行時のコンテキストに載せたくない。

- スキル起動時にロードされるのは `name`+`description`（常時）、`SKILL.md` 本体（起動時）、SKILL.md から参照され Claude が読むと判断したファイルのみ。**未参照ファイルはゼロトークン**（[best-practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)）。
- ただし、同梱ファイルはプロジェクト探索中に誤って読まれトークンを浪費する事例が報告されている（[anthropics/skills #763](https://github.com/anthropics/skills/issues/763)）。「非参照」は*デフォルトで読まれない*を保証するが*絶対に読まれない*ではない。
- `.skillignore` 的な公式の除外機構は存在しない（[agentskills.io spec](https://agentskills.io/specification)）。SKILL.md 本文への人間専用コメント機能も未実装（[#177](https://github.com/anthropics/skills/issues/177)）。
- → 最も確実な隔離は**物理的にスキル外へ置く**こと。`make install` の symlink はスキルディレクトリ単位なので、`docs/` 配下は `~/.claude/skills/` に露出せず、誤読リスクがゼロになる。リポジトリ既存の `docs/template-management.md` と同じ運用。
