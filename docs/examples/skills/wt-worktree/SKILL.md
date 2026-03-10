---
name: wt-worktree
description: Use wt commands to manage git worktrees safely and consistently, especially for parallel agent workflows.
---

# wt-worktree

## When to use
- worktree 생성/탐색/정리가 필요한 작업
- 병렬 에이전트 실행 시 브랜치별 격리가 필요한 작업

## Required checks
- `wt --version`
- `wt list`

## Standard flow
1. `wt path --create <branch>`로 작업 경로 확보
2. 필요한 작업 실행
3. 필요 시 `wt cleanup`/`wt prune`로 정리

## Safety
- destructive 옵션은 사용자 명시 요청 시에만 사용
- remove/prune는 기본적으로 preview 먼저 실행
