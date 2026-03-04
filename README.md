# wt

`wt`는 실제 작업용 `git worktree` helper CLI입니다. 현재 `main`에 반영된 기능 기준으로 `list`, `path`, `create`, `root`, `run`, `remove`, `prune`, `cleanup`, `init`, TUI picker 흐름을 제공합니다.

## Quick Start

설치:

```sh
./scripts/install.sh
```

또는:

```sh
go install ./cmd/wt
```

셸 helper 추가:

```sh
eval "$(wt init zsh)"
```

기본 흐름:

```sh
wt list
wt path <query>
wt create <branch>
wt remove <query> --dry-run
wt prune
```

## What It Does

- 현재 repo의 registered worktree를 조회한다: `wt list`, `wt list --verify`, `wt list --verify-hosting`
- worktree 경로를 안전하게 선택한다: `wt path`, `wt path --tui`, `wt root`, `wt run`
- 없으면 생성하거나 기존 브랜치에 attach 한다: `wt create`, `wt path --create`
- stale entry와 안전한 제거 대상을 분리해서 정리한다: `wt prune`, `wt remove`, `wt cleanup`
- 셸 이동 helper와 completion 연동을 제공한다: `wt init <shell>`, `wt completion <shell>`

## Interactive Flows

- `wt path --tui`: 전체 목록 또는 다중 매칭 후보에서 worktree를 고른다.
- `wt remove --tui`: 대상을 고른 뒤 기존 remove safety와 confirm 흐름을 그대로 적용한다.
- `wt prune --tui`: prunable entry를 TUI로 미리 보고, `--apply`일 때만 confirm 뒤 prune 한다.

TUI는 `stdin`과 `stderr`가 모두 TTY일 때만 동작하며, 화면은 `stderr`, 최종 결과는 `stdout`에 유지된다.

## User Docs

- CLI 규칙: [docs/spec/cli.md](docs/spec/cli.md)
- 셸 helper / completion: [docs/ux/shell.md](docs/ux/shell.md)
- TUI 사용 흐름: [docs/ux/tui.md](docs/ux/tui.md)
- 변경 이력: [docs/release/notes.md](docs/release/notes.md)

## Development

```sh
make build
make test
make premerge
```

## License

MIT. See [LICENSE](LICENSE).
