# CLI spec (draft)

이 문서는 `wt`의 명령/옵션 동작을 “테스트 가능한 규칙” 수준으로 고정하기 위한 초안입니다.

## Global rules
- Git 컨텍스트(repo root)는 현재 디렉토리에서 `git rev-parse --show-toplevel`로 결정한다. (정책: `docs/policy/worktree.md`)
- 정상 출력은 stdout, 에러는 stderr.
- 사람이 읽는 출력과 스크립트 출력이 충돌하면 `--json` 또는 `--porcelain`로 분리한다.
- 경로 정책/오버라이드(환경변수, git config 등)는 `docs/policy/worktree.md`에 정의한다.

## `wt list`
목표: 현재 repo의 worktree를 나열한다.

출력(기본):
- 각 worktree의 `path`, 연결된 `branch`(있으면), `HEAD`(짧은 해시), `locked` 여부를 표시
- 파생 신호가 있으면 짧은 마커로 함께 표시한다.
  - `current`, `primary`: 현재 worktree / primary worktree
  - `missing-path`, `missing-git`, `stale`: 엔트리 이상 징후
  - `merged`: 로컬 git 기준 base ref로 merge됨 (`git merge-base --is-ancestor`)
  - `merged-hosting:<provider>`: hosting(PR/MR) 기준 merge됨
  - `safe-remove`, `recommend:remove|prune`: 검토 추천 신호

옵션(초안):
- `--json`: 구조화 출력
- `--porcelain`: git처럼 안정적 포맷(필드 고정, 파싱 용도)
- `--verify`: worktree entry 검증(경로/.git 존재 + base ref 기준 merged 여부)
  - `--base <ref>`: `--verify`의 base ref 지정(기본: `origin/HEAD` 또는 `main`)
- `--verify-hosting`: 호스팅(PR/MR) 기준 merged 여부를 추가 검증
  - `--verify`와 배타적이지 않다. 둘을 함께 쓰면 로컬 git 검증과 호스팅 검증을 둘 다 표시한다.
  - `--verify-hosting`만 쓰면 호스팅 필드만 추가하고, `pathExists`/`dotGitExists`/`valid`/`mergedIntoBase`는 포함하지 않는다.
  - 현재 범위(in-scope): GitHub(`gh`) / GitLab(`glab`) 지원
  - 바이너리 탐색 순서:
    - GitHub: `WT_GH_BIN` > `PATH`
    - GitLab: `WT_GLAB_BIN` > `PATH`
  - GitHub는 `gh auth status` 후 merged PR(`number/title/url`)를 조회한다.
  - GitLab은 `glab auth status` 후 `glab api projects/:fullpath/merge_requests?...`로 merged MR(`iid/title/web_url`)를 조회한다.
  - 현재 범위(out-of-scope): 자동 로그인, 자동 브라우저 인증, 자동 fetch
  - 실패 정책: `gh`/`glab`가 없거나 인증/조회에 실패해도 명령 전체를 실패시키지 않고 텍스트 출력엔 note, JSON에는 `mergedViaHosting: null` + `hostingReason`으로 표현

현재 구현 상태:
- 로컬 git 기준 `[merged]`는 `git merge-base --is-ancestor <branch> <base>` 의미를 유지한다.
- 호스팅 기준 merged는 별도 필드/마커로 분리한다:
  - 텍스트: `[merged-hosting:<provider>]`
  - JSON: `hostingProvider`, `hostingKind`, `mergedViaHosting`, `hostingReason`, `hostingChangeNumber`, `hostingChangeTitle`, `hostingChangeUrl`
- provider 감지는 `origin` remote URL 기준 자동 감지다.
<<<<<<< HEAD
- GitLab remote는 merged MR이 확인되면 `hostingProvider=gitlab`, `hostingKind=mr`, `hostingChangeNumber`(MR IID), `hostingChangeTitle`, `hostingChangeUrl`를 채운다.
- GitHub/GitLab 조회가 불가능하면 provider/kind는 유지하고 `mergedViaHosting=null`, `hostingReason`으로 degrade 한다.
- 파생 신호는 기본 text/JSON에 항상 포함한다:
  - JSON: `current`, `primary`, `stale`, `recommendedAction`, `safeToRemove`
  - `recommendedAction`은 `prune`, `remove`, `none` 중 하나다.
  - `stale`은 `prunable` 이거나 worktree `path`/`.git`가 누락된 경우 `true`다.
  - `safeToRemove`는 `prunable`, `current`, `primary`, `detached`, `locked`, `missing-path`, `missing-git`가 아닌 항목 중 로컬 git merged 또는 hosting merged가 확인된 경우만 `true`다.
  - `recommendedAction=prune`은 stale/prunable 엔트리 정리 검토를 뜻한다.
  - `recommendedAction=remove`는 linked worktree 디렉토리 제거 검토를 뜻하며, 로컬 git merged와 hosting merged의 의미 차이는 기존 필드(`mergedIntoBase`, `mergedViaHosting`)로 계속 분리한다.
