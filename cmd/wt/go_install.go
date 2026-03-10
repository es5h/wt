package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/es5h/wt/internal/runner"
)

func installWithGo(ctx context.Context, workDir string, installDir string, packageRef string) (runner.Result, error) {
	cmd := exec.CommandContext(ctx, "go", "install", packageRef)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", installDir))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := runner.Result{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}
	return result, err
}

func resolveLatestVersion(ctx context.Context, workDir string, modulePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-json", "-versions", modulePath)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("go list failed: %s", msg)
	}

	var module struct {
		Versions []string `json:"Versions"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &module); err != nil {
		return "", fmt.Errorf("failed to parse go list output: %w", err)
	}
	if len(module.Versions) == 0 {
		return "", fmt.Errorf("no released versions found for %s", modulePath)
	}
	return module.Versions[len(module.Versions)-1], nil
}
