# Release notes

README에는 변경 이력을 중복해서 적지 않고, 사용자-facing 변경은 이 문서에만 기록한다.

## Unreleased

- 2026-03-05: 팀 표준 재사용 프롬프트 템플릿 추가(`docs/prompts/feature-pr-template.md`). PR 가이드에서 해당 템플릿을 기본 규격으로 연결해 에이전트 작업 시 구현/테스트/문서 동기화 체크가 일관되게 적용되도록 정리했다.
- 2026-03-05: 문서를 `main` 기준으로 재정렬했다. PR `#1`~`#23`의 실제 반영 내용을 다시 확인한 뒤 `README`, `CLI spec`, `policy`, `shell`, `TUI`, `roadmap`, `PR guide`를 현재 구현과 맞추고 역할 경계를 정리했다.
- 2026-03-05: `wt create` / `wt path --create`에 최종 생성 경로 preflight를 추가했다. 기존 파일, 비어있지 않은 디렉터리, symbolic link를 포함한 기타 타입은 usage error(exit code 2)로 즉시 실패하고, 경로가 없거나 빈 디렉터리면 기존 생성 흐름을 유지한다. `--dry-run`도 동일 preflight를 수행한다.
- 2026-03-05: `wt prune --tui` 추가. prunable entry를 TUI preview로 확인할 수 있고, `--apply`와 함께 쓰면 preview 뒤 confirm을 거쳐 `git worktree prune --expire now`를 실행한다. 취소는 exit code `130`, `--json`과는 함께 쓸 수 없다.
- 2026-03-05: `wt remove --tui` 추가. query 생략 시 전체 registered worktree에서 고르고, 다중 후보 query는 TUI로 확정할 수 있다. 선택 뒤에도 current/primary/prunable safety는 그대로 유지한다.
- 2026-03-05: `wt path --tui` 추가. query 없이 전체 worktree를 고르거나, 다중 매칭 후보를 TUI로 고를 수 있다. TUI 화면은 `stderr`, 최종 path는 `stdout`에 유지한다.
- 2026-03-05: `wt path --create`와 `wt create`의 생성 안전 규칙을 정리했다. 일반 `wt path`는 registered path를 그대로 반환하고, `--create` 계열은 registered `prunable` entry가 남아 있으면 자동 복구 대신 `wt prune --apply`를 안내하며 실패한다.
- 2026-03-05: `wt cleanup` 추가. `wt list`의 `recommendedAction`을 실제 `prune`/`remove` 실행과 연결하며 기본값은 preview-only 다.
- 2026-03-05: linked worktree 안에서 `wt root`가 primary repo root를 반환하도록 수정했다.
- 2026-03-05: `wt remove` 추가. interactive TTY에서는 confirm prompt를 사용하고, non-interactive 환경에서는 `--dry-run` 또는 `--force`를 요구한다.
- 2026-03-05: `wt root` 추가. 기본 모드는 path-only, `--json`은 `{root}`를 출력한다.
- 2026-03-05: `wt init <shell>` 출력에 `wtr`, `wcd`를 추가하고, zsh completion bridge가 `wtg`와 `wcd` 모두를 `wt path` completion에 연결하도록 정리했다.
- 2026-03-05: `wt list`에 `stale`, `recommendedAction`, `safeToRemove`, `current`, `primary` 파생 신호를 추가했다.
- 2026-03-05: `wt prune` 추가. 기본값은 preview-only 이고, `--apply`일 때만 stale/prunable entry 정리를 수행한다.
- 2026-03-05: `wt goto`를 제거하고 `wt path`를 정식 경로 선택 명령으로 고정했다.
- 2026-03-05: `wt list --verify-hosting`가 GitLab `glab` 기반 merged MR 조회를 지원하도록 확장했다.
- 2026-03-05: `wt run <query> -- <cmd...>` 추가. `wt path`와 같은 매칭 규칙으로 worktree를 고른 뒤 해당 디렉터리에서 명령을 실행한다.
- 2026-03-05: `wt create`와 `wt path --create`가 공통 root override 정책을 사용하도록 정리했다. 우선순위는 `--root` > `WT_ROOT` > repo-local `wt.root` > default root 다.
- 2026-03-05: `wt list --json --verify` 출력 스키마를 고정했다. `pathExists`, `dotGitExists`, `valid`, `mergedIntoBase`, `baseRef`를 안정적으로 포함하고, detached 항목은 `mergedIntoBase: null`을 사용한다.
- 2026-03-04: `wt list` 구현. `--json`, `--porcelain`, `--verify`, `--base` 지원.
- 2026-03-04: `wt path` 구현. 기본 출력은 path-only, `--json` 지원.
- 2026-03-04: `wt path <query>` 동적 completion 구현. 기본값은 현재 registered worktree 후보, `WT_PATH_COMPLETE_REMOTE=1`이면 원격 브랜치 후보를 추가한다.
- 2026-03-04: `wt init <shell>` 구현. 출력-only 방식으로 셸 helper를 제공한다.
- 2026-03-04: `wt create <branch>` 구현과 `wt path --create` 지원 추가.
- 2026-03-04: Cobra 기반 CLI 구조와 `wt --version` 지원 추가.
- 2026-03-04: `./scripts/install.sh` 개선. 버전 출력과 명시적 overwrite(`--force`)를 지원하고, completion/TUI 설치는 자동으로 하지 않는다.
- 2026-03-04: `make premerge` 게이트 추가. `make check`와 테스트를 머지 전 기본 검증으로 사용한다.
- 2026-03-04: 문서 레이아웃을 `docs/spec`, `docs/policy`, `docs/ux`, `docs/release`, `docs/roadmap` 구조로 정리했다.
