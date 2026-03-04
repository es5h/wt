package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func preflightCreateTargetPath(commandName string, targetPath string) error {
	path := strings.TrimSpace(targetPath)
	if path == "" {
		return fmt.Errorf("%s: target path cannot be empty", commandName)
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("%s: failed to inspect target path: %w", commandName, err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("%s: failed to inspect target path: %w", commandName, err)
		}
		if len(entries) > 0 {
			return usageError(fmt.Errorf("%s: target path is a non-empty directory: %s", commandName, path))
		}
		return nil
	}

	if info.Mode().IsRegular() {
		return usageError(fmt.Errorf("%s: target path is an existing file: %s", commandName, path))
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return usageError(fmt.Errorf("%s: target path is a symbolic link (unsupported): %s", commandName, path))
	}

	return usageError(fmt.Errorf("%s: target path has unsupported file type: %s", commandName, path))
}
