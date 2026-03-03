# wt

`wt`는 `git worktree`를 더 쉽게 쓰기 위한 CLI 헬퍼입니다.

> 상태: WIP (스펙/문서부터 정리 중)

## Goals
- `wt list`: 현재 Git 컨텍스트의 worktree 목록 보기
- `wt goto <name>`: worktree 선택 → **경로를 stdout으로 출력** (셸에서 `cd "$(wt goto ...)"` 용도)
- `wt goto <name> --create`: 없으면 worktree 생성 후 경로 출력
- 선택형 TUI: `wt goto`에서 목록을 보고 고르기(vim-ish 키바인딩 포함)
- 셸 자동완성: `wt goto <TAB>` 등으로 후보 제안

## Docs
- 명령/옵션 스펙: `docs/cli.md`
- 셸 통합(초안): `docs/shell-completion.md`
- TUI(초안): `docs/tui.md`

## Quickstart (개발용)
- 빌드: `go build ./...`
- 실행: `go run . --help`

