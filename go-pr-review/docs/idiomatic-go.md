# Idiomatic Go（Go 風コード）ガイド

Go 風（Idiomatic Go）とは、Go の設計思想に沿った、自然で読みやすく保守しやすいコードを書くことである。このガイドは公式ドキュメント・Go 設計者の発言・コミュニティのベストプラクティスを原典として構成している。

> **TL;DR** — Clarity > Simplicity > Concision > Maintainability > Consistency の優先度で設計する。エラーは値として扱い、インターフェースは小さく消費側で定義し、ゼロ値を活用する。可変なグローバル状態を避け、goroutine のライフサイクルを管理し、gofmt を無条件に適用する。

---

## 1. 設計原則（Style Principles）

Google Go Style Guide [^1] が定義する 5 つの原則を、優先度順に示す。

1. **Clarity（明快さ）** — コードの目的と根拠が読み手に明確であること
2. **Simplicity（単純さ）** — 目的を最も単純な方法で達成すること
3. **Concision（簡潔さ）** — シグナル対ノイズ比が高いこと
4. **Maintainability（保守性）** — 将来の変更に耐えること
5. **Consistency（一貫性）** — 既存コードベースとの統一

Rob Pike は「Clear is better than clever」と述べており [^2]、「Go at Google」[^10] ではソフトウェア工学の観点から Go の設計判断を解説している。特にこの講演で Pike は「The key point here is our programmers are Googlers, they're not researchers」と述べ、Go が研究者ではなく実務のソフトウェアエンジニアのために設計された言語であることを明確にした。Clarity が最優先される根拠はここにある。

Dave Cheney は The Zen of Go [^3] で「最終的なゴールは maintainability だ」と結論づけている。

## 2. Go Proverbs — 19 の格言

Rob Pike が GopherFest SV 2015 で提唱した格言 [^2] から、コーディングに直結するものを以下に分類する。なお、この分類は本ガイド独自のものであり、原典の講演では分類なしに一連の格言として提示されている。

### 並行処理

- **Don't communicate by sharing memory, share memory by communicating.**
- **Concurrency is not parallelism.**
- **Channels orchestrate; mutexes serialize.**

### 設計・抽象化

- **The bigger the interface, the weaker the abstraction.**
- **Make the zero value useful.**
- **interface{} says nothing.**
- **A little copying is better than a little dependency.**

### エラーと安全性

- **Errors are values.**
- **Don't just check errors, handle them gracefully.**
- **Don't panic.** [^4]

### 明快さ

- **Clear is better than clever.**
- **Reflection is never clear.**
- **Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.**

### ドキュメントとアーキテクチャ

- **Design the architecture, name the components, document the details.**
- **Documentation is for users.**

### 境界

- **Cgo is not Go.**
- **With the unsafe package there are no guarantees.**
- **Syscall must always be guarded with build tags.**
- **Cgo must always be guarded with build tags.**

> **注記:** 18 個の格言は Go Proverbs サイト [^2] から Rob Pike の講演動画の各タイムスタンプにリンクされている。ただし「Don't panic.」だけは例外で、Go Wiki の CodeReviewComments [^4] にリンクされている。

## 3. 実践パターン

以下は、公式 CodeReviewComments [^4]、Effective Go [^5]、Google Go Style Guide [^1]、The Zen of Go [^3]、および Damian Gryski がまとめた Idiomatic Go Resources [^7] から抽出した実践パターンである。

### 3.1 命名

パッケージ名は短い小文字の名詞 1 語が理想 [^9]。利用側で `pkg.Name` と読むため、パッケージ名との重複（stuttering）を避ける。

```go
// ✗ stuttering
http.HTTPServer

// ✓
http.Server
```

ローカル変数は短く、スコープが広がるほど説明的に。レシーバ名は 1〜2 文字。

```go
func (b *Buffer) Read(p []byte) (n int, err error)
```

### 3.2 ゼロ値を活用する

型のゼロ値が即座に使える設計にすると、コンストラクタなしで動作する。これは `sync.Mutex`、`bytes.Buffer`、`sync.WaitGroup` など標準ライブラリ全体で徹底されている。

```go
// ゼロ値で即使える
var mu sync.Mutex
var buf bytes.Buffer
buf.WriteString("hello")
```

goaux [^6] のパッケージ群もこのパターンを一貫している。たとえば `waitgroup.Sync` はゼロ値で即座に使える。

```go
// goaux/waitgroup — ゼロ値で即使用可
var sy waitgroup.Sync
sy.Go(func() { /* 並行タスク */ })
sy.Wait()
```

