# writing-style スキル 保守ガイド

このドキュメントは **writing-style スキルを更新・改善する保守者向け**。スキルを*使う*ためのものではない（使い方は各 `SKILL.md` / `styles/*.md` が持つ）。設計の構造と「なぜそうしたか」を残し、将来の変更判断を助ける。

> 配置の意図: このファイルはリポジトリの `docs/` 配下に置き、**スキルディレクトリには含めない**。`mise run install` の symlink はスキルディレクトリ単位のため、`docs/` は `~/.claude/skills/` に露出せず、スキル実行時にロードされない。詳細は末尾「なぜスキル外に置くか」を参照。

## 構造

```text
writing-style/
├── SKILL.md            # ルーター（薄い）：文書種別を判定し、共通＋該当スタイルを読ませ、レビュー手順を走らせる
├── styles/
│   ├── common.md       # 全スタイル共通コア（最小核）
│   ├── casual.md       # casual デルタ（レポート/ブログ/Zenn）
│   └── formal.md       # formal デルタ（Issue/PR/Design Doc/仕様書）
├── textlint/
│   ├── casual.json     # 機械ルールのみ（文体には介入しない）
│   └── formal.json     # preset-ja-technical-writing ＋ 実測にもとづく調整
└── scripts/
    ├── wordcount.sh    # 文字数・原稿用紙換算（実行のみ、読み込まない）
    └── textlint.sh     # textlint ラッパー（実行のみ、読み込まない）
```

### 各ファイルの責務

| ファイル | 責務 | ロードされる条件 |
| --- | --- | --- |
| `SKILL.md` | ルーティング（種別判定）＋共有レビュー手順＋frontmatter（name/description） | description は常時、本体はスキル起動時 |
| `styles/common.md` | 接続詞・文の自然な展開・引用の根拠URL・技術用語/コマンド例/ファイル参照の記法 | 起動後、両スタイルとも必ず読む |
| `styles/casual.md` | 人間味・体験談・3部構成・脚注設計・段落と文末の散文ルール・casual固有チェック | casual 判定時のみ |
| `styles/formal.md` | 技術文書。骨子・結論先出し・能動態・冗長さの排除・箇条書きの役割分担・検証可能性・formal固有チェック | formal 判定時のみ |
| `textlint/*.json` | スタイル別の機械検査ルール | textlint.sh が読む（Claude は読まない） |
| `scripts/wordcount.sh` | 分量測定 | レビュー手順で実行（内容は読まない） |
| `scripts/textlint.sh` | 機械検査の起動と textlint の探索 | レビュー手順で実行（内容は読まない） |

## なぜこの構造か

### なぜ「ルーター型単一スキル」か（別スキル2本にしない理由）

文書種別で文体を切り替えたいが、casual と formal は共通部分を持つ（接続詞による論理の明示、文の自然な展開、引用の根拠URL、技術用語・コマンド例・ファイル参照の記法）。加えてレビュー手順と `wordcount.sh` を共有する。

- 別スキル2本にすると、この共通部分が二重管理になり、`wordcount.sh` も複製または脆いパス結合が必要になる（`mise run install` は「1ディレクトリ=1スキル=symlink」で、バンドルはスキルディレクトリ内に置く必要があるため）。
- 単一スキル＋共通コア＋デルタなら、共有資産を1か所に集約でき、DRY を保てる。

なお 2026-09-08 の改稿前は「共通が約80%」と書いていたが、これは formal が実質 casual の写しだったことの裏返しであり、設計の妥当性の根拠ではなかった（後述）。改稿後の共通部分は縮小しているが、共有レビュー手順とスクリプトがあるためルーター型は維持している。

公式の裏付け: single responsibility は「スキル間」の原則であり、**相互排他・併用しない文脈は別ファイルへ分割してトークンを節約せよ**と公式が明言している（[best-practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)）。ルーター型（必要なスタイルだけ読む）は許容・推奨される構成。

### なぜ「common ＋ デルタ」か（各 style 自己完結にしない理由）

