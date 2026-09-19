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

書き始める前に、読み手・アウトプット・課題・主張を1行ずつ書き出し、著者に確認する（`styles/common.md` の「書き始める前に決めること」）。既存の下書きを推敲する場合は、下書きから読み取った4点を示して確認する。主張の無い文書（調査レポートやまとめなど）は「主張なし」と確認し、主張を作らない。

## 3. 書き終えたあとのレビュー手順

下書きができたら、この順で走らせる。**削ってから直す**。簡潔化は仕上げに見えるが、隠れた論理の飛躍を露出させる検査工程である。

進捗をこのチェックリストに写して、済んだものから消していく。

```text
推敲の進捗:
- [ ] 1 削る
- [ ] 2 論理を点検
- [ ] 3 根拠と突き合わせる
- [ ] 4 通し読み
- [ ] 5 文字数を測る
- [ ] 6 機械検査
- [ ] 7 境界への照らし戻し（`##` の見出しが3つ以上の文書だけ）
- [ ] 8 チェックリスト（未達があれば該当工程へ戻る）
- [ ] 9 著者に示して確定
```

1. **削る**：主張に奉仕していない文・語を落とす。「the fact that」的な空語（`generally`、`typically`、`〜することが可能である` 等）を消す
2. **論理を点検**：削って露出した段差（前の文と次の文の間に抜けた論理）を埋める
3. **根拠と突き合わせる**：数値・固有名詞・引用を1つずつ拾い、実測値か確認済みの出典と照合する。リンクは開いて、引用した内容が書かれているかを確かめる。対応が無いものは落とすか `TODO(未確認: ...)` に置き換える。著者の主張には TODO や限定を付けない（`styles/common.md` の「根拠と断定の強さを揃える」）
4. **通し読み**：
   - casual：声に出して読み、「友達に話すならこう言うか」で各文を点検する。会話に聞こえない箇所を直す
   - formal：読み手になったつもりで通読し、各段落で「だから何？」が残らないかを点検する。削った結果として生じた論理の段差と、指示語の不明瞭さを直す
5. **文字数を測る**：`scripts/wordcount.sh` で原稿用紙換算を出す。長すぎれば、主軸が2つ以上ないかを疑い、分割を検討する

   ```bash
   bash ${CLAUDE_SKILL_DIR}/scripts/wordcount.sh path/to/draft.md
   ```

6. **機械検査**：`scripts/lint.sh` は必ず走らせる。textlint は導入済みなら走らせる

   ```bash
   bash ${CLAUDE_SKILL_DIR}/scripts/lint.sh path/to/draft.md
   bash ${CLAUDE_SKILL_DIR}/scripts/textlint.sh formal path/to/draft.md
   bash ${CLAUDE_SKILL_DIR}/scripts/textlint.sh casual path/to/draft.md
   ```

   `lint.sh` の ERROR は落とす。WARN は誤検出があり得るので、自分で直さず該当箇所を著者に示す。TODO の件数は残っていてよい。禁止語そのものを列挙している箇所は `<!-- lint-off -->` と `<!-- lint-on -->` で囲んで除外する。

   textlint は利用者の環境にあるものを使う。textlint の導入や `package.json`・`mise.toml` の変更はしない。textlint が無い、ルールを読み込めない、本体の版が違うといった場合、スクリプトは警告と同梱のルール設定（`textlint/casual.json`・`formal.json`）の所在を表示する。警告はそのまま利用者に伝える。

   どちらも見るのは表記と語法と語句の型だけで、論理・構成・根拠は見ない。

7. **境界への照らし戻し**（`##` の見出しが3つ以上の文書だけ）：各節の役割を1行で書き出し、書き始める前に確認した課題・主張とどう対応するかを添えて著者に示す。節を残すかどうかは著者が決める。AI の判断で節を削らない
8. **チェックリスト**：`styles/common.md` の共通チェックリストと、該当スタイルの固有チェックリストを確認する。**未達が1つでもあれば、9 へ進まない**。未達の項目を原稿の該当箇所とセットで書き出し、下の表で戻り先の工程を決め、その工程だけをやり直してから 8 に戻る

   | 未達の項目 | 戻る工程 |
   | --- | --- |
   | 書き始める前の確認 | 2節（著者に確認する） |
   | 照合、リンク先、TODO、主張を作った・弱めた | 3 |
   | 表記ゆれ | 8 の中で直す |
   | `lint.sh`・textlint の ERROR | 6 |
   | 骨子と結論の位置（formal） | 2 |

   同じ項目で3周しても抜けない場合は、文の直しでは解決しない。材料が足りていない可能性が高いので、何が不足しているかを添えて著者に返す。

9. **著者に示して確定**：次を著者に渡し、言い回しと判断を確定してもらう。lint やチェックリストを通ったことを完了とみなさない
   - `lint.sh` の WARN の箇所
   - 直截な言い換えがありそうな比喩の候補
   - 残した TODO の一覧
   - 工程 7 で書き出した節の役割（該当する場合）

## 参考（スタイルの背景）

共通の背景：

- mannered prose の定義（Anthropic / Prompting Claude Fable 5.1）: <https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-fable-5-1>
- 埋め草が生じる機構（不確実性より当て推量に報酬が付く）: Kalai et al., *Why Language Models Hallucinate*, arXiv:2509.04664 <https://arxiv.org/abs/2509.04664>

casual の背景：

- AI slop のフィールドガイド: <https://www.ignorance.ai/p/the-field-guide-to-ai-slop>
- stop-slop: <https://github.com/hardikpandya/stop-slop>
- AI authorship signals: <https://zenn.dev/transmedia_blog/articles/ai_authorship_signals>

formal の背景：

- 高品質な技術文書を書くには（riywo）: <https://blog.riywo.com/2021/01/how-to-write-high-quality-technical-doc/>
- Google Technical Writing One: <https://developers.google.com/tech-writing/one>
- Google developer documentation style guide（Lists）: <https://developers.google.com/style/lists>
