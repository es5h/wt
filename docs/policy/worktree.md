# Policy (draft)

이 문서는 `wt` 내부 정책(경로/정규화 등)을 “구현 근거가 되는 규칙” 형태로 기록합니다.

## Git context
- “Git 컨텍스트(repo root)”는 현재 디렉토리에서 `git rev-parse --show-toplevel`로 결정한다.

## Registered worktree matching
- `wt path` 계열의 선택 기준은 filesystem scan이 아니라 `git worktree list`에 등록된 entry다.
- 기본 `wt path`는 path-only 철학을 유지하기 위해, `wt list --verify` 없이 path 존재 여부나 `.git` 상태를 추가 검사하지 않는다.
- 즉, `wt path`는 “현재 등록된 worktree path를 고르는 명령”이고, filesystem 검증/정리 신호는 `wt list --verify`/`wt prune`가 담당한다.

## Default worktree root
기본 생성 경로 정책:
- 기본 생성 경로는 `<primary-root>/.wt/<branch>` 이다.
- `.wt/`는 레포의 산출물/로컬 작업 디렉토리이므로 git에서 추적하지 않는다(`/.wt/`를 `.gitignore`에 추가).
- 이 기본값은 재현성을 위해 유지하고, 다른 레이아웃은 opt-in 오버라이드로만 적용한다.

`<primary-root>` 결정 규칙:
- `git rev-parse --path-format=absolute --git-common-dir`로 “공유 git 디렉토리”를 구한다.
- 그 부모 디렉토리를 `<primary-root>`로 사용한다.

의도:
- linked worktree 내부에서 `wt create`/`wt path --create`를 실행해도, 기본 경로가 “현재 worktree 아래”로 잡혀 중첩되는 문제(예: `.wt/a/.wt/b/...`)를 방지한다.

## Overrides (opt-in)
기본 정책은 재현성을 위해 고정하되, 사용자/팀 환경에 맞게 “명시적으로” 오버라이드할 수 있게 한다.

우선순위(높음 → 낮음):
1) CLI flag `--root`
2) Environment variable `WT_ROOT`
3) Repo-local git config `wt.root` (`git config --local wt.root ...`)
4) Default policy (`<primary-root>/.wt`)

규칙:
- 이 우선순위는 `wt create`와 `wt path --create`에 동일하게 적용한다.
- `--path`가 지정되면 최종 생성 경로를 직접 지정한 것으로 보고 root 정책보다 우선한다.
- `--root`, `WT_ROOT`, `wt.root` 값이 상대 경로이면 모두 `<primary-root>` 기준으로 해석한다.
- `wt.root`는 repo-local config만 읽는다. global/system git config는 이 정책에 포함하지 않는다.

지원 environment variable:
- `WT_ROOT`: worktree 루트 디렉토리(절대/상대 경로)

repo-local git config 예시:
- `git config --local wt.root ../.wt`
- `git config --local wt.root .worktrees`

주의:
- 자동으로 사용자 홈/셸 설정을 변경하는 동작은 기본값으로 두지 않는다.
- 오버라이드는 “없는 경우에만 적용”이 아니라, 명시적 설정이 있는 경우에만 적용하는 것을 원칙으로 한다.

## Branch name → directory name normalization
- 브랜치명 → 폴더명 정규화 규칙을 유틸로 통일한다.
  - 현재 기본 경로는 git 브랜치의 `/`를 하위 디렉토리로 유지해 `<root>/<branch>` 형태를 만든다.
  - 절대 경로/상위 디렉토리 탈출(`..`)이 되는 브랜치명은 기본 경로 계산에 사용할 수 없다.

## Create safety
- `wt create`와 `wt path --create`는 생성 전에 먼저 `git worktree list`의 registered entry를 확인한다.
- 동일 브랜치의 live registered worktree가 있으면 새 worktree를 만들지 않고 그 registered path를 반환한다.
- 로컬 브랜치가 이미 있고 live registered worktree만 없는 경우에는 새 브랜치를 만들지 않는다. `git worktree add <path> <branch>`로 해당 로컬 브랜치를 attach/check out 한다.
- 로컬 브랜치가 없고 `origin/<branch>`만 있으면 `git worktree add -b <branch> <path> origin/<branch>`를 사용한다.
- 동일 브랜치 또는 query에 대응되는 registered `prunable` entry가 있으면, 자동 생성/복구를 하지 않고 명시적으로 실패시킨다.
- 이 경우 사용자는 먼저 `wt prune --apply`로 stale registered entry를 정리해야 한다.

## Hosting verify scope
호스팅(PR/MR) merged 여부는 로컬 git 검증과 분리된 opt-in 기능으로 다룬다.

현재 in-scope:
- `wt list --verify-hosting`
- `--verify`와 독립적으로 동작하는 호스팅 전용 검증 필드
- provider 자동 감지(`origin` remote URL 기준)
- GitHub(`gh`) / GitLab(`glab`) 실제 조회 지원
- `gh` 바이너리는 `WT_GH_BIN`, `PATH` 순서로만 탐색
- `glab` 바이너리는 `WT_GLAB_BIN`, `PATH` 순서로만 탐색
- 실패 시 hard error 대신 결과를 `null` + reason으로 반환
- merged 확인 성공 시 change metadata(number/title/url) 반환