### 3.3 エラー処理

エラーは値であり、戻り値として明示的に扱う。エラーを黙って無視してはならない。

```go
// ✗ 黙殺
fi, _ := os.Stat(path)

// ✓ 明示的に処理（goaux/stacktrace を使用）
fi, err := os.Stat(path)
if err != nil {
    return stacktrace.Errorf("stat %s: %w", path, err)
}
```

エラーのラップには `github.com/goaux/stacktrace/v2` [^11] を使う。`stacktrace.Errorf` は `fmt.Errorf` の直接的な置き換えであり、スタックトレース情報を自動的に付与する。`%w` によるラップの意味論は `fmt.Errorf` と同一なので、`errors.Is` / `errors.As` による判別も従来通り機能する。既存のエラーをそのままラップする場合は `stacktrace.With` を使う。

```go
import "github.com/goaux/stacktrace/v2"

// 新規エラー生成（errors.New の置き換え）
return stacktrace.New("invalid argument")

// コンテキスト付きラップ（fmt.Errorf の置き換え）
return stacktrace.Errorf("open config %s: %w", path, err)

// 既存エラーにスタックトレースだけ付与
return stacktrace.With(err)

// エラー表示時にスタックトレースを含める
fmt.Println(stacktrace.Format(err))
```

スタックトレースが付くことで、本番環境でのデバッグが格段に楽になる。

### 3.4 インターフェースは小さく、消費側で定義する

Go のインターフェースは暗黙的に満たされる（structural typing）。したがって、インターフェースは利用する側（consumer）が必要最小限で定義するのが Go 流だ。

```go
// 標準ライブラリの好例
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

1〜2 メソッドのインターフェースが最も強力な抽象化を提供する。

### 3.5 ガード節で早期リターン（Line of Sight）

happy path を左端に保ち、ネストを浅くする。Mat Ryer はこれを「line of sight coding」と呼んでいる。

```go
// ✗ 深いネスト
func process(r *http.Request) error {
    if r != nil {
        if r.Header != nil {
            // 本来の処理...
        }
    }
    return nil
}

// ✓ ガード節で早期リターン
func process(r *http.Request) error {
    if r == nil {
        return errors.New("nil request")
    }
    if r.Header == nil {
        return errors.New("nil header")
    }
    // 本来の処理...
    return nil
}
```

### 3.6 パッケージレベルの可変状態を避ける

避けるべきはミュータブルな（変更される）パッケージレベル変数である。Uber Go Style Guide [^8] が「Avoid Mutable Globals」セクションで詳述しているように、グローバルな可変状態はテストしにくく、並行実行と相性が悪い。依存は型のフィールドとして注入する。

```go
// ✗ パッケージレベルの可変状態
var db *sql.DB

func GetUser(id int) (*User, error) {
    return queryUser(db, id)
}

// ✓ 構造体に注入
type UserStore struct {
    DB *sql.DB
}

