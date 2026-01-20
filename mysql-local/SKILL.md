---
name: mysql-local
description: Goプロジェクトにローカルテスト用MySQLコンテナを追加。docker-compose.yml、初期化SQL、Makefileターゲットを作成。
---

# mysql-local

Goプロジェクトにローカル開発用MySQLコンテナを追加するスキル。

## 実行内容

1. docker-compose.ymlを作成（MySQL 8.0）
2. mysql/initdb.d/ディレクトリと初期化SQLを作成
3. .envrc.sampleを作成
4. MakefileにDBターゲットを追加
5. .gitignoreにmysql/data/を追加

## 手順

### 1. プロジェクト名・DB名の決定

優先順位:
1. go.modが存在する場合、モジュール名の最後の部分を使用
2. go.modがない場合、カレントディレクトリ名を使用

```bash
# go.modからプロジェクト名取得
if [ -f go.mod ]; then
  basename $(head -1 go.mod | awk '{print $2}')
else
  basename $(pwd)
fi
```

プロジェクト名はコンテナ名とデフォルトDB名に使用。

### 2. docker-compose.yml作成

以下の内容でdocker-compose.ymlを作成（PROJECT_NAMEは実際のプロジェクト名に置換）：

```yaml
services:
  mysql:
    image: mysql:8.0
    container_name: PROJECT_NAME-mysql
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-rootpassword}
      MYSQL_DATABASE: ${MYSQL_DATABASE:-appdb}
      MYSQL_USER: ${MYSQL_USER:-appuser}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD:-apppassword}
      TZ: Asia/Tokyo
    ports:
      - "${MYSQL_PORT:-3306}:3306"
    volumes:
      - ./mysql/data:/var/lib/mysql
      - ./mysql/initdb.d:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-time-zone=Asia/Tokyo
```

### 3. mysql/initdb.d/ディレクトリ作成

```bash
mkdir -p mysql/initdb.d
```

初期化SQLファイルを作成：

**mysql/initdb.d/01_init.sql**
```sql
-- 初期化SQL
-- ここにテーブル定義やシードデータを追加

-- 例: usersテーブル
-- CREATE TABLE IF NOT EXISTS users (
--     id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
--     email VARCHAR(255) NOT NULL UNIQUE,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4. .envrc.sample作成

```bash
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=appuser
export MYSQL_PASSWORD=apppassword
export MYSQL_DATABASE=appdb
export DSN="${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DATABASE}?parseTime=true&loc=Asia%2FTokyo"
```

### 5. Makefileターゲット追加

既存のMakefileに以下のターゲットを追加：

```makefile
.PHONY: db-up db-down db-logs db-client db-clean

db-up: ## MySQLコンテナ起動
	docker compose up -d mysql

db-down: ## MySQLコンテナ停止
	docker compose down

db-logs: ## MySQLログ表示
	docker compose logs -f mysql

db-client: ## MySQL接続
	docker compose exec mysql mysql -u appuser -papppassword appdb

db-clean: db-down ## データ含めて削除
	rm -rf mysql/data/*
```

### 6. .gitignore更新

.gitignoreに以下を追加：

```gitignore
# MySQL data
mysql/data/

# Environment
.envrc
```

## 注意事項

- 既存ファイルがある場合は上書きしない（Makefileはターゲット追加のみ）
- docker-compose.ymlが既存の場合はmysqlサービスを追加
- ポート3306が使用中の場合は.envrcでMYSQL_PORTを変更

## セットアップ後の手順

.envrc.sampleの内容を.envrcにコピーして環境変数を有効化：

```bash
cat .envrc.sample >> .envrc
direnv allow
```
