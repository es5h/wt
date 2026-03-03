# AGENTS.md (repo: wt)

이 파일은 `wt` 레포에서 자동화 코딩 에이전트가 일관되고 안전하게 작업하기 위한 로컬 규칙/런북입니다.

## Project
- 목적: `git worktree`를 더 쉽게 쓰기 위한 CLI 헬퍼
- 핵심 UX:
  - `wt list`: 현재 Git 컨텍스트의 worktree 목록
  - `wt goto <name>`: worktree 선택(이름/브랜치/부분매칭) → **경로를 stdout으로 출력**
  - `wt goto <name> --create`: 없으면 worktree 생성 후 경로 출력
  - (추가) `wt create`, `wt remove`, `wt prune`, `wt init <shell>`

## Non-goals (명시적으로 안 하는 것)
- Git 자체를 대체하는 복잡한 UI/대규모 설정 시스템
- 사용자의 로컬 환경을 파괴할 수 있는 기본 동작(삭제/초기화 등)은 반드시 opt-in

## Repo layout
- 현재: `main.go` (임시/스캐폴딩)
- 권장(성장 시):
  - `cmd/wt/` : 바이너리 엔트리포인트(`main`)
  - `internal/` : git/worktree 로직, resolver, config
  - `scripts/` : 개발 보조 스크립트(선택)

## Tooling
- Go 버전은 `go.mod`를 따른다.
- 포맷: `gofmt` (필수)
- 의존성: 꼭 필요할 때만 추가. CLI 프레임워크를 쓰면 `cobra` 또는 `urfave/cli` 중 하나로 통일(혼용 금지).

## Common commands
- 빌드: `go build ./...`
- 실행(현 구조): `go run . --help`
- 실행(권장 구조): `go run ./cmd/wt --help`
- 테스트: `go test ./...`

## Output/UX rules
- `wt goto ...` 계열은 셸에서 `cd "$(wt goto ...)"`가 가능하도록 **경로만** 출력하는 모드를 기본으로 유지한다.
- 사람이 보는 출력과 스크립트용 출력이 충돌하면:
  - 기본은 사람이 보기 좋게
  - `--json` 또는 `--porcelain`(스크립트용) 옵션을 제공
- 에러 메시지는 stderr, 정상 출력은 stdout.

## Safety rules (중요)
- 사용자가 명시하지 않은 파괴적 동작 금지:
  - `git reset --hard`, `git clean -fdx`, 임의의 브랜치 삭제, 강제 prune 등
- `wt remove`/`--create` 같이 파일/폴더를 건드리는 동작은:
  - 대상(worktree path, branch)을 명확히 출력
  - 기본적으로 확인(confirmation) 또는 드라이런(`--dry-run`) 지원을 고려

## Worktree policy (초안)
- “Git 컨텍스트”는 현재 디렉토리에서 `git rev-parse --show-toplevel`로 결정한다.
- 기본 worktree 루트 경로 정책은 한 가지로 고정하고 문서화한다(예: repo 상위 `../.wt/<repo>/...`).
- 브랜치명 → 폴더명 정규화 규칙(`/` → `-`, 공백/특수문자 처리, 충돌 시 suffix)을 유틸로 통일한다.

## Shell integration
- `wt goto`는 프로세스가 직접 `cd`할 수 없으므로, `wt init zsh|bash|fish`가 셸 함수/alias를 출력하는 방식으로 통합한다.
- 셸 스크립트는 가능한 한 idempotent(중복 적용해도 안전)하게 만든다.

## Specs (docs)
- CLI/옵션 스펙(초안): `docs/cli.md`
- 셸 통합/완성(초안): `docs/shell-completion.md`
- TUI 스펙(초안): `docs/tui.md`
