# Feature PR Prompt Template

이 템플릿은 팀원이 어떤 에이전트를 사용하더라도 PR 품질 기준이 흔들리지 않도록 만든 “재사용 프롬프트 규격”이다.
특히 병렬 에이전트 작업에서도 같은 수준의 작업 지시 품질을 반복 가능하게 만드는 것을 목표로 한다.

## How to use

1. 작업 시작 컨텍스트로 아래 문서를 함께 붙인다.
   - `@docs/prompts/feature-pr-template.md`
   - `@docs/pr-guidelines.md`
   - `@AGENTS.md`
2. 아래 템플릿의 `<...>` 변수를 현재 작업 값으로 치환한다.
3. 치환하지 못한 변수는 `TODO:`로 남기지 말고, 범위를 줄이거나 가정을 명시한다.
4. 병렬 작업이면 각 에이전트에 고유 branch/worktree 이름을 준다.
5. 에이전트에게 그대로 전달한다.
6. 결과물(PR 본문/커밋/테스트)이 `Definition of Done`을 만족하는지 확인한다.

## Prompt design goals

- 시작 3~6줄 안에 `Goal`, `Constraints`, `Definition of Done`이 드러나야 한다.
- 에이전트가 바로 실행할 수 있게 branch/worktree 전략을 포함해야 한다.
- 구현 범위, 비범위, 검증 기준, 문서 반영 범위를 한 번에 판단할 수 있어야 한다.
- PR 본문 품질 기준이 프롬프트 안에 그대로 연결돼야 한다.
- 애매한 TODO 슬롯보다, 실제 결정을 강제하는 체크포인트가 많아야 한다.

## Agent contract

프롬프트를 받은 에이전트는 아래를 반드시 지켜야 한다.

- 범위를 벗어난 리팩터링을 하지 않는다.
- 구현 + 테스트 + 문서 동기화를 한 PR 안에서 완료한다.
- stdout/stderr/exit code 계약이 바뀌면 PR 본문에 명시한다.
- 머지 게이트(`make premerge`)를 통과시킨다.
- 불확실한 내용은 추측으로 문서화하지 않는다.
- 구현 작업은 기본적으로 `wt` 분리 워크트리에서 진행한다.
- 사용자-facing 변경 PR에서는 E2E 명령을 실제 실행하고 결과를 PR 본문에 남긴다.
- 사용자-facing 변경 PR에서는 같은 PR에서 `VERSION`을 반드시 bump 한다.

## Recommended prompt shape

프롬프트는 아래 4개 층위를 유지하는 것이 좋다.

1. 작업 요약
- `Goal`
- `Constraints`
- `Definition of Done`

2. 실행 지시
- 사용할 skill
- worktree/branch 이름
- 선행 확인 명령
- 구현 대상 파일/모듈

3. 품질 기준
- 기능 계약
- 테스트 요구사항
- 문서 동기화 범위
- 검증 명령

4. PR 산출물 요구사항
- PR 본문 섹션
- E2E 기록 방식
- 민감정보/로컬 절대경로 금지

## Template

```md
# [TASK TYPE]: <task_title>

@docs/prompts/feature-pr-template.md
@docs/pr-guidelines.md
@AGENTS.md

Goal
- <goal_1>
- <goal_2>

Constraints
- <constraint_1>
- <constraint_2>
- <$skill_name> 스킬을 사용한다.
- 시작 전에 `wt --version`, `wt list`를 실행한다.
- `wt path --create <branch_name>`로 분리 worktree를 만든 뒤 그 worktree에서만 작업한다.

Definition of Done
- <dod_1>
- <dod_2>
- `make premerge`를 통과한다.
- PR 본문은 한글로 작성하고 `Summary`, `User impact`, `Behavior`, `Safety`, `Tests`, `E2E guide`, `E2E Done` 섹션을 포함한다.

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

## Worktree plan
- skill: <$skill_name>
- branch: <branch_name>
- worktree 확보: `wt path --create <branch_name>`
- 작업 디렉터리: <worktree_path_or_rule>

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

## E2E Execution Requirements
- 작업 분리: <wt_split_command>
- 실행 명령:
  - <e2e_command_1>
  - <e2e_command_2>
- 결과 기록:
  - 각 명령의 exit code
  - stdout/stderr 핵심 요약
  - 실패/스킵 사유
- TUI 항목 처리:
  - <tui_execution_or_skip_rule>

## Constraints
- 금지: <unrelated_refactor_destructive_changes>
- 유지: <existing_policy_to_keep>

## PR body requirements
- Summary, User impact, Behavior, Safety, Tests, E2E guide, E2E Done 섹션을 모두 포함한다.
- 사용자-facing 변경이면 E2E guide를 생략하지 않는다.
- 사용자-facing 변경이면 E2E Done에서 실행 결과를 체크리스트로 남긴다.
- 예시에는 민감정보/로컬 절대경로를 넣지 않는다.

## Definition of Done
- [ ] 기능 계약 충족
- [ ] 테스트 통과
- [ ] 문서/릴리즈노트 동기화
- [ ] 사용자-facing 변경이면 `VERSION` bump 반영
- [ ] 검증 명령 통과
- [ ] E2E 실행 결과(명령/exit code/요약/스킵 사유) 기록
```

