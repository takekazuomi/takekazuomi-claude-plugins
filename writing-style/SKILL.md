---
name: writing-style
description: >-
  日本語の文書を書き出す・推敲するときに使う。文書種別に応じて文体を切り替える。
  レポート・ブログ（Blogspot）・Zenn記事は casual スタイル（人間味のある親しみやすい散文）、
  仕様書・技術文書は formal スタイル（正確さ優先・箇条書きは体言止め）を適用する。
  著者の散文スタイル、記事の構成パターン、脚注・コマンド例・技術用語の記法を適用し、
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
| 仕様書・設計書・技術文書 | formal | `styles/common.md` ＋ `styles/formal.md` |

判定ルール：

1. ユーザーが種別やスタイルを明示していれば、それに従う
2. 明示がなければ文脈から推論する（記事・ブログ・Zenn → casual、仕様書・設計書 → formal）
3. 推論できない場合のみ、著者に casual / formal のどちらかを確認する

## 2. スタイルの適用

判定したら、**`styles/common.md` を必ず読み、該当する `styles/<種別>.md` を併せて読んで**執筆・推敲する。両スタイルを同時に読み込まない（混在を避ける）。

- 共通コア: [styles/common.md](./styles/common.md)
- casual: [styles/casual.md](./styles/casual.md)
- formal: [styles/formal.md](./styles/formal.md)

## 3. 書き終えたあとのレビュー手順

下書きができたら、この順で走らせる。**削ってから直す**——簡潔化は仕上げではなく、隠れた論理の飛躍を露出させる検査工程である。

1. **削る**：主張に奉仕していない文・語を落とす。「the fact that」的な空語（generally、typically、〜することが可能である 等）を消す
2. **論理を点検**：削って露出した段差（前の文と次の文の間に抜けた論理）を埋める
3. **音読**：声に出して読み、「友達に話すならこう言うか」で各文を点検する。会話に聞こえない箇所を直す
4. **文字数を測る**：`scripts/wordcount.sh` で原稿用紙換算を出す。長すぎれば、主軸が二つ以上ないかを疑い、別記事への分割を検討する

   ```bash
   bash ~/.claude/skills/writing-style/scripts/wordcount.sh path/to/draft.md
   ```

5. **チェックリスト**：`styles/common.md` の共通チェックリストと、該当スタイルの固有チェックリストを順に確認する
6. **最後の言い回しは著者が確定**する

## 参考（スタイルの背景）

- AI slop のフィールドガイド: <https://www.ignorance.ai/p/the-field-guide-to-ai-slop>
- stop-slop: <https://github.com/hardikpandya/stop-slop>
- AI authorship signals: <https://zenn.dev/transmedia_blog/articles/ai_authorship_signals>
