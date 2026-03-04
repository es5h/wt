#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

FORCE=0
if [ "${1:-}" = "--force" ]; then
  FORCE=1
  shift
fi

VERSION_FILE="./VERSION"
if [ -f "$VERSION_FILE" ]; then
  VERSION="$(tr -d ' \t\r\n' <"$VERSION_FILE")"
else
  VERSION="dev"
fi

if command -v go >/dev/null 2>&1; then
  :
else
  echo "wt: go not found in PATH" >&2
  exit 1
fi

echo "Installing wt v$VERSION (development) ..."
echo "Note: shell completion/TUI integrations are NOT installed automatically."

INSTALL_DIR=""
if command -v wt >/dev/null 2>&1; then
  WT_PATH="$(command -v wt)"
  INSTALL_DIR="$(dirname "$WT_PATH")"
fi

if [ -z "$INSTALL_DIR" ]; then
  if command -v go >/dev/null 2>&1; then
    GOPATH="$(go env GOPATH 2>/dev/null || true)"
  else
    GOPATH=""
  fi

  # Prefer ~/.local/bin when it's likely on PATH in typical dev setups.
  if [ -d "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
  elif [ -n "$GOPATH" ]; then
    INSTALL_DIR="$GOPATH/bin"
  else
    echo "wt: failed to determine install directory" >&2
    exit 1
  fi
fi

echo "Install dir: $INSTALL_DIR"
TARGET="$INSTALL_DIR/wt"

if command -v wt >/dev/null 2>&1; then
  if wt --version >/dev/null 2>&1; then
    echo "Current: $(wt --version)"
  else
    echo "Current: wt (unknown version)"
  fi
fi

if [ -f Makefile ]; then
  make build >/dev/null
fi

mkdir -p "$INSTALL_DIR"
if [ -e "$TARGET" ]; then
  if [ "$FORCE" -eq 1 ]; then
    rm -f "$TARGET"
  else
    echo "wt: target already exists: $TARGET" >&2
    if [ -L "$TARGET" ]; then
      echo "wt: target is a symlink to: $(readlink "$TARGET")" >&2
    fi
    echo "wt: refusing to overwrite. Re-run with: ./scripts/install.sh --force" >&2
    exit 1
  fi
fi
GOBIN="$INSTALL_DIR" go install -ldflags "-X wt/internal/buildinfo.Version=$VERSION" ./cmd/wt

echo "Done."
if [ -x "$INSTALL_DIR/wt" ]; then
  if "$INSTALL_DIR/wt" --version >/dev/null 2>&1; then
    echo "Installed: $("$INSTALL_DIR/wt" --version)"
  else
    echo "Installed: wt (unknown version)"
  fi
fi
echo "Tip: ensure your Go bin dir is in PATH (e.g. \$(go env GOPATH)/bin or GOBIN)."
