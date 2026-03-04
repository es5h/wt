package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type event struct {
	PullRequest *struct {
		Body string `json:"body"`
	} `json:"pull_request"`
}

func main() {
	eventPath := strings.TrimSpace(os.Getenv("GITHUB_EVENT_PATH"))
	if eventPath == "" {
		// Not running in GitHub Actions.
		os.Exit(0)
	}
	b, err := os.ReadFile(eventPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read GITHUB_EVENT_PATH=%s: %v\n", eventPath, err)
		os.Exit(1)
	}

	var e event
	if err := json.Unmarshal(b, &e); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse GitHub event json: %v\n", err)
		os.Exit(1)
	}

	body := ""
	if e.PullRequest != nil {
		body = e.PullRequest.Body
	}
	body = strings.ReplaceAll(body, "\r\n", "\n")

	if strings.TrimSpace(body) == "" {
		fmt.Fprintln(os.Stderr, "PR body is empty. Include an E2E section or 'N/A' with a reason.")
		os.Exit(1)
	}

	// Require an explicit E2E section to prevent regressions in contributor habits.
	// Allow either "## E2E guide" (recommended) or "## E2E".
	reHeader := regexp.MustCompile(`(?m)^##\s+E2E(\s+guide)?\s*$`)
	if !reHeader.MatchString(body) {
		fmt.Fprintln(os.Stderr, "Missing required PR section: '## E2E guide' (or '## E2E').")
		os.Exit(1)
	}

	os.Exit(0)
}
