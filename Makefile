# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory

SKILLS_DIR := $(HOME)/.claude/skills
REPO_DIR := $(shell pwd)
SKILLS := bash-script-template go-project-scaffold mysql-container-setup pr-workflow

.PHONY: help
help: ## ヘルプ表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: install
install: ## 全スキルをインストール
	@mkdir -p $(SKILLS_DIR)
	@for skill in $(SKILLS); do \
		if [ -L "$(SKILLS_DIR)/$$skill" ]; then \
			echo "既存: $$skill"; \
		else \
			ln -s "$(REPO_DIR)/$$skill" "$(SKILLS_DIR)/$$skill"; \
			echo "インストール: $$skill"; \
		fi \
	done

.PHONY: uninstall
uninstall: ## 全スキルをアンインストール
	@for skill in $(SKILLS); do \
		if [ -L "$(SKILLS_DIR)/$$skill" ]; then \
			rm "$(SKILLS_DIR)/$$skill"; \
			echo "削除: $$skill"; \
		fi \
	done

.PHONY: list
list: ## インストール状態を表示
	@for skill in $(SKILLS); do \
		if [ -L "$(SKILLS_DIR)/$$skill" ]; then \
			echo "[インストール済] $$skill"; \
		else \
			echo "[未インストール] $$skill"; \
		fi \
	done
