# Feature backlog

이 문서는 현재 구현과 안전 규칙을 기준으로, 다음 개발 논의에 바로 사용할 수 있는 후보 피처를 정리한다.

## Prioritized candidates

1. `wt cleanup --tui` 또는 선택형 apply
- 현재 `wt cleanup`는 추천 액션 계산과 실제 실행 엔진은 있지만, 결과를 사람이 고르는 단계가 없다.
- 기존 picker/TUI와 `recommendedAction`, `safeToRemove` 신호를 재사용할 수 있어서 구현 비용 대비 효과가 크다.
- 기본값은 계속 preview-only 로 유지하고, apply 는 명시적 opt-in 이어야 한다.

2. `wt list` 필터 옵션
- 예: `--stale`, `--safe-to-remove`, `--recommended remove|prune`
- 현재도 `cmd/wt/list.go`에서 같은 신호를 계산하므로 새 검증 로직 없이 출력 계층 확장으로 접근할 수 있다.
- 스크립트와 사람이 보는 출력 모두에 직접 이득이 있다.

3. `wt doctor`
- 점검 대상 후보: Git context, primary root 해석, `wt.root`, `WT_ROOT`, `gh`/`glab`, shell completion 설치 여부
- `--verify-hosting`와 shell integration 쪽 문제를 진단할 때 반복되는 수동 확인을 줄일 수 있다.
- 기본 출력은 사람이 읽기 좋게 두고, 필요하면 `--json`을 제공하는 방향이 맞다.

4. Shell/completion 설치 helper
- 자동 rc 수정은 현재 non-goal 과 충돌하므로 피한다.
- 대신 completion 파일 생성 위치 안내, snippet export, 설치 dry-run 같은 안전한 helper 는 범위 안에 있다.
- `wt init`과 Cobra completion 을 묶는 좁은 UX 정리가 적절하다.

5. Structured JSON consistency
- `action`, `reason`, `removed`, exit code 표현이 명령별로 조금씩 다르다.
- 스크립트 소비자를 고려하면 명령 간 schema naming 과 preview/apply 상태 표현을 맞추는 후속 작업이 가치가 있다.

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