=======
- GitLab remote는 merged MR이 확인되면 `hostingProvider=gitlab`, `hostingKind=mr`, `hostingChangeNumber`(MR IID), `hostingChangeTitle`, `hostingChangeUrl`를 채운다.
- GitHub/GitLab 조회가 불가능하면 provider/kind는 유지하고 `mergedViaHosting=null`, `hostingReason`으로 degrade 한다.
>>>>>>> 4311044 (feat(list): add GitLab hosting verify support)

`--json --verify` 출력 규칙:
- 각 항목은 항상 `pathExists`, `dotGitExists`, `valid`, `mergedIntoBase`, `baseRef` 필드를 포함한다.
- `mergedIntoBase`는 boolean 또는 `null`이다.
- `mergedIntoBase: null`은 merged 여부를 계산할 브랜치 ref가 없는 경우에 사용한다.
  - 예: detached worktree
  - 예: branch 정보가 없는 entry
- `baseRef`는 `--verify`가 켜진 JSON 출력에서 항상 문자열로 포함된다.
- `current`, `primary`, `stale`, `recommendedAction`, `safeToRemove`는 `--verify` 여부와 무관하게 항상 포함된다.

## `wt path [query]`
목표: query로 worktree를 선택하고 “경로”를 stdout으로 출력한다.

규칙:
- 기본 모드는 **경로만 출력**한다(추가 텍스트/색상 금지).
- 후보가 0개면 non-zero exit + stderr에 후보/가이드 출력.
- 후보가 1개면 자동 선택.
- 후보가 2개 이상이면:
  - 기본 동작은 정책 확정 전(로드맵: `docs/roadmap/README.md`)
  - `--tui`가 있으면 TUI로 선택(스펙: `docs/ux/tui.md`)
  - `--no-tui`가 있으면 에러(스크립트 안전)

옵션(초안):
- `--create`: query에 해당하는 브랜치 worktree가 없으면 생성
- `--path <dir>`: `--create` 시 생성 경로를 직접 지정
- `--root <dir>`: `--create` 시 기본 생성 경로의 root 지정. 우선순위는 `--root` > `WT_ROOT` > repo-local git config `wt.root` > `<primary-root>/.wt`
- `--from <ref>`: `--create` 시 새 브랜치의 start point 지정(기본: `origin/<branch>`가 있으면 그걸 사용, 없으면 `origin/HEAD` 또는 `main`)
- `--tui`: 후보가 여러 개면 TUI로 선택(비-interactive 환경이면 에러)
- `--no-tui`: 후보가 여러 개면 에러(스크립트 안전)
- `--json`: 선택 결과를 json으로 출력(예: `{path, branch, reason}`)

현재 구현 상태:
- `--tui`는 아직 미구현이며 지정 시 사용법 에러로 종료한다.
- 후보가 2개 이상이면 기본 동작은 “에러 + 후보 출력”이다(TUI 구현 전까지).

TUI 동작/키바인딩 상세는 `docs/ux/tui.md` 참고.

## `wt root`
목표: 현재 Git 컨텍스트의 repo root 경로를 stdout으로 출력한다.

규칙:
- 기본 모드는 **경로만 출력**한다(추가 텍스트/색상 금지).
- 출력은 `git rev-parse --show-toplevel` 기준 repo root path다.

옵션:
- `--json`: `{root}` 출력

## `wt run <query> -- <cmd...>`
목표: `wt path`와 같은 매칭 규칙으로 worktree를 선택한 뒤, 그 디렉토리에서 `<cmd...>`를 실행한다.

규칙:
- `<query>` 매칭/모호성 처리/에러 코드는 `wt path`와 동일하다.
- 기본 모드는 하위 프로세스의 stdout/stderr를 그대로 전달하고, 하위 프로세스의 종료 코드를 그대로 반환한다.
- `--json`은 stdout에 선택된 worktree와 실행 결과를 JSON으로 출력한다.

