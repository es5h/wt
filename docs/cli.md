# CLI spec (draft)

이 문서는 `wt`의 명령/옵션 동작을 “테스트 가능한 규칙” 수준으로 고정하기 위한 초안입니다.

## Global rules
- Git 컨텍스트(repo root)는 현재 디렉토리에서 `git rev-parse --show-toplevel`로 결정한다.
- 정상 출력은 stdout, 에러는 stderr.
- 사람이 읽는 출력과 스크립트 출력이 충돌하면 `--json` 또는 `--porcelain`로 분리한다.

## `wt list`
목표: 현재 repo의 worktree를 나열한다.

출력(기본):
- 각 worktree의 `path`, 연결된 `branch`(있으면), `HEAD`(커밋 해시/짧은 해시), `locked` 여부를 표시

옵션(초안):
- `--json`: 구조화 출력
- `--porcelain`: git처럼 안정적 포맷(필드 고정, 파싱 용도)

## `wt goto [query]`
목표: query로 worktree를 선택하고 “경로”를 stdout으로 출력한다.

규칙:
- 기본 모드는 **경로만 출력**한다(추가 텍스트/색상 금지).
- 후보가 0개면 non-zero exit + stderr에 후보/가이드 출력.
- 후보가 1개면 자동 선택.
- 후보가 2개 이상이면:
  - 기본은 TUI/interactive 여부에 따라 다르게 동작(아래 참고)

옵션(초안):
- `--create`: query에 해당하는 브랜치 worktree가 없으면 생성
- `--tui`: 후보가 여러 개면 TUI로 선택(비-interactive 환경이면 에러)
- `--no-tui`: 후보가 여러 개면 에러(스크립트 안전)
- `--json`: 선택 결과를 json으로 출력(예: `{path, branch, reason}`)

TUI 기본 동작(초안):
- 터미널이면 `wt goto`(query 생략) 시 TUI를 기본으로 고려
- 파이프/리다이렉트(`stdout`이 터미널이 아님)면 TUI를 자동 비활성화

## `wt create <branch>`
목표: `<branch>`에 대한 worktree를 생성한다.

옵션(초안):
- `--path <dir>`: 생성 경로 지정(기본은 정책에 따름)
- `--track origin/<branch>`: 원격을 추적하는 브랜치로 생성
- `--dry-run`: 실행될 git 커맨드/경로만 출력

## `wt remove <name>`
목표: worktree를 제거한다.

규칙(초안):
- 기본은 확인 프롬프트(또는 `--force`).
- 브랜치 삭제는 기본 동작이 아니다(별도 옵션).

옵션(초안):
- `--force`: 확인 생략
- `--dry-run`

## `wt init <shell>`
목표: `goto`가 `cd`될 수 있도록 셸 함수/alias를 출력한다.

지원 셸(초안): `zsh`, `bash`, `fish`

예시(개념):
- `wt init zsh` → stdout에 function 정의를 출력(사용자가 rc에 추가)

