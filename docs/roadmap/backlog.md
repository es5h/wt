# Feature backlog

이 문서는 현재 구현과 안전 규칙을 기준으로, 다음 개발 논의에 바로 사용할 수 있는 후보 피처를 정리한다.

## Prioritized candidates

1. Structured JSON consistency
- 핵심 `action`/`applied`/`removed` semantics 는 이미 정리됐다.
- 남은 범위는 `list --json`과 `cleanup --json`의 verify field 범위 차이, `hostingChangeUrl` 같은 세부 key naming 정리처럼 스크립트 소비자를 위한 마감 작업에 가깝다.

2. `wt doctor` follow-up polish
- `doctor`는 이제 `main`에 들어왔지만, 실제 사용 결과를 기준으로 check naming, warning copy, shell/completion 판별 방식, JSON/text parity 를 다듬을 여지가 있다.
- 명령 추가보다 진단 결과 신뢰도와 문제 재현성을 높이는 후속 정리가 먼저일 가능성이 크다.

3. Agent/shell UX follow-up hardening
- 최근 머지로 shell/completion 설치 가이드와 에이전트용 prompt template 이 들어갔다.
- 실제 사용 결과를 보고 문서 길이, example 밀도, helper 범위, upgrade/install 안내 문구를 줄이거나 보정하는 후속 정리가 필요할 수 있다.
- 새 기능 추가보다는 실제 사용 흔적을 기준으로 다듬는 성격이 강하다.

4. `wt cleanup`/`wt list` 후속 polish
- `cleanup --tui`와 list 필터는 이미 들어갔지만, 실제 사용 후 review/apply copy, filter discoverability, help text 는 더 다듬을 여지가 있다.
- 기능 확장보다 naming/help/output polish 위주로 접근하는 편이 맞다.

## Audit notes

- `wt doctor`는 `main`에 구현 및 문서화되었으므로 신규 feature backlog 에서 제외한다.
- `wt upgrade`는 이미 구현 및 문서화되어 backlog 후보가 아니라 shipped 범위로 본다.
- 현재 기능 표면 기준으로는 새 대형 명령을 더 늘리기보다 좁아진 출력 스키마 정리와 `doctor` 후속 polish 가 우선이다.

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
