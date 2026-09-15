#!/usr/bin/env bash
#
# ensure-binary.sh は workspace-mcp バイナリを ${CLAUDE_PLUGIN_DATA}/bin に用意する。
# SessionStart フックから呼ばれ、MCP サーバ起動前に取得を済ませる（起動タイムアウト回避）。
#
# 取得経路（設計 §5 のビルドと配布）:
#   主経路       GitHub Releases のクロスコンパイル済みバイナリ（checksum 検証）※リリース整備後に有効化
#   フォールバック go install（モジュールプロキシ。go.sum で検証）
#
# 段階1ではフォールバックの go install を用いる。リリース CI（段階2）整備後に主経路を有効化する。
set -euo pipefail

readonly MODULE="github.com/takekazuomi/takekazuomi-claude-plugins/mcp/workspace"
readonly VERSION="${WORKSPACE_MCP_VERSION:-latest}"
readonly DEST_DIR="${CLAUDE_PLUGIN_DATA:?CLAUDE_PLUGIN_DATA is required}/bin"
readonly DEST="${DEST_DIR}/workspace-mcp"

die() {
  echo "ensure-binary: $*" >&2
  exit 1
}

install_via_go() {
  command -v go >/dev/null 2>&1 || return 1
  local tmp
  tmp="$(mktemp -d)"
  # go install は最後のパス要素 'workspace' を GOBIN に置く。それを workspace-mcp へ移す。
  if GOBIN="$tmp" go install "${MODULE}@${VERSION}"; then
    mv "${tmp}/workspace" "$DEST"
    chmod +x "$DEST"
    rm -rf "$tmp"
    return 0
  fi
  rm -rf "$tmp"
  return 1
}

main() {
  if [[ -x "$DEST" ]]; then
    return 0 # 取得済み（${CLAUDE_PLUGIN_DATA} は更新をまたいで永続）
  fi
  mkdir -p "$DEST_DIR"

  # TODO(段階2): GitHub Releases からの取得＋checksum/署名検証を主経路として実装する。

  install_via_go && return 0
  die "failed to install workspace-mcp: go toolchain not found and GitHub Releases fetch not yet configured"
}

main "$@"
