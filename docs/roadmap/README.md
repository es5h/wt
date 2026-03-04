# Roadmap

이 문서는 현재 구현 상태를 기준으로, 이미 완료된 범위와 다음 현실적 작업 순서를 구분해 기록한다.

## Already Shipped

- Core listing: `wt list`, `--json`, `--porcelain`, `--verify`, `--verify-hosting`
- Path selection: `wt path`, `--json`, `--create`, `--tui`, `--no-tui`
- Root and execution: `wt root`, `wt run`
- Worktree lifecycle: `wt create`, `wt remove`, `wt prune`, `wt cleanup`
- Shell integration: `wt init <shell>`, `wtg`, `wcd`, `wtr`, `wt completion <shell>`
- TUI flows: reusable picker core, `wt path --tui`, `wt remove --tui`, `wt prune --tui`
- Hosting integration: GitHub PR / GitLab MR merged verification

## Next Likely Work

현재 구조와 최근 머지 흐름을 기준으로 보면 다음 순서가 가장 자연스럽다.

1. Shell/completion 설치 UX 정리
현재는 `wt completion <shell>`과 `wt init <shell>`이 모두 존재하지만 설치는 전부 수동이다. 설치 스크립트와 문서에서 opt-in 설치 경로를 더 분명히 하거나, 안전한 범위의 helper 명령을 추가하는 작업이 다음 단계로 적합하다.

2. Cleanup selection ergonomics
`wt cleanup`는 이미 추천 신호와 실행 엔진을 갖고 있지만 현재는 일괄 preview/apply 중심이다. 지금 있는 `list` 파생 신호와 TUI picker를 재사용해 선택적 review/apply 흐름을 붙이는 것이 현실적인 확장이다.

3. Structured output consistency hardening
`list`, `path`, `run`, `remove`, `prune`, `cleanup`에 JSON이 이미 존재한다. 스크립트 사용성을 높이려면 명령 간 action/reason/exit code 표현을 더 일관되게 다듬는 후속 작업이 자연스럽다.

## Not Current Scope

- Git 자체를 대체하는 대규모 UI
- 자동 로그인, 자동 fetch, 자동 브라우저 인증
- 사용자의 rc 파일이나 로컬 환경을 기본값으로 자동 수정하는 설치 방식
- 기본값이 파괴적인 remove/prune/reset 류 동작
