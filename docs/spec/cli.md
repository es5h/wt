# CLI spec

이 문서는 `main`에 실제 구현된 `wt` CLI의 사용자-facing 규칙만 기록한다.

## Global Rules

- Git 컨텍스트는 현재 디렉토리에서 결정한다.
- 정상 출력은 `stdout`, 에러와 note는 `stderr`를 사용한다.
- 사람이 보는 출력과 스크립트 출력이 충돌하면 `--json` 또는 `--porcelain`을 제공한다.
- 경로를 돌려주는 명령(`wt path`, `wt root`, `wt create`)의 기본 출력은 path-only 다.
- `--tui` 화면에서 긴 줄(branch/path/help/filter)은 현재 터미널 가로폭에 맞춰 말줄임(`...`) 처리한다.

## `wt list`

현재 repo의 registered worktree를 나열한다.

기본 출력:

- 한 줄에 `basename  branch  short-head  path`를 출력한다.
- 추가 신호가 있으면 `[...]` 마커를 붙인다.
- 마커는 `locked`, `prunable`, `current`, `primary`, `missing-path`, `missing-git`, `merged`, `merged-hosting:<provider>`, `stale`, `safe-remove`, `recommend:prune|remove`를 사용한다.

옵션:

- `--json`: 구조화 JSON 출력
- `--porcelain`: `git worktree list --porcelain` 원문 출력
- `--verify`: path/.git 상태와 base ref merge 여부를 함께 검증
- `--verify-hosting`: hosting(PR/MR) 기준 merge 여부를 함께 검증
- `--base <ref>`: `--verify` 또는 `--verify-hosting`의 기준 ref. 기본값은 `origin/HEAD` 또는 `main`

조합 규칙:

- `--json`과 `--porcelain`은 함께 쓸 수 없다.
- `--porcelain`과 `--verify-hosting`은 함께 쓸 수 없다.

검증 규칙:

- `--json --verify`는 각 항목에 `pathExists`, `dotGitExists`, `valid`, `mergedIntoBase`, `baseRef`를 포함한다.
- detached 또는 branch 없는 항목은 `mergedIntoBase: null`을 사용한다.
- `--verify-hosting`은 GitHub(`gh`)와 GitLab(`glab`)만 지원한다.
- hosting 검증 실패는 명령 전체 실패로 승격하지 않는다.
- hosting 검증이 불가능하면 텍스트 출력에는 note, JSON에는 `mergedViaHosting: null`과 `hostingReason`을 남긴다.
- GitHub 바이너리 탐색 순서는 `WT_GH_BIN` 후 `PATH`다.
- GitLab 바이너리 탐색 순서는 `WT_GLAB_BIN` 후 `PATH`다.

파생 신호:

- `stale=true`: `prunable=true` 이거나 path/.git 누락
- `recommendedAction=prune`: `prunable=true`
- `recommendedAction=remove`: `prunable=false`, `current=false`, `primary=false`, `detached=false`, `locked=false`, `missing-path=false`, `missing-git=false`, 그리고 로컬 merge 또는 hosting merge가 확인된 경우
- `safeToRemove=true`: `recommendedAction=remove`와 같은 안전 기준을 만족한 경우

## `wt path [query]`

query와 매칭되는 registered worktree path를 출력한다.

규칙:

- 선택 기준은 filesystem scan이 아니라 `git worktree list`의 registered entry다.
- 기본 출력은 path-only 다.
- `query` 없이 실행하려면 `--tui`가 필요하다.
- 후보가 0개면 exit code `1`과 함께 `no matches` 에러를 반환한다.
- 후보가 1개면 자동 선택한다.
- 후보가 2개 이상이면 기본 동작은 에러 + 후보 목록이다.
- ambiguous 상황에서 `--tui`를 주면 후보만 TUI로 다시 고른다.
- `wt path <query> --tui`는 `query`를 초기 filter 값으로 사용한다.
- 취소(`Esc`, `Ctrl+C`)는 exit code `130`이다.

옵션:

