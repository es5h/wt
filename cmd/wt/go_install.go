package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

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
