# Feature backlog

이 문서는 현재 구현과 안전 규칙을 기준으로, 다음 개발 논의에 바로 사용할 수 있는 후보 피처를 정리한다.

## Prioritized candidates

1. `wt doctor`
- 점검 대상 후보: Git context, primary root 해석, `wt.root`, `WT_ROOT`, `gh`/`glab`, shell completion 설치 여부
- `--verify-hosting`와 shell integration 쪽 문제를 진단할 때 반복되는 수동 확인을 줄일 수 있다.
- 기본 출력은 사람이 읽기 좋게 두고, 필요하면 `--json`을 제공하는 방향이 맞다.

2. Structured JSON consistency
- `action`, `reason`, `removed`, exit code 표현이 명령별로 조금씩 다르다.
- 스크립트 소비자를 고려하면 명령 간 schema naming 과 preview/apply 상태 표현을 맞추는 후속 작업이 가치가 있다.

3. Agent/shell UX follow-up hardening
- 최근 머지로 shell/completion 설치 가이드와 에이전트용 prompt template 이 들어갔다.
- 실제 사용 결과를 보고 문서 길이, example 밀도, helper 범위를 줄이거나 보정하는 후속 정리가 필요할 수 있다.
- 새 기능 추가보다는 실제 사용 흔적을 기준으로 다듬는 성격이 강하다.

4. `wt cleanup`/`wt list` 후속 polish
- `cleanup --tui`와 list 필터는 이미 들어갔지만, 실제 사용 후 review/apply copy, filter discoverability, help text 는 더 다듬을 여지가 있다.
- 기능 확장보다 naming/help/output polish 위주로 접근하는 편이 맞다.

## Candidate selection criteria

- 기존 safety rule 을 약화시키지 않을 것
- 이미 있는 신호나 helper 를 재사용할 수 있을 것
- path-only / stdout-stderr 계약을 깨지 않을 것
- non-interactive 자동화와 interactive UX 둘 다 이득이 있을 것

## Deferred ideas

- 대규모 상시 TUI 모드
- Git operation orchestration 전반을 감싸는 UI
- 기본값이 파괴적인 자동 정리
- 자동 인증, 자동 fetch, 자동 브라우저 연동
