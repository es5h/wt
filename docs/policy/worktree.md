# Policy (draft)

이 문서는 `wt` 내부 정책(경로/정규화 등)을 “구현 근거가 되는 규칙” 형태로 기록합니다.

## Git context
- “Git 컨텍스트(repo root)”는 현재 디렉토리에서 `git rev-parse --show-toplevel`로 결정한다.

## Default worktree root
- 기본 worktree 루트 경로 정책은 한 가지로 고정하고 문서화한다.
  - 예: repo 상위 `../.wt/<repo>/...`

## Overrides (opt-in)
기본 정책은 재현성을 위해 고정하되, 사용자/팀 환경에 맞게 “명시적으로” 오버라이드할 수 있게 한다.

우선순위(높음 → 낮음):
1) CLI flag (예: `--root`, `--layout`) — 가장 명시적이며 스크립트에 안전
2) Environment variables — CI/개인 환경에 유용
3) Repo-local git config (`git config --local ...`) — repo 단위로 팀/에이전트가 따라가기 쉬움
4) Default policy

권장 environment variables(초안):
- `WT_ROOT`: worktree 루트 디렉토리(절대/상대 경로)
- `WT_LAYOUT`: 레이아웃 프리셋(예: `dotwt`, `flat` 등 — 명칭은 추후 확정)
- `WT_NORMALIZE`: 브랜치→폴더명 정규화 규칙 프리셋(추후 확정)

권장 repo-local git config(초안):
- `git config --local wt.root ../.wt`
- `git config --local wt.layout dotwt`

주의:
- 자동으로 사용자 홈/셸 설정을 변경하는 동작은 기본값으로 두지 않는다.
- 오버라이드는 “없는 경우에만 적용”이 아니라, 명시적 설정이 있는 경우에만 적용하는 것을 원칙으로 한다.

## Branch name → directory name normalization
- 브랜치명 → 폴더명 정규화 규칙을 유틸로 통일한다.
  - 예: `/` → `-`
  - 공백/특수문자 처리 규칙
  - 충돌 시 suffix 부여 규칙
