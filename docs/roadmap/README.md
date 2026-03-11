# Roadmap

이 문서는 현재 구현 상태를 기준으로, 이미 완료된 범위와 다음 현실적 작업 순서를 구분해 기록한다.

## Already Shipped

- Core listing: `wt list`, `--json`, `--porcelain`, `--verify`, `--verify-hosting`
- Derived-signal filters: `wt list --stale`, `--safe-to-remove`, `--recommended <none|prune|remove>`
- Path selection: `wt path`, `--json`, `--create`, `--tui`, `--no-tui`
- Root and execution: `wt root`, `wt run`
- Worktree lifecycle: `wt create`, `wt remove`, `wt prune`, `wt cleanup`
- Environment diagnostics: `wt doctor`, text/JSON checks for Git context, root config, hosting CLI, shell setup
- Shell integration: `wt init <shell>`, `wtg`, `wcd`, `wtr`, `wt completion <shell>`, init/completion install guidance
- TUI flows: reusable picker core, `wt path --tui`, `wt remove --tui`, `wt prune --tui`, `wt cleanup --tui`
- Hosting integration: GitHub PR / GitLab MR merged verification
- Install/update flow: `scripts/install.sh`, `wt upgrade`, latest release resolution
- Agent workflow docs: reusable prompt template, skill registration guide, roadmap/backlog docs

## Next Likely Work

현재 구조와 최근 머지 흐름을 기준으로 보면 다음 순서가 가장 자연스럽다.

1. Structured output consistency hardening
핵심 action/applied/removed semantics 는 이미 정리됐다. 남은 범위는 `list`/`cleanup` 사이 verify field 범위 차이, deprecated alias 정리 시점 같은 스크립트 소비자 관점의 마감 정리에 가깝다.

2. `wt doctor` follow-up polish
새로 들어간 `doctor`는 진단 범위가 넓어서 실제 사용 결과를 보고 check naming, warning copy, shell/completion 판별 신호, exit semantics를 조금 더 다듬을 여지가 있다. 새 명령을 더 추가하기보다 현재 진단 결과를 신뢰하기 쉽게 만드는 쪽이 먼저다.

3. Agent/shell UX follow-up hardening
최근에 `cleanup --tui`, list 필터, shell/completion 설치 가이드, prompt template까지 한 번에 들어갔다. 이제는 새 기능을 더 붙이기보다 실제 사용 결과를 보고 도움말, 문서 밀도, helper 범위를 줄이거나 다듬는 후속 정리가 더 현실적이다.

구체 후보와 우선순위 메모는 [docs/roadmap/backlog.md](backlog.md)에 분리해 둔다.

## Not Current Scope

- Git 자체를 대체하는 대규모 UI
- 자동 로그인, 자동 fetch, 자동 브라우저 인증
- 사용자의 rc 파일이나 로컬 환경을 기본값으로 자동 수정하는 설치 방식
- 기본값이 파괴적인 remove/prune/reset 류 동작
