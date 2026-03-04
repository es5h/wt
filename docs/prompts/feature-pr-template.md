# Feature PR Prompt Template

이 템플릿은 팀원이 어떤 에이전트를 사용하더라도 PR 품질 기준이 흔들리지 않도록 만든 “재사용 프롬프트 규격”이다.

## How to use

1. 아래 템플릿의 `<...>` 변수를 현재 작업 값으로 치환한다.
2. 치환하지 못한 변수는 `TODO:`로 명시한다.
3. 에이전트에게 그대로 전달한다.
4. 결과물(PR 본문/커밋/테스트)이 `Definition of Done`을 만족하는지 확인한다.

## Agent contract

프롬프트를 받은 에이전트는 아래를 반드시 지켜야 한다.

- 범위를 벗어난 리팩터링을 하지 않는다.
- 구현 + 테스트 + 문서 동기화를 한 PR 안에서 완료한다.
- stdout/stderr/exit code 계약이 바뀌면 PR 본문에 명시한다.
- 머지 게이트(`make premerge`)를 통과시킨다.
- 불확실한 내용은 추측으로 문서화하지 않는다.

## Template

```md
# [TASK TYPE]: <task_title>

## Context
- 이전 상태: <previous_pr_or_state>
- 현재 한계/문제: <current_limitation>
- 이번 작업 범위: <scope_summary>

## Goal
- <goal_1>
- <goal_2>

## Non-goals
- <non_goal_1>
- <non_goal_2>

## Implementation Requirements
- 대상 명령/모듈: <affected_commands_or_modules>
- 동작 계약:
  - 성공 조건: <success_contract>
  - 실패 조건: <failure_contract>
  - 에러 코드/메시지: <error_contract>
- 공통화 경계: <shared_helper_or_boundary>
- 출력 규약: <stdout_stderr_exitcode_contract>

## Test Requirements
- 단위 테스트:
  - <test_case_1>
  - <test_case_2>
- 회귀 테스트:
  - <regression_case_1>
- 옵션/모드 테스트:
  - <dry_run_or_tui_or_json_case>

## Docs Synchronization
- 수정 문서:
  - <doc_1>
  - <doc_2>
  - <release_notes_doc>
- 반영 내용:
  - <doc_behavior_delta>

## Validation
- 필수 검증: <verification_command>

## Constraints
- 금지: <unrelated_refactor_destructive_changes>
- 유지: <existing_policy_to_keep>

## PR body requirements
- Summary, User impact, Behavior, Safety, Tests, E2E guide 섹션을 모두 포함한다.
- 사용자-facing 변경이면 E2E guide를 생략하지 않는다.
- 예시에는 민감정보/로컬 절대경로를 넣지 않는다.

## Definition of Done
- [ ] 기능 계약 충족
- [ ] 테스트 통과
- [ ] 문서/릴리즈노트 동기화
- [ ] 검증 명령 통과
```

## Filled example (this repo style)

아래는 변수 치환 예시다. 실제 작업마다 값만 바꿔 재사용한다.

```md
# [Feature]: wt create/path --create final path preflight

## Context
- 이전 상태: docs-only PR에서 preflight 미구현 한계를 명시함
- 현재 한계/문제: final path가 파일/비어있지 않은 디렉터리여도 git 에러에 의존
- 이번 작업 범위: create + path --create preflight + tests + docs sync

## Goal
- 파일/비어있지 않은 디렉터리 경로를 사전검증으로 즉시 실패
- dry-run 포함 동일 규칙 적용

...
```
