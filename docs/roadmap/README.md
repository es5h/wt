# Roadmap (draft)

이 문서는 `wt`의 “미구현 기능/아이디어/우선순위”를 기록합니다. 현재 스펙/정책은 아래 문서를 우선합니다.
- 사용자 스펙: `docs/spec/cli.md`
- 정책: `docs/policy/worktree.md`
- 개발/에이전트 런북: `AGENTS.md`

## Milestones (idea)
- M0 Core: `list`, `path` 기본 UX + 안정적 출력 포맷
- M1 Create/Remove: `path --create`, `create`, `remove`, `prune` (안전장치 포함)
- M2 Shell: `init <shell>` + completion 스크립트
- M3 TUI: `path` TUI picker

버전(semver)은 로드맵이 아니라 `VERSION` + `docs/release/notes.md`에서만 관리한다.

## Open questions
- completion 설치 UX를 `wt init`에 포함할지, 별도 `wt completion <shell>`로 분리할지
- `wt path` 다중 후보 시 기본 동작(TUI 자동 vs 에러 vs 프롬프트)
- TUI 취소 시 exit code(130 vs 1) 정책

## Dependencies / implementation choices
- CLI 프레임워크를 쓸 경우 `cobra` 통일(혼용 금지)
- TUI는 초기엔 최소 구현, 필요 시 `bubbletea` 계열 검토
