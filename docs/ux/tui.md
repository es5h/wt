# TUI picker spec

## Goal
여러 worktree 후보 중 하나를 빠르게 선택할 수 있는 최소 공통 picker 기반을 제공한다.

현재 범위:
- 이번 변경에서는 `wt path --tui`, `wt remove --tui`, `wt prune --tui`를 실제로 연결한다.
- picker 상태/필터/선택 로직은 공통 패키지로 유지하고, 명령별 semantics와 safety rule은 각 명령이 별도로 책임진다.

## Entry
- `wt path --tui` → 전체 목록에서 선택
- `wt path <query> --tui` → 기존 `wt path` 매칭 결과가 2개 이상일 때만 picker로 선택
- `wt remove --tui` → 전체 목록에서 선택한 뒤 기존 remove safety/confirm 흐름으로 진행
- `wt remove <query> --tui` → 기존 `wt remove` 매칭 결과가 2개 이상일 때만 picker로 선택
- `wt prune --tui` → prunable entry만 preview
- `wt prune --tui --apply` → prunable entry preview 후 confirm prompt를 거쳐 prune 실행

query가 있을 때:
- 매칭 0개: 기존 `no matches` 에러를 그대로 반환한다.
- 매칭 1개: picker를 띄우지 않고 바로 해당 명령의 기존 단일 선택 흐름으로 진행한다.
- 매칭 2개 이상: 해당 후보만 picker에 넣고, `query`를 초기 filter 값으로 사용한다.

`wt remove --tui` 추가 규칙:
- `query`가 없으면 전체 registered worktree 목록을 picker에 넣는다.
- picker 선택 뒤 실제 삭제 전에는 기존 `wt remove` 확인 단계가 반드시 선행된다.
- `--dry-run`이면 picker 뒤 preview만 출력하고 확인/삭제는 하지 않는다.
- `--force`이면 picker 뒤 추가 확인 없이 즉시 삭제한다.
- current worktree, primary worktree, prunable target은 picker에서 선택될 수 있더라도 삭제 단계로 진행하지 않고 명확한 에러/안내를 반환한다.

비활성 조건:
- `stdin` 또는 화면 렌더링에 사용하는 `stderr`가 터미널이 아니면 TUI 실행 금지
- 이 경우 명령은 non-zero exit + 명확한 에러로 종료한다

참고:
- 선택 결과는 계속 stdout으로 출력하므로, stdout은 파이프/command substitution 이어도 된다.
- TUI 화면은 stdout 오염을 피하기 위해 stderr에 렌더링한다.
- `wt prune --tui`는 interactive preview 자체가 목적이므로, picker 선택 결과를 별도 stdout 값으로 내보내지 않는다. 닫은 뒤에는 기존 `wt prune` text 출력 규칙만 적용한다.

관련 CLI 스펙은 `docs/spec/cli.md` 참고.

## UI layout
- 상단: filter 입력
- 본문: 후보 리스트(브랜치/표시 이름, path, 짧은 HEAD, locked/prunable 같은 짧은 메타)
- 하단: match 개수와 키 도움말

`wt prune --tui` row 규칙:
- 목록에는 `prunable=true` entry만 포함한다.
- label은 branch 우선, branch가 없으면 path basename을 사용한다.
- detail은 full path를 보여준다.
- meta에는 최소한 `prunable`과 `pruneReason`을 함께 노출한다.

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

`wt remove --tui` 선택 확정 시:
- `--dry-run`: stdout에 기존 `wt remove` preview line 출력
- interactive confirm: stderr에 기존 확인 프롬프트 출력 후 승인 시 삭제
- `--json`: stdout 스키마는 기존 `wt remove`와 동일

`wt prune --tui` 선택 확정 시:
- preview-only(`--apply` 없음): preview를 닫고 기존 `would-prune ...` text 출력으로 돌아간다.
- `--apply`: preview를 닫은 뒤 `Prune <N> stale worktree entr(y|ies) with git worktree prune --expire now? [y/N]` confirm prompt를 표시하고, 승인 시에만 prune을 실행한다.

취소 시:
- stderr: 취소 메시지
- exit code: `130`

## Implementation note
- 리스트/필터/선택 상태는 순수 모델로 유지한다.
- 터미널 raw mode/렌더링은 얇은 러너로 분리한다.
- 의존성은 기존 `golang.org/x/term`만 사용한다.
