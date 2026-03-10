# Roadmap

이 문서는 현재 구현 상태를 기준으로, 이미 완료된 범위와 다음 현실적 작업 순서를 구분해 기록한다.

## Already Shipped

- Core listing: `wt list`, `--json`, `--porcelain`, `--verify`, `--verify-hosting`
- Derived-signal filters: `wt list --stale`, `--safe-to-remove`, `--recommended <none|prune|remove>`
- Path selection: `wt path`, `--json`, `--create`, `--tui`, `--no-tui`
- Root and execution: `wt root`, `wt run`
- Worktree lifecycle: `wt create`, `wt remove`, `wt prune`, `wt cleanup`
- Shell integration: `wt init <shell>`, `wtg`, `wcd`, `wtr`, `wt completion <shell>`, init/completion install guidance
- TUI flows: reusable picker core, `wt path --tui`, `wt remove --tui`, `wt prune --tui`, `wt cleanup --tui`
- Hosting integration: GitHub PR / GitLab MR merged verification
- Agent workflow docs: reusable prompt template, skill registration guide, roadmap/backlog docs

## Next Likely Work

현재 구조와 최근 머지 흐름을 기준으로 보면 다음 순서가 가장 자연스럽다.

1. `wt doctor`
hosting verify, shell completion, `wt.root`, `WT_ROOT`, `gh`/`glab` 같은 환경 의존성이 점점 늘고 있다. 설치 상태와 현재 repo 컨텍스트를 빠르게 점검하는 진단 명령이 있으면 팀 온보딩과 에이전트 환경 점검이 쉬워진다.

2. Structured output consistency hardening
`list`, `path`, `run`, `remove`, `prune`, `cleanup`에 JSON이 이미 존재한다. 스크립트 사용성을 높이려면 명령 간 action/reason/exit code 표현을 더 일관되게 다듬는 후속 작업이 자연스럽다.

3. Agent/shell UX follow-up hardening
최근에 `cleanup --tui`, list 필터, shell/completion 설치 가이드, prompt template까지 한 번에 들어갔다. 이제는 새 기능을 더 붙이기보다 실제 사용 결과를 보고 도움말, 문서 밀도, helper 범위를 줄이거나 다듬는 후속 정리가 더 현실적이다.

구체 후보와 우선순위 메모는 [docs/roadmap/backlog.md](backlog.md)에 분리해 둔다.

## Not Current Scope

- Git 자체를 대체하는 대규모 UI
- 자동 로그인, 자동 fetch, 자동 브라우저 인증
- 사용자의 rc 파일이나 로컬 환경을 기본값으로 자동 수정하는 설치 방식
- 기본값이 파괴적인 remove/prune/reset 류 동작