各 style を自己完結させると、共通ルールの改修時に2ファイルを直す必要があり、片方の直し忘れが起きる。共通コアを `common.md` に一本化し、各 style は差分のみに保つことで保守点を1か所に絞った。

ただし**共通に置くのは、両スタイルで方針が一致するものだけ**とする。方針が正反対の項目（段落の長さ、箇条書きの比率など）を common に置くと、片方のスタイルに誤ったルールが読み込まれ続ける。実際にそれが起きたのが 2026-09-08 の改稿である。

### なぜ formal を技術文書として全面改稿したか（2026-09-08）

改稿前の `formal.md` は「仕様書・技術文書」を対象と宣言しながら、実体が casual の写しだった。基本方針が「親しみやすい文章」「読者との距離感を近く」、文末が「〜してる」、比喩が「魔窟化」「難儀する」、時間表現が「最近」「いつのまにか」を**推奨**、記事構成が「きっかけ・動機」のブログ型導入、参考文献に「素晴らしい」「お勧めですよ」の主観評価。formal 固有の差分は「箇条書きの行末に句点を付けない」ほぼ1点だけだった。

利用者からの申告は「formal の文章が冗長」だったが、冗長さは症状にすぎない。駆動源は `common.md` にあった散文の味付けルール（1段落3-5文、文末表現の多様化、断言調を避けて説明調を混ぜる、文の長さのバリエーション）である。これらは読み物の質を上げる一方、技術文書では水増しとして働く。

そこで formal を Issue・PR・Design Doc・仕様書・議事録を対象とする技術文書スタイルへ書き直した。土台としたのは次の3系統。

