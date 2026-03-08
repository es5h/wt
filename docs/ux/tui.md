# TUI flows

이 문서는 현재 구현된 TUI 진입 조건과 실제 사용 흐름만 다룬다.

## Global Rules

- TUI는 `stdin`과 `stderr`가 모두 TTY일 때만 동작한다.
- 화면은 `stderr`에 렌더링한다.
- 최종 결과는 기존 명령 규칙대로 `stdout`에 남긴다.
- 취소 키는 `Esc`와 `Ctrl+C`다.
- 취소 exit code는 `130`이다.

## Common UI

- 상단: filter 입력
- 본문: 후보 리스트
- 하단: match 수와 키 도움말
- 긴 줄(branch/path/help)은 현재 터미널 가로폭에 맞춰 말줄임(`...`) 처리한다.

공통 키:

- 이동: `Up`, `Down`, `Ctrl-P`, `Ctrl-N`
- 페이지: `PageUp`, `PageDown`
- 처음/끝: `Home`, `End`
- 선택/진행: `Enter`
- 취소: `Esc`, `Ctrl+C`

## `wt path --tui`

진입:

- `wt path --tui`: 현재 repo의 전체 registered worktree 목록
- `wt path <query> --tui`: 먼저 기존 매칭 규칙 적용 후, 다중 후보일 때만 TUI

동작:

- 매칭 0개면 `wt path: no matches for "<query>"`로 실패한다.
- 매칭 1개면 TUI를 띄우지 않고 바로 path-only 출력한다.
- 매칭 2개 이상이면 해당 후보만 TUI에 넣고 `query`를 초기 filter 값으로 사용한다.
- 선택 성공 시 기본 모드는 path만 `stdout`에 출력한다.
- `--json`이면 `{path, branch}`를 출력한다.

non-TTY:

- `wt path: --tui requires a TTY on stdin and stderr`

## `wt remove --tui`

진입:

- `wt remove --tui`: 전체 registered worktree 목록에서 선택
- `wt remove <query> --tui`: 다중 후보일 때만 TUI, 단일 후보면 바로 선택

동작:

- 선택 뒤에도 기존 remove safety를 그대로 적용한다.
- current worktree, primary worktree, prunable target은 선택 후 제거 단계에서 거부된다.
- `--dry-run`이면 preview line만 출력한다.
- `--force`이면 추가 confirm 없이 삭제한다.
- interactive TTY에서 `--dry-run`과 `--force`가 모두 없으면 confirm prompt를 보여준다.

confirm semantics:

- 프롬프트는 `stderr`에 출력된다.
- 메시지 형식은 `Remove worktree <path> (<branch>)? [y/N]`
- `y`, `yes`만 승인으로 처리한다.
- 거부하면 `wt remove: aborted`로 종료한다.

취소:

- picker 취소는 `wt remove: selection cancelled`
- exit code `130`

## `wt prune --tui`

역할:

- prunable entry 전용 interactive preview

진입:

- `wt prune --tui`
- `wt prune --tui --apply`

동작:

- 후보는 `git worktree list`에서 `prunable`인 entry만 포함한다.
- 각 row는 branch 또는 basename, path, `prunable`, `pruneReason`을 보여준다.
- `wt prune --tui`는 preview를 닫은 뒤에도 기존 text 출력(`would-prune ...`)을 계속 사용한다.
- `wt prune --tui --apply`는 preview 후 confirm prompt를 거친 뒤 `git worktree prune --expire now`를 한 번 실행한다.
- `--json`과 함께 쓸 수 없다.

confirm semantics:

- 프롬프트는 `stderr`에 출력된다.
- 메시지는 대상 개수에 따라 `Prune 1 stale worktree entry ...` 또는 `Prune N stale worktree entries ...` 형식이다.
- `y`, `yes`만 승인한다.
- 거부하면 `wt prune: aborted`로 종료한다.

취소:

- preview 취소는 `wt prune: preview cancelled`
- exit code `130`

## Non-TTY Policy

아래 조합은 모두 usage error로 거부한다.

- `wt path --tui`
- `wt remove --tui`
- `wt prune --tui`

이때 메시지는 각 명령 이름을 포함한 `--tui requires a TTY on stdin and stderr` 형식을 사용한다.