- `--json`: `{path, branch}`
- `--create`: 없으면 worktree 생성
- `--path <dir>`: `--create` 시 최종 경로 지정
- `--root <dir>`: `--create` 시 기본 root 지정
- `--from <ref>`: `--create` 시 start point 지정
- `--dry-run`: `--create` 실행 계획만 출력(실제 생성 없음)
- `--tui`: 전체 목록 또는 다중 후보에서 TUI 선택
- `--no-tui`: 다중 후보 시에도 TUI 없이 실패

조합 규칙:

- `--tui`와 `--no-tui`는 함께 쓸 수 없다.
- `--tui`와 `--create`는 함께 쓸 수 없다.
- `--path`, `--root`, `--from`, `--dry-run`은 `--create`가 있을 때만 허용된다.
- `query` 없이 `--no-tui`는 허용되지 않는다.

`--create` 규칙:

- 동일 브랜치의 live registered worktree가 있으면 그 path를 그대로 반환한다.
- 로컬 브랜치가 있고 live registered worktree만 없으면 `git worktree add <path> <branch>`로 attach 한다.
- 로컬 브랜치가 없고 `origin/<branch>`가 있으면 `git worktree add -b <branch> <path> origin/<branch>`를 사용한다.
- 둘 다 없으면 `git worktree add -b <branch> <path> <from>`을 사용한다.
- 동일 브랜치 또는 query에 대응되는 registered `prunable` entry가 있으면 자동 복구하지 않고 실패하며 `wt prune --apply`를 안내한다.
- 최종 생성 경로 preflight를 먼저 수행한다:
  - 경로가 없으면 통과
  - 기존 파일이면 usage error(exit code 2)
  - 기존 디렉터리가 비어있으면 통과
  - 기존 디렉터리가 비어있지 않으면 usage error(exit code 2)
  - symbolic link를 포함한 기타 타입은 usage error(exit code 2)
- `--dry-run`도 동일 preflight를 수행하고, 통과 시에만 `stderr`에 preview command를 출력한다.

## `wt root`

현재 Git 컨텍스트의 primary repository root를 출력한다.

규칙:

- 기본 출력은 path-only 다.
- linked worktree 안에서 실행해도 primary repo root를 출력한다.

옵션:

- `--json`: `{root}`

## `wt run <query> -- <cmd...>`

`wt path`와 같은 매칭 규칙으로 worktree를 고른 뒤 그 디렉토리에서 명령을 실행한다.

규칙:

- query 매칭과 ambiguous 처리, exit code는 `wt path`와 같다.
- 기본 모드는 하위 프로세스의 `stdout`/`stderr`와 exit code를 그대로 전달한다.
- `--json` 사용 시 하위 프로세스 출력은 중계하지 않고 `{path, command, exitCode}`만 `stdout`에 쓴다.

옵션:

- `--json`

## `wt create <branch>`

branch용 worktree를 만든다.

규칙:

- 기본 출력은 path-only 다.
- 기본 경로는 `<primary-root>/.wt/<branch>`다.
- `--root`, `WT_ROOT`, repo-local git config `wt.root`는 모두 `<primary-root>` 기준으로 해석한다.
- 우선순위는 `--path` > `--root` > `WT_ROOT` > repo-local `wt.root` > default root 다.
- 동일 브랜치의 live registered worktree가 있으면 그 path를 반환한다.
- 동일 브랜치의 registered `prunable` entry가 있으면 실패하고 `wt prune --apply`를 안내한다.

옵션:

- `--path <dir>`
- `--root <dir>`
- `--from <ref>`
- `--dry-run`

`--dry-run` 규칙:

- 실제 생성 대신 실행될 `git worktree add ...` 명령을 `stderr`에 출력한다.
- 반환값은 실제 생성 시 사용할 path다.
- 실제 실행 전 최종 생성 경로 preflight를 동일하게 수행한다.

## `wt remove [query]`

선택한 worktree를 제거한다.

규칙:

- `query` 없이 실행하려면 `--tui`가 필요하다.
- `--dry-run`이면 preview-only 다.
- `--force`이면 추가 확인 없이 즉시 제거한다.
- `--dry-run`과 `--force`가 모두 없으면 interactive TTY에서만 confirm prompt를 사용한다.
- non-interactive 환경에서는 `--dry-run` 또는 `--force`가 필요하다.
- 현재 실행 중인 worktree는 제거할 수 없다.
- primary worktree는 제거할 수 없다.
- `prunable` entry는 remove 대상이 아니며 `wt prune --apply`를 사용해야 한다.
- 실제 삭제는 `git worktree remove --force <path>`를 사용한다.

옵션:

- `--dry-run`
- `--force`
- `--json`: `{path, branch, action, removed}`
- `--tui`: query 생략 또는 다중 후보 시 TUI 선택

TUI 규칙:

- `wt remove --tui`는 전체 registered worktree 목록을 대상으로 선택한다.
- `wt remove <query> --tui`는 0개면 실패, 1개면 바로 선택, 2개 이상이면 TUI로 고른다.
- TUI를 써도 current/primary/prunable safety rule은 그대로 유지된다.
- 취소는 exit code `130`이다.

출력 규칙:

- text 출력은 `would-remove` 또는 `removed` 한 줄이다.
- interactive confirm prompt는 `stderr`에 `Remove worktree <path> (<branch>)? [y/N]` 형식으로 출력한다.

## `wt prune`

stale/prunable registered entry를 미리 보거나 정리한다.

규칙:

- 기본 동작은 preview-only 다.
- 대상은 `git worktree list`에서 `prunable`로 표시되는 entry만이다.
- `--apply`가 있을 때만 `git worktree prune --expire now`를 실행한다.
- 정상 worktree 디렉토리를 직접 삭제하지 않는다.

옵션:

- `--apply`
- `--json`: `{path, branch, pruneReason, action, removed}` 배열
- `--tui`: prunable entry preview를 TUI로 표시

TUI 규칙:

- `--tui`와 `--json`은 함께 쓸 수 없다.
- `--tui`는 `stdin`과 `stderr`가 모두 TTY일 때만 허용된다.
- `wt prune --tui`는 preview-only 기본값을 유지한다.
- `wt prune --tui --apply`는 preview 후 confirm prompt를 거친 뒤 prune 한다.
- preview 취소는 exit code `130`, confirm 거부는 `wt prune: aborted`다.

출력 규칙:

- text 출력은 `would-prune`, `pruned`, `kept`를 사용한다.
- `--apply` 후에는 prune 전 후보 목록과 prune 후 목록을 비교해 `removed`를 계산한다.

## `wt cleanup`

`wt list`의 추천 신호를 실제 prune/remove 실행과 연결한다.

규칙:

- 기본 동작은 preview-only 다.
- `recommendedAction=prune`는 `wt prune`과 같은 정책으로 처리한다.
- `recommendedAction=remove`는 `safeToRemove=true`인 항목에만 적용한다.
- 실제 prune은 `git worktree prune --expire now`
- 실제 remove는 `git worktree remove --force <path>`

옵션:

- `--apply`
- `--json`

출력 규칙:

- text 출력은 `would-prune`, `would-remove`, `skip`, `pruned`, `removed`, `kept`를 사용한다.
- 각 line은 `action  path  (branch)  [reason]` 형식이다.
- remove 이유는 `merged:<base>` 또는 `merged-hosting:<provider>[#number]`처럼 짧게 출력한다.

## `wt init <shell>`

셸 helper 함수를 `stdout`으로 출력한다.

지원 셸:

- `zsh`
- `bash`
- `fish`

규칙:

- rc 파일을 자동 수정하지 않는다.
- 현재 출력에는 `wtr`, `wtg`, `wcd`가 포함된다.
- `wtr`는 `wt root`, `wtg`와 `wcd`는 `wt path`를 호출해 셸이 직접 `cd`하도록 연결한다.
- `wt init zsh`에는 `_wt` completion이 설치된 경우 `wtg`와 `wcd`를 `wt path` completion에 연결하는 bridge도 포함된다.