1. 実務で運用してきた技術文書の作法から、文書骨子、frontmatter と絶対日付、決定・根拠・未決事項、真実の源へのリンク、能動態、用語の一貫性、散文とリストの併用ルール
2. Google のテクニカルライティング（Technical Writing One、developer documentation style guide）から、1文1アイデア、不要語の削除、リストの並列構造と番号付きの命令形動詞始まり
3. [高品質な技術文書を書くには（riywo, 2021）](https://blog.riywo.com/2021/01/how-to-write-high-quality-technical-doc/) から、次の観点

   - 執筆前に読み手・アウトプット・課題を定める手順
   - skills-forward の回避
   - 具体的な数値で書くこと
   - MECE な選択肢と推奨案の明示
   - one-way / two-way door の区別
   - 本体15〜20分と Appendix 送り
   - 仕上げのチェック観点（typo・壊れたリンク・コピペ残骸）

**第3のスタイル（`technical.md`）を追加する案は採らなかった**。formal と守備範囲が重なり、「仕様書はどちらか」がルーティングで曖昧になるうえ、壊れた formal を残すことになるため。

### formal の規定を追加・調整した点（2026-09-08）

全面改稿のあと、運用しながら次の5点を追記・調整した。

- **段落の規定を「1段落1アイデア」へ** — 改稿前の formal.md は「段落は1-2文」だった。[Google developer documentation style guide](https://developers.google.com/style/paragraph-structure) が文数の下限を置かず、1文段落も単一アイデアなら6文超も認めていることを根拠に、判断軸を文数からアイデア数へ移した。5-6文を超えたら詰め込みすぎというサインは目安として残す
- **一文150字の上限を明記** — `textlint.sh` の `sentence-length` 設定と本文の規定が対応していなかった
- **未決事項の呼称を「要確認」「未定」に統一** — 改稿前は3章が「未決事項」「確認事項」、6章が「要確認」「未定」で表記ゆれが起きていた
- **議事録の引用は原文の文体を残す** — 「です・ます」と混在させない規定の例外。引用を改変しない側を優先する
- **絶対日付を求める理由の明示** — 禁止を列挙するだけでは、例外にあたるかどうかを読み手が判断できない

採らなかったのは次の3点。技術文書の作法として提案されうるが、本スキルは意図的に分岐させている。

| 検討した規定 | formal での扱い | 理由 |
| --- | --- | --- |
| 箇条書きを50%未満に制限 | 数値上限を課さない | Issue・PR は本来リスト主体の文書であり、上限は不自然な制約になる |
| 砕けた文末（「〜してる」）を許容 | 禁止を維持 | formal 改稿の目的そのものを取り消すことになる。砕けた文体が要る文書は casual で書く |
| 「です・ます」の混在を許容 | 混在させない | 引用の例外だけ取り込み、本文の規定は変えない |

本スキルは配布先を選ばない。特定の組織や現場の運用を前提とした規定はそのまま持ち込まず、**根拠を示せるものだけを選別して取り込む**。

### なぜ common を最小核へ縮小したか（2026-09-08）

上記の改稿にあわせて、`common.md` から段落の長さ、箇条書きの比率、文末表現の多様化、文の長さのバリエーション、断言調と説明調のバランス、読者との距離感、casual 寄りの禁止事項（です・ます混在OK、賞賛姿勢、失敗談を隠さない）を `casual.md` へ移した。

これらは「共通」に見えて実際は casual の作法であり、common に置く限り formal にも読み込まれ続ける。共通に残したのは、接続詞による論理の明示、前の文からの自然な展開、引用の根拠URL、技術用語・コマンド例・ファイル参照の記法のみ。

段落の長さと箇条書きの比率は**スタイルごとに方針が正反対**（casual は1段落3-5文・箇条書き30%以下、formal は1-2文・比率上限なし）のため、common では規定せず各スタイルに委ねる旨を明記した。

### 箇条書きに関する立場と、その根拠の強度

「箇条書きは文脈を壊すので使わない方がよい」という論旨を検証した。結論は**根拠はあるが「使わない方がよい」は行き過ぎ**であり、正確な主張は「箇条書きは項目間の関係を表現できない。したがって論証の代替に使うと論理が失われる」。

支持側。Tufte の [*The Cognitive Style of PowerPoint*](https://www.edwardtufte.com/book/the-cognitive-style-of-powerpoint-pitching-out-corrupts-within-ebook/)（10事例・スライド2,000枚・非PowerPointの対照32件）は、分析が因果的・多変量的・比較的・証拠依存的になるほど箇条書きの害が大きくなるとし、CAIB報告書がNASAのスライド常用を問題視した事例を引く。[Shaw らの3M事例（HBR 1998）](https://hbr.org/1998/05/strategic-stories-how-3m-is-rewriting-business-planning)は、箇条書きの事業計画が「重要な関係を未指定のまま残す」と指摘。[Amazon の6ページメモ](https://www.cnbc.com/2018/04/23/what-jeff-bezos-learned-from-requiring-6-page-memos-at-amazon.html)（2004年の PowerPoint 廃止）も同趣旨の運用判断である。

反対側。[NN/g のアイトラッキング調査](https://www.nngroup.com/articles/f-shaped-pattern-reading-web-content-discovered/)は、読み手が精読せず走査すること（F型パターン）を示し、リストと見出しが走査効率を上げるとする。Google の Technical Writing One は「Lists and tables」を独立ユニットに持ち、長文の分割手段としてリスト化を推奨している。

**根拠の強度に注意**。Tufte も 3M も Amazon も事例分析と専門家判断であり、対照試験ではない。散文と箇条書きを直接比較した実証研究は見当たらなかった。したがって禁止は採らず、役割分担（散文＝なぜ・因果・トレードオフ、箇条書き＝走査・追跡・参照する原子）と「論証を箇条書きだけで済ませない」という限定的な禁止に落とした。比率の数値上限も技術文書には課さない。Issue や PR は本来リスト主体の文書であり、上限は不自然な制約になるため。

### なぜ voice bleed が起きないか

casual と formal は文体が正反対（砕け↔正確さ優先）。progressive disclosure により、**使用時は `common.md` ＋ 片方の style だけがコンテキストに載る**。両スタイルが同時にロードされないため、casual ルールが formal 文書に混入する事故を構造的に防げる。

### なぜ `wordcount.sh` は scripts/ か

スクリプトは「実行するだけで内容はロードしない」層（progressive disclosure の第3段）。文字数測定ロジックを本文に書かず scripts/ に逃がすことで、SKILL.md を薄く保ちトークンを節約する。

### `wordcount.sh` の文字数計数（2026-09-08 改修）

改修前は `tr -d '[:space:]' | wc -m` で数えていた。これには日本語で2つの不具合があった。

1. **ロケール依存** — `wc -m` はロケールが `C` / `POSIX` のときバイト数を返す。UTF-8 の日本語は1文字3バイトのため、「あいうえお」が15字と報告される。CI・cron・最小構成コンテナは既定が `C` になりやすく、環境によって結果が3倍変わる
2. **全角スペースが残る** — `tr` はバイト単位で動くため `[:space:]` は ASCII 空白しか消せない。全角スペース（U+3000）や NBSP（U+00A0）が文字として数えられる

計数そのものは現在も `wc -m` で行い、この2点を前処理で潰している。

- **ロケール** — `ensure_utf8_locale()` で `locale charmap` を確認する。UTF-8 でなければ `C.UTF-8` / `C.utf8` / `en_US.UTF-8` / `ja_JP.UTF-8` の順に `locale -a` から探して `LC_ALL` へ設定し、1つも見つからなければ `die` で停止する。黙って3倍の数値を出すより、止まって理由を告げるほうが害は小さい
- **全角空白** — `sed` のリテラル置換で U+3000 と U+00A0 を先に除去してから `tr -d '[:space:]'` に渡す

数える単位は Unicode 符号点。

#### perl（書記素クラスタ）を採らなかった理由

計数方式の候補を実測で比較した（空白除去後）。

| 入力 | `wc -m` | perl 符号点 | python `len()` | awk `length()` | 継続バイト除去 | NFC＋`\X` |
| --- | --- | --- | --- | --- | --- | --- |
| `あいうえお漢字ABC123、。` | 15 | 15 | 15 | 15 | 15 | 15 |
| か＋結合濁点（見た目1字） | 2 | 2 | 2 | 2 | 2 | **1** |
| 家族絵文字 ZWJ（見た目1字） | 5 | 5 | 5 | 5 | 5 | **1** |

`perl -CS -MUnicode::Normalize` で NFC 正規化してから `\X`（書記素クラスタ）を数える案も検討した。理屈のうえではこれが最も正確である。Unicode の [UAX #29](https://unicode.org/reports/tr29/) は「人が1文字と認識する単位（user-perceived character）」を書記素クラスタと定義しており、原稿用紙の「1マス」と概念が一致する。`-CS` は標準入出力を UTF-8 として扱うため、ロケール確保の処理も不要になる。

採らなかった理由は、**この用途では差が出ないため**。表のとおり差が生じるのは結合文字と絵文字だけで、実文書では `formal.md` 5,789字・PR草稿 1,316字がいずれも符号点数と同値だった。日本語の記事の分量を測るのに perl 依存を足す見返りが無い。

残る差は「UTF-8 ロケールが1つも無い環境」でのふるまい（perl 版は動き、現行版は停止する）だが、そうした最小構成のコンテナには perl も無いことが多く、実質的な優位は薄い。

**空白の扱い**は「小説家になろう」「カクヨム」および一般的な文字数カウントツールと同じく、改行・半角空白・全角空白（U+3000）・NBSP（U+00A0）を除外する方針にそろえた。

**原稿用紙換算**は文字数÷400 の概算であり、実際の組版では改行と字下げが空きマスを消費するため実枚数はこれより多くなる。誤解を避けるため出力にその旨を添えた。

存在しないパスを渡したときに黙って標準入力へフォールバックする挙動も、`die` で停止するよう改めた（タイプミスが無言のハングや誤集計になっていたため）。

## textlint による機械検査（2026-09-08 導入）

チェックリストは論理と構成を見るのに向くが、表記の機械的な誤りと冗長表現は人間も LLM も取りこぼす。この層を textlint に委譲した。

### なぜ技術文書プリセットが casual ではなく formal に適合したか

導入のきっかけは [textlint 設定の記事](https://taiyolab.com/ja/2025/03/03/textlint-settings/)。当初は casual への適用を想定したが、調べると逆だった。理由は3つある。

**目的関数が formal と一致している。** JTF は日本翻訳連盟のスタイルガイド、`preset-ja-technical-writing` も「技術文書向け」を明示する。どちらも最適化対象は一意に読めることであり、読みやすさや面白さではない。これは formal.md の基本方針「読み手のレベルによらず、得られるアウトプットが一定に伝わるか」と同じである。

**casual の価値は、プリセットが削る対象そのものである。** `ja-no-weak-phrase` が消す `かもしれない` は、casual の「率直な判断・感想」にあたる。`no-mix-dearu-desumasu` が統一する揺れは「です・ます混在OK」と「人間味」に、`sentence-length` が抑える長文は「短文と長文を戦略的に配置」に対応する。

プリセットは書き手の痕跡を消す方向に働き、casual は痕跡を残す方向に働く。設計目標が反対である。

**「技術記事」と「技術文書」の語が衝突している。** casual の対象である Zenn の技術記事は技術的な話題を扱う読み物であり、textlint-ja のいう技術文書（仕様書・マニュアル・API ドキュメント）とは別物。参照記事が個人ブログ発であることも錯覚を強めたが、記事自身が JTF プリセットを「小説執筆時は無効化」と書いており、著者も用途で切り替えている。

### なぜこの配布方式か（設定だけ配り、node_modules は同梱しない）

3案を比較し、**設定ファイルのみ同梱し、ルールは利用者プロジェクトに導入する案**を採った。

| 案 | 利用者の手間 | 作者の運用 | 判定 |
| --- | --- | --- | --- |
| npm の shareable config を公開 | 小（2パッケージ） | npm publish の運用が増える | 将来の移行先 |
| 設定のみ同梱・ルールは利用者が導入 | 中（`npm i -D`） | なし | **採用** |
| スキル配下へ自己完結インストール | 小 | symlink とプラグイン更新で壊れる | 却下 |

**`npx --package` 方式は実測で失敗した。** プリセットを `--package` で渡しても `No rules found, textlint hasn't done anything` が返る。npx の一時ディレクトリと textlint のルール解決パスが噛み合わないためである。

したがって「利用者に何もインストールさせない」案は成立しない。

一方、**プロジェクト外にある設定を `--config` で指定しても、ルールは実行時のカレントディレクトリの `node_modules` から解決される**ことを実測で確認した。これが本方式の成立条件である。`textlint.sh` はこの前提に立つため、**検査対象プロジェクトのルートで実行する**必要がある。

利用者が増えて `npm i -D` の手間が問題になったら、同梱の JSON をそのまま config パッケージの中身にして shareable config へ移行できる。

### ルール選定は実測にもとづく

推測でルールを選ばず、リポジトリの Markdown 24ファイルに全ルールを当てて誤検知の実数を数えた。プリセット既定のままでは 195 件、調整後は 43 件（error 18 / warning 25）。casual 設定は 0 件。

| 調整 | 根拠 |
| --- | --- |
| `ja-no-mixed-period: false` | 67 件全てが「例:」「採用理由:」等のリード文と箇条書き。formal.md の「箇条書きの行末に句読点を付けない」と正面衝突する |
| `sentence-length: {max: 150}` | 既定 100 では 63 件。大半は英文行とインラインコード・URL が字数を押し上げたもの。150 で 7 件に減り、残りは実際に長い文 |
| `no-exclamation-question-mark: false` | 4 件全てが「だから何？」等の引用内の疑問符。formal.md 自身が使う表現 |
| `no-mix-dearu-desumasu: {preferInBody: "である"}` | 既定は「ですます」。formal は常体が基本のため反転させる。24 件が 4 件に減る |
| `no-doubled-joshi: {allow: ["も"]}` + warning | 「Tufte も 3M も Amazon も」型の並列助詞が誤検知の主因。`allow` で 27 件が 22 件に。残りは判断が割れるため warning 止まり |

`min_interval` は**上げるほど検出が増える**（1→27件、2→78件、3→146件）。既定の 1 が最も緩い。直感と逆なので変更時は注意する。

**プリセットのルールにオプションを渡すと、プリセットが設定した既定値は上書きで消える。** `max-kanji-continuous-len` に `severity` だけ渡したところ `max: 6` が失われ、検出が 1 件から 14 件に増えた。severity だけ変える場合も既定値を書き直すこと。

### 環境が欠けたときの挙動（2026-09-08 検証）

スキルは node の無い環境にも配布される。実際に PATH を削って挙動を測り、2件の不具合を直した。

| 環境 | 修正前 | 修正後 |
| --- | --- | --- |
| node なし・textlint なし | 案内を出して exit 0（正常） | 変更なし |
| node なし・`node_modules/.bin/textlint` あり | `/usr/bin/env: 'node': No such file or directory`、**rc=127** | node が無い旨を添えて exit 0 |
| `bc` なし（wordcount.sh） | `bc: command not found` を出しつつ **0.0 枚と誤表示** | bc に依存せず正しく換算 |

`textlint.sh` は実体の存在だけでなく `--version` の起動可否で判定する。textlint は node スクリプトであり、ファイルがあっても node が無ければ動かないためである。

**不足しているものに応じて導入手順を出し分ける。** 利用環境には mise がある前提とし、node が無ければ `mise use node@24` を、textlint が無ければ `npm i -D` を提案する。mise 自体が無い場合だけ、その導入先 URL を添える。node が入っていれば npm の行だけを出す。

`wordcount.sh` の換算は bash の整数演算に置き換え、`bc` への依存を外した。`bc` は最小構成のコンテナに無いことがあり、欠けると**黙って誤った数値**を出していた。`wc -m` のロケール問題と同じ失敗の型である。外部コマンドは `sed`・`tr`・`wc`・`grep`・`locale` のみになった。

機械検査は任意工程であり、環境が欠けてもレビュー手順を止めない（exit 0 でスキップ）。一方、分量測定は誤った数値を出すくらいなら止めるべきなので、UTF-8 ロケールが無ければ `die` する。**この非対称は意図的である**。

### 検査の対象範囲

textlint が見るのは表記と語法だけで、論理・構成・根拠は見ない。チェックリストの代替にはならない。SKILL.md とチェックリストの両方にこの但し書きを置いている。

## 更新の指針

- **共通ルールの変更** → `styles/common.md` のみ。casual/formal 両方に効く。
- **片方の文体だけ変更** → 該当する `styles/casual.md` か `styles/formal.md`。
- **新しい文書種別の追加**（例: メール、プレゼン） → `styles/<種別>.md` を追加し、SKILL.md の「文書種別の判定」表に1行追加。common は触らない。ただし既存スタイルと守備範囲が重なる追加は避ける（ルーティングが曖昧になる）。
- **description の変更** → 自動ディスパッチの精度に直結する。種別キーワード（レポート/ブログ/Zenn ⇔ Issue/PR/Design Doc/仕様書/設計書/議事録）を維持し、「会話的短文・コード生成は除外」の但し書きを残す。
- **チェックリストの追加** → 共通なら common、種別固有なら該当 style の「固有チェックリスト」へ。
- **共通に置くか迷ったら style 側へ置く**。誤った共通化は、そのスタイルの利用者が気づかないまま品質を下げる。
- 変更後は `npx markdownlint-cli2 "plugins/writing-style/**/*.md"` でlintを通す（リポジトリ規約）。

## なぜスキル外（docs/）に置くか

このガイドは「スキルを実行するときには不要、更新するときだけ要る」情報。スキル実行時のコンテキストに載せたくない。

- スキル起動時にロードされるのは `name`+`description`（常時）、`SKILL.md` 本体（起動時）、SKILL.md から参照され Claude が読むと判断したファイルのみ。**未参照ファイルはゼロトークン**（[best-practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)）。
- ただし、同梱ファイルはプロジェクト探索中に誤って読まれトークンを浪費する事例が報告されている（[anthropics/skills #763](https://github.com/anthropics/skills/issues/763)）。「非参照」は*デフォルトで読まれない*を保証するが*絶対に読まれない*ではない。
- `.skillignore` 的な公式の除外機構は存在しない（[agentskills.io spec](https://agentskills.io/specification)）。SKILL.md 本文への人間専用コメント機能も未実装（[#177](https://github.com/anthropics/skills/issues/177)）。
- → 最も確実な隔離は**物理的にスキル外へ置く**こと。`mise run install` の symlink はスキルディレクトリ単位なので、`docs/` 配下は `~/.claude/skills/` に露出せず、誤読リスクがゼロになる。リポジトリ既存の `docs/template-management.md` と同じ運用。

## 参考文献

本スキルの設計で参照した公開資料

### textlint が依拠するもの

- [textlint](https://textlint.org/) — 検査エンジン本体
- [textlint-rule-preset-ja-technical-writing](https://github.com/textlint-ja/textlint-rule-preset-ja-technical-writing) — `formal.json` / `casual.json` が使うプリセット。各ルールの意図と既定値は README に記載。既定は「少し厳しめ」と明言されており、文書に合わせた調整を前提とする
- [textlint-rule-no-kangxi-radicals](https://github.com/xl1/textlint-rule-no-kangxi-radicals) — 康熙部首（見た目が漢字に酷似する別符号）の検出。プリセット外で追加
- [JTF 日本語標準スタイルガイド](https://www.jtf.jp/tips/styleguide) — 日本翻訳連盟による表記規約。[textlint-rule-preset-JTF-style](https://github.com/textlint-ja/textlint-rule-preset-JTF-style) が実装。導入検討時に比較したが本スキルでは不採用
- [textlint 設定の記事（taiyolab, 2025）](https://taiyolab.com/ja/2025/03/03/textlint-settings/) — 導入のきっかけ。プリセットを用途で切り替える運用の出典

### 技術文書の作法

- [Google Technical Writing One](https://developers.google.com/tech-writing/one) — 1文1アイデア、不要語の削除、リストの並列構造。段落の長さは [Paragraphs](https://developers.google.com/tech-writing/one/paragraphs) を参照
- [Google developer documentation style guide](https://developers.google.com/style) — 番号付きリストの命令形動詞始まりなどの表記規約。1段落1アイデアの規定は [Paragraph structure](https://developers.google.com/style/paragraph-structure)
- [高品質な技術文書を書くには（riywo, 2021）](https://blog.riywo.com/2021/01/how-to-write-high-quality-technical-doc/) — 執筆前の読み手・アウトプット・課題の定義、one-way / two-way door、Appendix 送りの判断

### 箇条書きの是非

- [Edward Tufte, *The Cognitive Style of PowerPoint*](https://www.edwardtufte.com/book/the-cognitive-style-of-powerpoint-pitching-out-corrupts-within-ebook/) — 箇条書きが因果・比較・証拠依存の分析を損なうとする事例分析
- [Shaw et al., "Strategic Stories: How 3M Is Rewriting Business Planning"（HBR, 1998）](https://hbr.org/1998/05/strategic-stories-how-3m-is-rewriting-business-planning) — 箇条書きの事業計画が項目間の関係を未指定に残すという指摘
- [Amazon の6ページメモ（CNBC, 2018）](https://www.cnbc.com/2018/04/23/what-jeff-bezos-learned-from-requiring-6-page-memos-at-amazon.html) — PowerPoint 廃止の運用判断
- [F-Shaped Pattern For Reading Web Content（NN/g）](https://www.nngroup.com/articles/f-shaped-pattern-reading-web-content-discovered/) — 走査読みの実測。リストと見出しが走査効率を上げるとする反対側の根拠

### 文字数の計数

- [UAX #29: Unicode Text Segmentation](https://unicode.org/reports/tr29/) — 書記素クラスタの定義。`wordcount.sh` の計数単位を検討した際の根拠

### Agent Skills の仕様

- [Agent Skills best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices) — progressive disclosure とロード対象の規定
- [agentskills.io specification](https://agentskills.io/specification) — 除外機構が存在しないことの確認
- [anthropics/skills #763](https://github.com/anthropics/skills/issues/763) — 同梱ファイルの誤読によるトークン浪費の報告
- [anthropics/skills #177](https://github.com/anthropics/skills/issues/177) — SKILL.md への人間専用コメント機能の要望
