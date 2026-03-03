package worktree

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Worktree struct {
	Path, HEAD, Branch         string
	Detached, Locked, Prunable bool
	LockReason, PruneReason    string
}

func ParsePorcelain(r io.Reader) ([]Worktree, error) {
	var (
		lineNo int
		out    []Worktree
		cur    *Worktree
	)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	finalize := func() error {
		if cur == nil {
			return nil
		}
		if strings.TrimSpace(cur.Path) == "" {
			return fmt.Errorf("invalid porcelain: missing worktree path")
		}
		out = append(out, *cur)
		cur = nil
		return nil
	}

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSuffix(scanner.Text(), "\r")
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		key, rest, hasRest := strings.Cut(line, " ")

		if hasRest {
			rest = strings.TrimLeft(rest, " ")
		} else {
			rest = ""
		}

		switch key {
		case "worktree":
			err := finalize()
			if err != nil {
				return nil, err
			}
			cur = &Worktree{Path: rest}
		default:
			if cur == nil {
				return nil, fmt.Errorf("invalid porcelain at line %d: unknown line before worktree: %q", lineNo, line)
			}
		}

		switch key {
		case "HEAD":
			cur.HEAD = rest
		case "branch":
			cur.Branch = rest
		case "detached":
			cur.Detached = true
		case "locked":
			cur.Locked = true
			cur.LockReason = rest
		case "prunable":
			cur.Prunable = true
			cur.PruneReason = rest
		default:
			if cur == nil {
				return nil, fmt.Errorf("invalid porcelain at line %d: unknown line before worktree: %q", lineNo, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := finalize(); err != nil {
		return nil, err
	}

	return out, nil
}
