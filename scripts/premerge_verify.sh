#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

version_file="./VERSION"
release_notes="./docs/release/notes.md"
gomod_file="./go.mod"
expected_module="github.com/es5h/wt"

if [ ! -f "$version_file" ]; then
  echo "premerge: missing VERSION file" >&2
  exit 1
fi

version="$(tr -d ' \t\r\n' <"$version_file")"
if [ -z "$version" ]; then
  echo "premerge: VERSION is empty" >&2
  exit 1
fi

if ! printf "%s" "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+([\-][0-9A-Za-z.-]+)?([+][0-9A-Za-z.-]+)?$'; then
  echo "premerge: VERSION is not valid semver: $version" >&2
  exit 1
fi

if [ ! -f "$release_notes" ]; then
  echo "premerge: missing release notes: $release_notes" >&2
  exit 1
fi

if ! grep -Eq '^## Unreleased[[:space:]]*$' "$release_notes"; then
  echo "premerge: release notes missing '## Unreleased' section: $release_notes" >&2
  exit 1
fi

if [ ! -f "$gomod_file" ]; then
  echo "premerge: missing go.mod file" >&2
  exit 1
fi

if ! grep -Eq "^module[[:space:]]+$expected_module\$" "$gomod_file"; then
  echo "premerge: go.mod module path must be '$expected_module'" >&2
  exit 1
fi

exit 0
