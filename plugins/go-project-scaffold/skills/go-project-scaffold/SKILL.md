---
name: go-project-scaffold
description: 新規Goプロジェクトの初期化。go mod init、Makefile、README.md、.gitignore、.golangci.yml（all linters有効）を作成。ghqディレクトリでのプロジェクト開始時に使用。
---

# go-project-scaffold

新規Goプロジェクトの初期化スキル。

## 実行内容

1. `go mod init` でgo.modを作成
2. Makefileを作成
3. README.mdを作成
4. .gitignoreを作成
5. .golangci.ymlを作成

## 手順

### 1. モジュール名の決定

優先順位:
1. 引数でモジュール名が指定されている場合はそちらを使用
2. ghqディレクトリ構造の場合、パスから自動取得
3. 上記以外はユーザーに確認

ghqディレクトリの判定と取得:
```bash
# パスにghq/が含まれるか確認
if pwd | grep -q '/ghq/'; then
  # モジュール名の取得
  pwd | sed 's|.*/ghq/||'
fi
```

例:
- `/home/takekazu/ghq/github.com/takekazu/project` → `github.com/takekazu/project`
- `/home/takekazu/ghq/github.com/takekazu/repo/subdir` → `github.com/takekazu/repo/subdir`

ghqディレクトリでない場合はAskUserQuestionでモジュール名を確認。

### 2. go mod init

```bash
go mod init <モジュール名>
```

### 3. Makefile作成

以下の内容でMakefileを作成：

```makefile
# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory
BIN := .tmp/bin
export PATH := $(abspath $(BIN)):$(PATH)
export GOBIN := $(abspath $(BIN))

.PHONY: help
help: ## ヘルプ表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: lint
lint: ## golangci-lint実行
	golangci-lint run -c .golangci.yml --timeout 10m

.PHONY: deps
deps: ## 依存ツールのインストール
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
	go install golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: build
build: ## ビルド
	go build ./...

.PHONY: test
test: ## テスト実行
	go test -v ./...

.PHONY: fmt
fmt: ## フォーマット
	go fmt ./...

.PHONY: clean
clean: ## クリーンアップ
	go clean
	rm -rf .tmp/
```

### 4. README.md作成

以下の内容でREADME.mdを作成（プロジェクト名は実際のものに置換）：

```markdown
# プロジェクト名

## 概要

[プロジェクトの説明]

## 開発

```bash
# 依存ツールのインストール
make deps

# ビルド
make build

# テスト
make test

# lint
make lint
```

## ライセンス

MIT
```

### 5. .gitignore作成

```gitignore
# Build output
*.exe
*.test
*.out

# Temporary
.tmp/

# IDE
.idea/
.vscode/

# OS
.DS_Store
```

### 6. .golangci.yml作成

```yaml
version: "2"

linters:
  default: all

  disable:
    - depguard          # stdlibだけに限定できないので無効化
    - testpackage       # テストを同一パッケージに入れることにするので無効化
    - wrapcheck         # stacktrace使用のため無効化
    - wsl               # うるさ過ぎるので無効化
    - gosmopolitan      # 日本語リテラルを許可したいので無効化
    - gochecknoinits    # cliではinit関数は便利なので無効化
    - varnamelen        # スコープが狭い変数は短い名前にすべきなので無効化
    - godot             # 日本語のコメントはピリオドを使用しないので無効化
    - gochecknoglobals  # global変数を使用してもいいので無効化
    - revive            # 現状のスタイルは必要なものだけ書く文化なので無効化
    - ireturn           # ジェネリクスを使っているので無効化

  settings:
    errcheck:
      check-type-assertions: true
    godox:
      keywords: [FIXME]
    lll:
      line-length: 150
      tab-width: 4
    funlen:
      lines: 80
      statements: 40
      ignore-comments: true
    cyclop:
      max-complexity: 15

  exclusions:
    warn-unused: true
    rules:
      - linters: [funlen]
        path: ".*_test\\.go$"

formatters:
  settings:
    gci:
      sections:
        - standard
        - default

output:
  formats:
    text:
      print-issued-lines: true
      print-linter-name: true
```

## 注意事項

- 既存ファイルがある場合は上書きしない
- ユーザーに確認してから作成
