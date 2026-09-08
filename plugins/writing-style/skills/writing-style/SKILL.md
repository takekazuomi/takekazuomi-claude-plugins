---
name: writing-style
description: >-
  日本語の文書を書き出す・推敲するときに使う。文書種別に応じて文体を切り替える。
  レポート・ブログ（Blogspot）・Zenn記事は casual スタイル（人間味のある親しみやすい散文）、
  Issue・PR の説明・Design Doc・仕様書・設計書・議事録は formal スタイル
  （技術文書。正確さと簡潔さを優先し、結論先出し・能動態・箇条書きは体言止め）を適用する。
  文書の骨子、散文と箇条書きの役割分担、脚注・コマンド例・技術用語の記法を適用し、
  書き終えたあとに引き締めレビュー（チェックリスト＋文字数カウント）を走らせる。
  会話的な短い回答や、コード生成そのものには適用しない。
---

<!-- 保守者向け: このスキルの設計意図・構造は docs/skills/writing-style.md を参照（実行時は読まなくてよい） -->

# 文章スタイル Skill

日本語の文書を書き出すときの文体と手順をまとめたルーター。文書種別を判定し、共通コアとスタイル別ルールを読み込んで適用する。

この Skill は下書きと推敲を助けるが、**最後の言い回しの確定は著者が行う**。Skill の出力をそのまま採用せず、必ず著者が直す。

## 1. 文書種別の判定

対象がどの種別かを判定し、適用するスタイルを決める。

| 文書種別 | スタイル | 読むファイル |
| --- | --- | --- |
| レポート・ブログ・Zenn記事 | casual | `styles/common.md` ＋ `styles/casual.md` |
| Issue・PR の説明・Design Doc・仕様書・設計書・議事録 | formal | `styles/common.md` ＋ `styles/formal.md` |

判定ルール：

1. ユーザーが種別やスタイルを明示していれば、それに従う
2. 明示がなければ文脈から推論する（記事・ブログ・Zenn → casual、Issue・PR・Design Doc・仕様書・設計書 → formal）
3. 推論できない場合のみ、著者に casual / formal のどちらかを確認する

## 2. スタイルの適用

判定したら、**`styles/common.md` を必ず読み、該当する `styles/<種別>.md` を併せて読んで**執筆・推敲する。両スタイルを同時に読み込まない（混在を避ける）。

- 共通コア: [styles/common.md](./styles/common.md)
- casual: [styles/casual.md](./styles/casual.md)
- formal: [styles/formal.md](./styles/formal.md)

## 3. 書き終えたあとのレビュー手順

下書きができたら、この順で走らせる。**削ってから直す**——簡潔化は仕上げではなく、隠れた論理の飛躍を露出させる検査工程である。

1. **削る**：主張に奉仕していない文・語を落とす。「the fact that」的な空語（`generally`、`typically`、`〜することが可能である` 等）を消す
2. **論理を点検**：削って露出した段差（前の文と次の文の間に抜けた論理）を埋める
3. **通し読み**：
   - casual — 声に出して読み、「友達に話すならこう言うか」で各文を点検する。会話に聞こえない箇所を直す
   - formal — 読み手になったつもりで通読し、各段落で「だから何？」が残らないかを点検する。削った結果として生じた論理の段差、指示語の不明瞭さ、数値のない程度表現を直す
4. **文字数を測る**：`scripts/wordcount.sh` で原稿用紙換算を出す。長すぎれば、主軸が2つ以上ないかを疑う。casual は別記事への分割を、formal は詳細の Appendix 送りを検討する（本体は15〜20分で読める分量）

   ```bash
   bash ~/.claude/skills/writing-style/scripts/wordcount.sh path/to/draft.md
   ```

5. **機械検査（任意）**：textlint で表記・冗長表現の機械的な誤りを潰す。導入済みでなければスキップしてよい

   ```bash
   bash ~/.claude/skills/writing-style/scripts/textlint.sh formal path/to/draft.md
   bash ~/.claude/skills/writing-style/scripts/textlint.sh casual path/to/draft.md
   ```

   未導入なら、検査対象プロジェクトのルートで次を実行する。何が足りないかはスクリプトが判定して提案する。

   ```bash
   mise use node@24    # node が無い場合のみ
   npm i -D textlint textlint-rule-preset-ja-technical-writing textlint-rule-no-kangxi-radicals
   ```

   **textlint が見るのは表記と語法だけ**で、論理・構成・根拠は見ない。チェックリストの代わりにはならない。

6. **チェックリスト**：`styles/common.md` の共通チェックリストと、該当スタイルの固有チェックリストを順に確認する
7. **最後の言い回しは著者が確定**する

## 参考（スタイルの背景）

casual の背景：

- AI slop のフィールドガイド: <https://www.ignorance.ai/p/the-field-guide-to-ai-slop>
- stop-slop: <https://github.com/hardikpandya/stop-slop>
- AI authorship signals: <https://zenn.dev/transmedia_blog/articles/ai_authorship_signals>

formal の背景：

- 高品質な技術文書を書くには（riywo）: <https://blog.riywo.com/2021/01/how-to-write-high-quality-technical-doc/>
- Google Technical Writing One: <https://developers.google.com/tech-writing/one>
- Google developer documentation style guide（Lists）: <https://developers.google.com/style/lists>
