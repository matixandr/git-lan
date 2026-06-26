#!/usr/bin/env bash
#
# git-lan installer (Unix). Builds from source and installs the binary so that
# `git lan` works (git resolves `git lan` to a `git-lan` executable on PATH).
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="git-lan"

# Pick an install directory that is writable and on PATH.
if [ -n "${PREFIX:-}" ]; then
  INSTALL_DIR="$PREFIX/bin"
elif [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
fi

echo "==> git-lan installer"

if ! command -v go >/dev/null 2>&1; then
  echo "error: Go toolchain not found. Install Go 1.26+ from https://go.dev/dl/" >&2
  exit 1
fi

if ! command -v git >/dev/null 2>&1; then
  echo "error: git not found. git-lan shells out to the git binary." >&2
  exit 1
fi

VERSION="$(git -C "$REPO_ROOT" describe --tags --always --dirty 2>/dev/null || echo dev)"
LDFLAGS="-s -w -X github.com/matixandr/git-lan/cmd.Version=${VERSION}"

echo "==> building ${BINARY} (${VERSION})"
( cd "$REPO_ROOT" && go build -ldflags "$LDFLAGS" -o "$BINARY" . )

echo "==> installing to ${INSTALL_DIR}"
mkdir -p "$INSTALL_DIR"
install -m 0755 "$REPO_ROOT/$BINARY" "$INSTALL_DIR/$BINARY"
rm -f "$REPO_ROOT/$BINARY"

echo "==> done. ${BINARY} -> ${INSTALL_DIR}/${BINARY}"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "note: add ${INSTALL_DIR} to your PATH to use 'git lan'." ;;
esac

echo "Try: git lan list"
