# Release notes (draft)

이 문서는 사용자에게 보이는 변경사항을 기록합니다. (README에는 패치 노트를 쓰지 않음)

## Unreleased
- 2026-03-05: `wt remove` 추가(기본 `--dry-run` preview-only, `--force`는 즉시 삭제, interactive TTY에서는 확인 프롬프트 지원; current/primary worktree 제거 금지, `prunable` entry는 `wt prune`로 분리)
- 2026-03-05: `wt root` 추가(기본 stdout path-only, `--json` 시 `{root}` 출력)
- 2026-03-05: `wt init <shell>` 출력에 `wtr`와 `wcd`를 추가하고, zsh completion bridge가 `wtg`뿐 아니라 `wcd`도 `wt path`처럼 완성하도록 조정
- 2026-03-05: `wt prune` 추가(기본 preview-only, `--apply`일 때만 `git worktree prune --expire now` 실행, stale/prunable entry 정리용)
- 2026-03-05: breaking: `wt goto`를 제거하고 `wt path`를 정식 경로 선택 명령으로 사용하도록 정리(하위 호환 alias 없음)
- 2026-03-05: `wt list --verify-hosting` 추가(GitHub `gh` 기반 opt-in; `[merged-hosting:<provider>]`/`mergedViaHosting`로 로컬 git `[merged]`와 의미 분리, `hostingChangeNumber`/`hostingChangeTitle`/`hostingChangeUrl` 제공, GitLab은 현재 감지만 지원)
- 2026-03-05: `wt run <query> -- <cmd...>` 추가(`wt path`와 같은 매칭 규칙 사용, 종료 코드 보존, `--json` 지원)
- 2026-03-05: `wt create`와 `wt path --create`가 동일한 worktree root 오버라이드 정책을 공유하도록 리팩터링 (`--root` > `WT_ROOT` > repo-local git config `wt.root` > `<repo>/.wt`)
- 2026-03-05: `wt list --json --verify` 출력 스키마를 고정해 `pathExists`, `dotGitExists`, `valid`, `mergedIntoBase`, `baseRef`를 항상 포함하도록 조정하고, detached/branch 없음 케이스의 `mergedIntoBase`를 `null`로 명시
- 2026-03-04: `wt list` 구현(`--json`, `--porcelain`, `--verify`, `--base` 지원)
- 2026-03-04: `wt path` 구현(`--json` 지원; `--tui`는 미구현)
- 2026-03-04: `wt path <query>`에서 “현재 worktree 브랜치”를 동적으로 자동완성 후보로 제공(셸 completion 설치 시)
- 2026-03-04: `wt init <shell>` 구현(출력-only: rc 자동 수정 없음)
- 2026-03-04: `wt create <branch>` 구현 + `wt path --create`로 “없으면 생성 후 이동” 지원
- 2026-03-04: cobra 기반 CLI 스캐폴딩 추가(`cmd/wt`), `wt --version` 지원(`VERSION` + 빌드 시 ldflags 주입)
- 2026-03-04: `./scripts/install.sh`가 버전 출력 및 “install만으로 update(덮어쓰기)” 동작하도록 개선(자동완성/TUI 자동 설치는 하지 않음)
- 2026-03-04: `make test`/`make build` 실행 전 `make check`(=`gofmt` + `go fix`)를 필수로 적용 + `make premerge` 게이트 도입
- 2026-03-04: 문서를 폴더 구조로 리카테고리(`docs/spec`, `docs/policy`, `docs/ux`, `docs/release`, `docs/roadmap`)
