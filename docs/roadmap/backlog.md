# Feature backlog

이 문서는 현재 구현과 안전 규칙을 기준으로, 다음 개발 논의에 바로 사용할 수 있는 후보 피처를 정리한다.

## Prioritized candidates

1. `wt cleanup`/`wt list` 후속 polish
- 현재 `main`에는 `cleanup --tui`, `list` 파생 신호 필터(`--stale`, `--safe-to-remove`, `--recommended`)가 이미 들어가 있다.
- 다음 단계는 기능 추가가 아니라 naming/help/output contract 를 안정화하는 작은 단위 작업으로 진행한다.

### Work packages (implementation-sized)

1. Package A: help/용어 정합 정리 (진입 작업)
- 목표: `wt list`/`wt cleanup`의 옵션 설명과 roadmap/backlog 용어를 동일하게 맞춘다.
- 산출물: help 문구 또는 스펙 문구의 최소 수정, 용어 테이블(추천 신호/preview/apply/review) 일치.
- 의존관계: 없음.

2. Package B: `wt cleanup --tui` review/apply 문구 고정
- 목표: review help, continue row, confirm/abort/cancel 메시지의 사용자 계약을 명확히 고정한다.
- 산출물: 텍스트 계약 문서화 + 기존 동작 회귀 테스트 정리(`review cancelled` exit 130, confirm 거부 시 `wt cleanup: aborted` 유지).
- 의존관계: Package A 이후(용어 기준선 재사용).

3. Package C: `wt list` 필터 discoverability 강화
- 목표: `--stale`/`--safe-to-remove`/`--recommended` 조합 규칙과 verify 의존성을 사용자가 즉시 이해하게 만든다.
- 산출물: help/spec 예시 보강, 텍스트/JSON 필터 semantics 동일성 근거 연결.
- 의존관계: Package A 이후.

4. Package D: `list --json`/`cleanup --json` verify 필드 범위 마감
- 목표: 두 명령의 verify 필드 범위를 계약 단위로 비교해 남은 차이를 정리하고 deprecated alias 정리 시점을 확정한다.
- 산출물: 필드 매트릭스(포함/조건/호환성), 필요한 테스트/스펙 보강, 하위호환 일정 메모.
- 의존관계: Package B/C 이후(메시지/필터 용어가 고정된 상태에서 스키마 마감).

## Audit notes

- 기존 backlog 1/2/3 성격(출력 정합, `doctor` 후속, agent/shell 후속)은 현재 `main`에 머지된 범위를 기준으로 우선순위 큐에서 내리고 shipped/follow-up 관점으로 관리한다.
- `wt upgrade`는 이미 구현 및 문서화되어 backlog 후보가 아니라 shipped 범위로 본다.
- 현재 우선순위는 `cleanup`/`list` polish 를 작은 작업 묶음으로 순차 마감하는 것이다.

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
