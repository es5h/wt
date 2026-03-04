#!/usr/bin/env python3
import json
import os
import re
import sys


def main() -> int:
    event_path = os.environ.get("GITHUB_EVENT_PATH", "")
    if not event_path or not os.path.exists(event_path):
        # Not running in GitHub Actions; do nothing.
        return 0

    with open(event_path, "r", encoding="utf-8") as f:
        event = json.load(f)

    pr = event.get("pull_request") or {}
    body = pr.get("body") or ""
    body = body.replace("\r\n", "\n")

    if not body.strip():
        print("PR body is empty. Include an E2E section or 'N/A' with a reason.", file=sys.stderr)
        return 1

    # Require an explicit E2E section to prevent regressions in contributor habits.
    # Allow either "## E2E guide" (recommended) or "## E2E".
    if not re.search(r"(?m)^##\\s+E2E(\\s+guide)?\\s*$", body):
        print("Missing required PR section: '## E2E guide' (or '## E2E').", file=sys.stderr)
        return 1

    # Accept either commands or explicit N/A+reason.
    # We only enforce presence; content can be iterated per project needs.
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