현재 out-of-scope:
- 자동 `gh auth login` / `glab auth login`
- 자동 브라우저 인증
- 네트워크 fetch로 remote 상태를 새로 동기화하는 동작

의도:
- squash merge 환경에서 로컬 git `[merged]`와 GitHub PR / GitLab MR merged 여부가 다를 수 있으므로, 의미를 분리해 사용자에게 명확히 보여준다.
- 텍스트 출력 마커는 provider 일반형(`[merged-hosting:<provider>]`)을 사용하고, 상세 의미는 JSON(`hostingProvider`, `hostingKind`, `hostingChangeNumber`, `hostingChangeTitle`, `hostingChangeUrl`)에 둔다.

## List recommendation signals
`wt list`는 사용자가 prune/remove 검토 대상을 빠르게 좁힐 수 있도록 파생 신호를 계산한다.

규칙:
- `stale=true`:
  - `prunable=true`
  - 또는 worktree `path`가 없거나 `.git`가 없다
- `recommendedAction=prune`:
  - `prunable=true`인 경우만 사용한다
- `recommendedAction=remove`:
  - `prunable=false`
  - `current=false`
  - `primary=false`
  - `detached=false`
  - `locked=false`
  - `missing-path=false`
  - `missing-git=false`
  - 그리고 로컬 git merged 또는 hosting merged 중 하나가 확인된 경우
- `recommendedAction=none`:
  - 위 두 조건에 해당하지 않는 경우
- `safeToRemove=true`:
  - `recommendedAction=remove`와 동일한 안전 기준을 만족하는 경우만 사용한다

예외 처리 의도:
- `current`, `primary`는 자동 추천에서 제외한다.
- `detached`는 연결된 branch 의미가 없으므로 remove 추천에서 제외한다.
- `missing-path`/`missing-git`는 stale 신호로는 보이되, `prune` 가능한 상태(`prunable`)가 아니면 자동으로 `remove` 추천하지 않는다.
- squash merge 때문에 `mergedIntoBase`와 `mergedViaHosting`가 다를 수 있으므로, `safeToRemove`/`recommendedAction=remove`는 둘 중 하나가 확인되면 허용하되 근거 필드는 분리해서 유지한다.

## Prune safety
- `wt prune`는 stale/prunable entry 정리 전용이다.
- 기본 동작은 preview-only 이어야 하며, 실제 변경은 명시적 opt-in(`--apply`)일 때만 수행한다.
- `wt prune --tui`는 preview-only 철학을 유지한 interactive preview 계층일 뿐이며, 실제 변경 정책을 바꾸지 않는다.
- `wt prune --tui --apply`도 preview 후 confirm을 거쳐야 하고, 실제 prune 대상은 계속 `git worktree list`의 `prunable` entry로 제한한다.
- 실제 prune은 `git worktree prune --expire now`로 제한하고, 정상 worktree 디렉토리를 직접 삭제하지 않는다.
- non-TTY에서는 `--tui`를 허용하지 않는다.

## Remove safety
- `wt remove`는 정상 worktree를 의도적으로 제거하는 기능이다.
- `wt remove --tui`는 선택 UX만 바꾸며, remove safety를 완화하지 않는다.
- `--dry-run`은 항상 preview-only 이다.
- `--force`는 확인 없이 즉시 제거하는 명시적 opt-in 이다.
- interactive TTY에서만 확인 프롬프트 기반 제거를 허용하고, non-interactive 환경에서는 `--dry-run` 또는 `--force`를 요구한다.
- `--tui`는 `stdin`과 `stderr` 모두 TTY일 때만 허용한다.
- `query`가 없거나 다중 후보일 때만 picker를 사용하고, 선택 후에도 기존 safety 검사(current/primary/prunable)를 다시 적용한다.
- primary worktree와 현재 실행 중인 worktree는 제거할 수 없다.
- `prunable` entry는 `wt remove`가 아니라 `wt prune`로 정리한다.

## Cleanup safety
- `wt cleanup`는 `wt list`의 실행형 companion 명령이다.
- 기본 동작은 항상 preview-only 이다.
- 실제 변경은 `--apply`일 때만 수행한다.
- `recommendedAction=prune`는 `wt prune`와 같은 prune 정책으로만 처리한다.
- `recommendedAction=remove`는 `safeToRemove=true`인 항목에만 적용하고, current/primary/detached/locked/missing-path/missing-git/prunable 예외는 그대로 유지한다.
- squash merge 때문에 `mergedIntoBase`와 `mergedViaHosting`가 다를 수 있으므로, remove 추천/출력에서는 근거를 분리해 유지한다.
- `wt list` 자체에는 destructive 동작을 추가하지 않는다.
