#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

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

go install -ldflags "-X wt/internal/buildinfo.Version=$VERSION" ./cmd/wt

echo "Done."
if command -v wt >/dev/null 2>&1; then
  if wt --version >/dev/null 2>&1; then
    echo "Installed: $(wt --version)"
  else
    echo "Installed: wt (unknown version)"
  fi
fi
echo "Tip: ensure your Go bin dir is in PATH (e.g. \$(go env GOPATH)/bin or GOBIN)."