## Prompt checklist

프롬프트를 넘기기 전에 아래를 확인한다.

- branch/worktree 이름이 명시되어 있는가
- skill 사용 여부가 명시되어 있는가
- 기능 계약과 non-goals가 분리되어 있는가
- 변경해야 할 문서와 코드 범위가 적혀 있는가
- 검증 명령과 E2E 기록 방식이 적혀 있는가
- PR 본문 섹션 요구사항이 포함되어 있는가

## Filled example (this repo style)

아래는 변수 치환 예시다. 실제 작업마다 값만 바꿔 재사용한다.

```md
# [Feature]: wt create/path --create final path preflight

@docs/prompts/feature-pr-template.md
@docs/pr-guidelines.md
@AGENTS.md

Goal
- 파일/비어있지 않은 디렉터리 경로를 사전검증으로 즉시 실패시킨다.
- dry-run에도 같은 preflight 규칙을 적용한다.

Constraints
- `wt-worktree` 스킬을 사용한다.
- 시작 전에 `wt --version`, `wt list`를 실행한다.
- `wt path --create feat/wt-create-preflight`로 분리 worktree를 만든 뒤 그 worktree에서만 작업한다.
- 범위를 벗어난 CLI 개편이나 remove/prune 정책 변경은 하지 않는다.

Definition of Done
- `wt create`와 `wt path --create`가 동일 preflight helper를 사용한다.
- 실패 케이스와 dry-run 케이스 테스트가 추가된다.
- 관련 문서와 릴리즈 노트가 동기화된다.
- `make premerge`를 통과한다.

## Context
- 이전 상태: docs-only PR에서 preflight 미구현 한계를 명시함
- 현재 한계/문제: final path가 파일/비어있지 않은 디렉터리여도 git 에러에 의존
- 이번 작업 범위: create + path --create preflight + tests + docs sync

## Non-goals
- `wt remove`, `wt prune`, `wt cleanup` 정책 변경
- path resolution 우선순위 재설계

## Worktree plan
- skill: `wt-worktree`
- branch: `feat/wt-create-preflight`
- worktree 확보: `wt path --create feat/wt-create-preflight`
- 작업 디렉터리: 생성된 분리 worktree 경로

## Implementation Requirements
- 대상 명령/모듈: `cmd/wt/create.go`, `cmd/wt/goto.go`, shared preflight helper
- 동작 계약:
  - 성공 조건: 경로가 없거나 빈 디렉터리면 기존 생성 흐름 유지
  - 실패 조건: 기존 파일, 비어있지 않은 디렉터리, symlink, 기타 unsupported 타입이면 usage error
  - 에러 코드/메시지: usage error(exit code 2)와 명령명 포함 에러 메시지 유지
- 공통화 경계: create/path --create 가 공유할 preflight helper 로 정리
- 출력 규약: 기본 stdout path-only 유지, dry-run preview는 stderr 유지

## Test Requirements
- 단위 테스트:
  - 기존 파일 경로면 즉시 실패
  - 비어있지 않은 디렉터리면 즉시 실패
- 회귀 테스트:
  - 경로가 없을 때 기존 create 동작 유지
- 옵션/모드 테스트:
  - `--dry-run`에도 동일 preflight 적용

## Docs Synchronization
- 수정 문서:
  - `docs/spec/cli.md`
  - `docs/policy/worktree.md`
  - `docs/release/notes.md`
- 반영 내용:
  - preflight 규칙, dry-run 동작, usage error 계약

## Validation
- 필수 검증: `make premerge`

## E2E Execution Requirements
- 작업 분리: `wt path --create feat/wt-create-preflight`
- 실행 명령:
  - `make test`
  - `make run ARGS="create feature-x --dry-run"`
- 결과 기록:
  - 각 명령의 exit code
  - stdout/stderr 핵심 요약
  - 실패/스킵 사유
- TUI 항목 처리:
  - 해당 없음

## Constraints
- 금지: 범위 밖 리팩터링, 파괴적 기본값, unrelated UX 변경
- 유지: stdout/stderr 분리, path-only 기본 출력, non-interactive safety 정책

## PR body requirements
- Summary, User impact, Behavior, Safety, Tests, E2E guide, E2E Done 섹션을 모두 포함한다.
- 사용자-facing 변경이면 E2E guide를 생략하지 않는다.
- 사용자-facing 변경이면 E2E Done에서 실행 결과를 체크리스트로 남긴다.
- 예시에는 민감정보/로컬 절대경로를 넣지 않는다.
```
