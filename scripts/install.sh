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
echo "Note: release install path is 'go install github.com/es5h/wt/cmd/wt@latest'."

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
GOBIN="$INSTALL_DIR" go install -ldflags "-X github.com/es5h/wt/internal/buildinfo.Version=$VERSION" ./cmd/wt

echo "Done."
if [ -x "$INSTALL_DIR/wt" ]; then
  if "$INSTALL_DIR/wt" --version >/dev/null 2>&1; then
    echo "Installed: $("$INSTALL_DIR/wt" --version)"
  else
    echo "Installed: wt (unknown version)"
  fi
fi
# Verify INSTALL_DIR is in PATH; offer to fix if not.
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo "" >&2
    echo "WARNING: '$INSTALL_DIR' is not in your PATH." >&2
    echo "The 'wt' command will not be found until you add it." >&2
    echo "" >&2

    # Detect current shell and its rc file.
    CURRENT_SHELL="$(basename "${SHELL:-sh}")"
    RC_FILE=""
    PATH_LINE=""
    case "$CURRENT_SHELL" in
      zsh)
        RC_FILE="$HOME/.zshrc"
        PATH_LINE="export PATH=\"$INSTALL_DIR:\$PATH\""
        ;;
      bash)
        RC_FILE="$HOME/.bashrc"
        PATH_LINE="export PATH=\"$INSTALL_DIR:\$PATH\""
        ;;
      fish)
        RC_FILE="$HOME/.config/fish/config.fish"
        PATH_LINE="fish_add_path $INSTALL_DIR"
        ;;
      *)
        RC_FILE="$HOME/.profile"
        PATH_LINE="export PATH=\"$INSTALL_DIR:\$PATH\""
        ;;
    esac

    # Ask user whether to append PATH config automatically.
    if [ -t 0 ] && [ -n "$RC_FILE" ]; then
      printf "Add '$INSTALL_DIR' to PATH in %s? [y/N] " "$RC_FILE" >&2
      read -r REPLY </dev/tty
      case "$REPLY" in
        [yY]|[yY][eE][sS])
          echo "" >> "$RC_FILE"
          echo "# Added by wt installer" >> "$RC_FILE"
          echo "$PATH_LINE" >> "$RC_FILE"
          echo "Added to $RC_FILE" >&2
          echo "Run 'source $RC_FILE' or open a new terminal to apply." >&2
          ;;
        *)
          echo "Skipped. You can add it manually:" >&2
          echo "  echo '$PATH_LINE' >> $RC_FILE" >&2
          echo "" >&2
          echo "  Or run directly: $INSTALL_DIR/wt" >&2
          ;;
      esac
    else
      # Non-interactive: just print instructions.
      echo "  Add this line to $RC_FILE:" >&2
      echo "    $PATH_LINE" >&2
      echo "" >&2
      echo "  Or run directly: $INSTALL_DIR/wt" >&2
    fi
    ;;
esac
