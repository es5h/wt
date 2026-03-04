# TUI picker spec

## Goal
여러 worktree 후보 중 하나를 빠르게 선택할 수 있는 최소 공통 picker 기반을 제공한다.

현재 범위:
- 이번 변경에서는 `wt path --tui`만 실제로 연결한다.
- picker 상태/필터/선택 로직은 공통 패키지로 분리해 이후 `wt remove`/`wt prune`가 재사용할 수 있게 둔다.

## Entry
- `wt path --tui` → 전체 목록에서 선택
- `wt path <query> --tui` → 필터링된 후보를 picker로 선택

비활성 조건:
- `stdin` 또는 화면 렌더링에 사용하는 `stderr`가 터미널이 아니면 TUI 실행 금지
- 이 경우 명령은 non-zero exit + 명확한 에러로 종료한다

참고:
- 선택 결과는 계속 stdout으로 출력하므로, stdout은 파이프/command substitution 이어도 된다.
- TUI 화면은 stdout 오염을 피하기 위해 stderr에 렌더링한다.

관련 CLI 스펙은 `docs/spec/cli.md` 참고.

## UI layout
- 상단: filter 입력
- 본문: 후보 리스트(브랜치/표시 이름, path, 짧은 HEAD, locked/prunable 같은 짧은 메타)
- 하단: match 개수와 키 도움말

## Keybindings
- 이동: `Up/Down`, `Ctrl+p/Ctrl+n`
- 페이지: `PageUp/PageDown`
- 처음/끝: `Home/End`
- 선택: `Enter`
- 취소: `Esc` 또는 `Ctrl+c`
- 검색: 타이핑 시 실시간 필터

현재 범위(out-of-scope):
- 다중 선택
- 마우스
- 토글성 뷰 옵션
- `--create`와 결합된 picker 흐름
- `query` 없이 `--no-tui`를 주는 강제 non-TUI 모드

## Selection result
선택 확정 시:
- stdout: 선택된 worktree의 path만 출력(기본 모드)
- exit code: 0

취소 시:
- stderr: 취소 메시지
- exit code: `130`

## Implementation note
- 리스트/필터/선택 상태는 순수 모델로 유지한다.
- 터미널 raw mode/렌더링은 얇은 러너로 분리한다.
- 의존성은 기존 `golang.org/x/term`만 사용한다.
