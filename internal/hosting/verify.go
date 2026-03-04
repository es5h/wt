package hosting

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"wt/internal/runner"
)

type Provider string

const (
	ProviderUnknown Provider = "unknown"
	ProviderGitHub  Provider = "github"
	ProviderGitLab  Provider = "gitlab"
)

type VerifyResult struct {
	Provider Provider
	Kind     string
	Merged   *bool
	Reason   string
}

func DetectProvider(remoteURL string) Provider {
	normalized := strings.ToLower(strings.TrimSpace(remoteURL))
	switch {
	case strings.Contains(normalized, "github.com"):
		return ProviderGitHub
	case strings.Contains(normalized, "gitlab.com"), strings.Contains(normalized, "gitlab"):
		return ProviderGitLab
	default:
		return ProviderUnknown
	}
}

func VerifyMerged(ctx context.Context, r runner.Runner, repoRoot string, provider Provider, branch string, baseRef string) (VerifyResult, error) {
	switch provider {
	case ProviderGitHub:
		return verifyGitHubMerged(ctx, r, repoRoot, branch, baseRef)
	case ProviderGitLab:
		return VerifyResult{Provider: ProviderGitLab, Kind: "mr", Reason: "unsupported-provider"}, nil
	default:
		return VerifyResult{Provider: provider, Kind: "unknown", Reason: "unsupported-provider"}, nil
	}
}

func verifyGitHubMerged(ctx context.Context, r runner.Runner, repoRoot string, branch string, baseRef string) (VerifyResult, error) {
	if strings.TrimSpace(branch) == "" {
		return VerifyResult{Provider: ProviderGitHub, Kind: "pr", Reason: "no-branch"}, nil
	}

	ghBin, ok := findGitHubCLI()
	if !ok {
		return VerifyResult{Provider: ProviderGitHub, Kind: "pr", Reason: "gh-auth-unavailable"}, nil
	}

	if _, err := r.Run(ctx, repoRoot, ghBin, "auth", "status"); err != nil {
		return VerifyResult{Provider: ProviderGitHub, Kind: "pr", Reason: "gh-auth-unavailable"}, nil
	}

	args := []string{"pr", "list", "--state", "merged", "--head", branch, "--json", "number", "--limit", "1"}
	if shortBase := shortRefName(baseRef); shortBase != "" {
		args = append(args, "--base", shortBase)
	}

	res, err := r.Run(ctx, repoRoot, ghBin, args...)
	if err != nil {
		return VerifyResult{Provider: ProviderGitHub, Kind: "pr", Reason: "gh-pr-query-failed"}, nil
	}

	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(res.Stdout, &prs); err != nil {
		return VerifyResult{Provider: ProviderGitHub, Kind: "pr", Reason: "gh-invalid-json"}, nil
	}

	merged := len(prs) > 0
	return VerifyResult{Provider: ProviderGitHub, Kind: "pr", Merged: &merged}, nil
}

func findGitHubCLI() (string, bool) {
	if explicit := strings.TrimSpace(os.Getenv("WT_GH_BIN")); explicit != "" {
		return explicit, true
	}

	if path, err := exec.LookPath("gh"); err == nil {
		return path, true
	}

	return "", false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func shortRefName(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "refs/heads/")
	ref = strings.TrimPrefix(ref, "refs/remotes/")
	ref = strings.TrimPrefix(ref, "origin/")
	ref = strings.TrimPrefix(ref, "upstream/")
	return ref
}