func (s *UserStore) GetUser(id int) (*User, error) {
    return queryUser(s.DB, id)
}
```

一方、イミュータブルな（初期化後に変更されない）パッケージ変数は問題ない。正規表現のコンパイルのように不変でコストが高いものは、むしろパッケージ変数として保持すべきだ。関数内で毎回コンパイルすると 12〜38 倍のパフォーマンス劣化を招く[^12]。

```go
// ✓ イミュータブルなパッケージ変数 — 推奨
var reEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ✗ 関数内で毎回コンパイル — 避ける
func validate(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}
```

標準ライブラリも同じ方針に従っている。センチネルエラー（`io.EOF`、`sql.ErrNoRows`）やコンパイル済みの正規表現は、パッケージ変数として公開されている。ただし、**公開されているパッケージ変数を外部から変更してはならない**。Go には言語レベルの `const` 保護がないが、これらは慣習的にイミュータブルとして扱う契約だ。

```go
// 標準ライブラリの例 — これらは変更してはならない
var EOF = errors.New("EOF")           // io パッケージ
var ErrNoRows = errors.New("sql: no rows in result set") // database/sql
```

### 3.7 goroutine のライフサイクルを管理する

goroutine を起動する前に「いつ停止するか」「停止をどう知るか」を決める。並行処理の起動は呼び出し側に任せる（Leave concurrency to the caller）。

```go
// ✓ context によるキャンセルと errgroup による待ち合わせ
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error {
    return fetchData(ctx)
})
if err := g.Wait(); err != nil {
    return err
}
```

### 3.8 gofmt は絶対

フォーマットの議論はしない。`gofmt`（または `goimports`）を無条件に適用する。機械的なスタイルの問題はツールに任せ、人間は設計とロジックに集中する。

### 3.9 少しのコピーは少しの依存に勝る

小さなヘルパーを外部パッケージに依存するよりも、コピーして持つ方がよい場合がある。依存を増やすことのコスト（ビルド時間、バージョン管理、推移的依存）を常に意識する。

ただし、この格言は「コピーか巨大な依存か」の二択ではない。goaux [^6] のような「単一責務・小さなパッケージ」の設計思想は第三の道を示している。goaux organization は 23 のリポジトリを公開しており、各パッケージがひとつの責務だけを担う。たとえば [timer](https://pkg.go.dev/github.com/goaux/timer) は context-aware な Sleep だけ、[trim](https://pkg.go.dev/github.com/goaux/trim) はインデント処理だけ、[headline](https://pkg.go.dev/github.com/goaux/headline) は最初の非空行の抽出だけを提供する。小さなパッケージへの依存は、大きなユーティリティライブラリへの依存とはコストが根本的に異なる。

さらに goaux は Go の最新機能も積極的に活用している。[goaux/iter/transform](https://pkg.go.dev/github.com/goaux/iter/transform) は Go 1.23 で導入された `iter.Seq` / `iter.Seq2` を活用した Map、Filter、Zip、Concat 等の変換関数を提供しており、[goaux/scope](https://pkg.go.dev/github.com/goaux/scope) は iterator を使ったリソースの自動クローズを実現している。Go の最新イディオムを実践的に体現したパッケージ群と言える。

### 3.10 `internal` パッケージは基本的に使わない

Go の `internal` ディレクトリ規約は、Russ Cox が Go 1.4（2014年）でツールチェーン強制の可視性メカニズムとして導入した [^13]。公式モジュールレイアウトガイド [^9] は「パッケージはできるだけ `internal` に置くことを推奨」と述べている。しかし、この推奨は **あらゆるプロジェクトにデフォルトで適用すべきもの** ではない。

Ian Lance Taylor が golang-nuts で行った定義的な発言が、`internal` の目的を明確に限定している。**「目的は何かを防ぐことでもなく、いかなるセキュリティを提供することでもない。パッケージを複数のパッケージに構造化する際に、エクスポートされ、サポートされ、ドキュメント化されたAPIを増やさずに済むようにすることだ」** [^14]。つまり `internal` はアクセス制御ではなく API サーフェス管理のためのメカニズムだ。

この設計意図を踏まえると、`internal` が有効なのは以下のケースに限定される。

- **公開ライブラリ** — 外部の利用者がいるため、公開 API サーフェスの管理が不可欠だ。実装の詳細を `internal` に隔離し、セマンティックバージョニングの制約下で互換性を守る判断は合理的である
- **標準ライブラリ** — `internal` が設計された本来の動機がまさにこれだ

一方、**非公開のアプリケーションやサービスでは `internal` を基本的に使わない**。理由は3つある。

第一に、Go modules がすでにモジュール境界によるインポートスコープを提供しており、`GOPATH` 時代に `internal` が補完していた機能は modules で代替される。第二に、チーム自身が著者であり利用者でもある非公開プロジェクトでは、`internal` が守るべき「外部消費者」がそもそも存在しない。第三に、`internal` はパッケージ間の再利用を妨げ、不要なコピペを誘発し、リファクタリング時のディレクトリ移動コストを増やす。

Russ Cox 自身がこの方向を裏付けている。`golang-standards/project-layout` に対して Issue #117 を立て、`cmd/internal/pkg` の三位一体を「標準」として普及させた同リポジトリを批判し、最小限の標準を提案した [^15]。**「ルートディレクトリに LICENSE ファイルを置く。ルートディレクトリに go.mod ファイルを置く。リポジトリに Go コードを置く。以上だ」。** この Issue は 1,400 件以上の賛成リアクションを受けた。

```text
// アプリケーションにおけるカプセル化の判断フロー
//
// 1. まず Go 組み込みの可視性（大文字/小文字）で制御する
// 2. 再利用可能なコードは独立モジュールに切り出す（goaux 方式）
// 3. 公開ライブラリで API サーフェスを制限したいときだけ internal を使う
```

なお、Pike の格言「A little copying is better than a little dependency」は「外部依存を減らすためにコピーを許容せよ」という文脈で語られたものだ。しかし `internal` が強制するコピーは「同一リポジトリ内での不必要な重複」であり、格言の意図とは異なる点に注意したい。

`internal` パッケージの起源、論争、そして代替アプローチの詳細については [Go の `internal` パッケージ：その起源、論争、そしてカーゴカルト的採用への疑問](go-internal-package-debate.md) を参照してほしい。

### 3.11 テストで API の振る舞いを固定する

テストはパッケージの契約書である。公開 API を追加・変更・削除するときは、必ずテストも更新する。テーブル駆動テストが Go の標準パターンだ。

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"zero", 0, 0, 0},
        {"negative", -1, 1, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### 3.12 パフォーマンスは推測でなくベンチマークで

「defer は遅い」「atomic を常に使え」といった教条に従う前に、`go test -bench` で証明する。Go は優れたプロファイリングツール（pprof、trace）を標準で備えている。

### 3.13 節度ある機能の使用

channel、goroutine、embedding、generics——どれも強力だが、使いすぎると複雑性が増す。シンプルな方法で十分なら、そちらを選ぶ。

## 4. コードレビューの観点

CodeReviewComments [^4] から頻出の指摘事項をまとめる。

| 観点 | 要点 |
| --- | --- |
| gofmt | 必ず適用。議論の余地なし |
| Doc Comments | 公開名は必ず名詞で始まるコメント |
| nil スライス | `var s []string` を優先（`[]string{}` は JSON 用途のみ） |
| エラー文字列 | 小文字始まり、句点なし（他メッセージと結合されるため） |
| レシーバ型 | 迷ったらポインタ。小さい不変値なら値レシーバ |
| Indent Error Flow | 正常系を左端に、エラー処理を早期リターン |
| Import | 標準ライブラリ / 外部 / 内部 の 3 グループに空行で分離 |
| Initialisms | `URL`、`HTTP`、`ID` など頭字語は全大文字 |

## 5. 命名規則クイックリファレンス

| 項目 | Go 風 | 非 Go 風 |
| ------ | ------- | --------- |
| 変数名 | 短く文脈依存（i, r, err） | 冗長（index, reader, error） |
| パッケージ名 | 小文字単一語（parser, ast） | アンダースコアや複合語 |
| エクスポート | PascalCase（ParseString） | その他 |
| 非エクスポート | camelCase（parseHeader） | その他 |
| 頭字語 | 全大文字（ID, URL, HTTP） | 混在（Id, Url） |
| レシーバ名 | 1〜2文字（p, b, s） | self, this |

## 6. 参照

[^1]: Google Go Style Guide — <https://google.github.io/styleguide/go/guide.html>
[^2]: Rob Pike, Go Proverbs (GopherFest SV 2015) — <https://go-proverbs.github.io/> / [動画](https://www.youtube.com/watch?v=PAAkCSZUG1c)
[^3]: Dave Cheney, The Zen of Go (GopherCon Israel 2020) — <https://dave.cheney.net/2020/02/23/the-zen-of-go>
[^4]: Go Wiki: CodeReviewComments — <https://go.dev/wiki/CodeReviewComments>
[^5]: Effective Go — <https://go.dev/doc/effective_go>
[^6]: goaux organization — <https://github.com/orgs/goaux>
[^7]: Damian Gryski, Idiomatic Go Resources — <https://medium.com/@dgryski/idiomatic-go-resources-966535376dba>
[^8]: Uber Go Style Guide — <https://github.com/uber-go/guide/blob/master/style.md>
[^9]: Organizing a Go module — <https://go.dev/doc/modules/layout>
[^10]: Rob Pike, "Go at Google: Language Design in the Service of Software Engineering" — <https://go.dev/talks/2012/splash.article>
[^11]: goaux/stacktrace v2 — <https://pkg.go.dev/github.com/goaux/stacktrace/v2>
[^12]: regexp.Compile の性能影響についての実測データ: Sudhanshu Patel, "The Performance Impact of regexp.Compile in Go" (約12.7倍) — <https://medium.com/@sudhanshuptl13/the-performance-impact-of-regexp-compile-in-go-function-level-vs-module-level-7032ecd4cc52> / Thanh Tung Nguyen, "My mistakes in Golang" (約38倍) — <https://tunghatbh.medium.com/programming-my-mistakes-in-golang-69bda628eb65>
[^13]: Go 1.4 リリースノート — <https://go.dev/doc/go1.4>
[^14]: Ian Lance Taylor, golang-nuts での internal パッケージ議論 — <https://groups.google.com/g/golang-nuts/c/A3KA8SNjcp0>
[^15]: Russ Cox, golang-standards/project-layout Issue #117 — <https://github.com/golang-standards/project-layout/issues/117>