옵션:
- `--json`: `{path, command, exitCode}` 출력

현재 구현 규칙:
- `command`는 JSON에서 argv 배열로 출력한다.
- `--json` 사용 시 stdout은 JSON 전용이며, 하위 프로세스의 stdout/stderr는 중계하지 않는다.

## `wt create <branch>`
목표: `<branch>`에 대한 worktree를 생성한다.

옵션(초안):
- `--path <dir>`: 생성 경로를 직접 지정
- `--root <dir>`: 기본 생성 경로의 root 지정. 우선순위는 `--root` > `WT_ROOT` > repo-local git config `wt.root` > `<primary-root>/.wt`
- `--from <ref>`: 새 브랜치의 start point 지정(기본: `origin/HEAD` 또는 `main`)
- `--dry-run`: 실행될 git 커맨드/경로만 출력(변경 없음)

현재 구현 규칙:
- 기본 생성 경로: `<primary-root>/.wt/<branch>`
- `--root`, `WT_ROOT`, `wt.root`가 상대 경로이면 `<primary-root>` 기준으로 해석한다.
- 로컬 브랜치가 이미 존재하면: `git worktree add <path> <branch>`
- 로컬 브랜치가 없고 `origin/<branch>`가 존재하면: `git worktree add -b <branch> <path> origin/<branch>`
- 둘 다 없으면: `git worktree add -b <branch> <path> <from>`

## `wt remove <name>`
목표: worktree를 제거한다.

규칙:
- `--dry-run`이면 preview-only 이고 실제 변경을 하지 않는다.
- `--force`이면 확인 없이 즉시 제거한다.
- `--dry-run`/`--force`가 둘 다 없으면 interactive TTY에서만 확인 프롬프트를 보여준다.
- non-interactive 환경에서는 기존처럼 `--dry-run` 또는 `--force`가 필요하다.
- primary worktree는 제거할 수 없다.
- 현재 실행 중인 worktree는 제거할 수 없다.
- `prunable` entry는 `remove` 대상이 아니며, `wt prune --apply`를 사용해야 한다.
- 브랜치 삭제는 기본 동작이 아니다(별도 옵션 없음).

옵션:
- `--force`: 확인 생략
- `--dry-run`
- `--json`: `{path, branch, action, removed}` 출력

현재 구현 규칙:
- interactive TTY에서는 stderr에 `Remove worktree <path> (<branch>)? [y/N]` 프롬프트를 출력하고, `y`/`yes`일 때만 삭제한다.
- 실제 삭제는 내부적으로 `git worktree remove --force <path>`를 사용한다.
- 기본 text 출력은 `would-remove` / `removed` 상태를 한 줄로 보여준다.

## `wt prune`
목표: stale/prunable worktree entry를 preview하거나 정리한다.

규칙:
- 기본 동작은 preview-only 이다. 실제 변경은 하지 않는다.
- `git worktree list`에서 `prunable`로 표시되는 entry만 대상이다.
- `--apply`가 있을 때만 `git worktree prune --expire now`를 실행한다.
- primary worktree나 정상 worktree는 직접 제거하지 않는다.

옵션:
- `--apply`: 실제 prune 실행
- `--json`: `{path, branch, pruneReason, action, removed}` 배열 출력

현재 구현 규칙:
- 기본 text 출력은 `would-prune` / `pruned` / `kept` 상태를 한 줄씩 보여준다.
- `--apply` 후에는 prune 전 후보 목록을 기준으로, prune 후 다시 조회해 `removed` 여부를 계산한다.

## `wt init <shell>`
목표: `path`가 `cd`될 수 있도록 셸 함수/alias를 출력한다.

지원 셸(초안): `zsh`, `bash`, `fish`

예시(개념):
- `wt init zsh` → stdout에 function 정의를 출력(사용자가 rc에 추가)

셸 통합/완성 관련 상세는 `docs/ux/shell.md` 참고.

규칙:
- `wt init`은 사용자의 rc 파일을 자동으로 수정하지 않는다(출력-only).
- 현재 출력에는 다음 helper가 포함된다:
  - `wtr`: `cd "$(wt root)"` 래퍼
  - `wtg`: `cd "$(wt path ...)"` 래퍼
  - `wcd`: `wtg`와 동일하게 `wt path`를 이용해 선택된 worktree로 이동하는 별칭용 래퍼
