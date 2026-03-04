# wt

`wt`는 `git worktree`를 더 쉽게 쓰기 위한 CLI 헬퍼입니다.

> 상태: WIP (스펙/문서부터 정리 중)

## Install (개발용)
아직 릴리즈/배포는 WIP입니다.

- 로컬 설치(권장): `./scripts/install.sh`
- 또는: `go install ./cmd/wt`

## Docs
- 명령/옵션 스펙: `docs/spec/cli.md`
- 셸 통합/자동완성(초안): `docs/ux/shell.md`
- TUI(초안): `docs/ux/tui.md`
- 릴리즈 노트: `docs/release/notes.md`

기타 내부 문서(정책/로드맵 등)는 `docs/README.md`에서 확인할 수 있습니다.

## Quickstart (개발용)
- 빌드: `make build`
- 테스트: `make test`
- 실행: `make run ARGS="--help"`

`make build`/`make test`는 `gofmt`/`go fix`가 깨끗한 상태(`make check`)를 요구합니다.

## Optional integrations (수동 설정)
사용자 데이터/셸 설정 보호를 위해, 설치 스크립트는 자동완성/셸 통합/TUI를 자동 설치하지 않습니다.

- 셸 통합(`wt path`를 `cd`로 연결): `docs/ux/shell.md` 참고 (예: `wtg() { cd "$(wt path "$@")" || return; }`)
- 자동완성(completion): `docs/ux/shell.md` 참고 (설치 방식은 셸별로 다름)
- TUI picker: `docs/ux/tui.md` 참고 (현재 초안/미구현일 수 있음)
