# Agent integration guide

이 문서는 `wt`를 에이전트 워크플로(Claude/Codex)에서 재사용하기 위한 실무 가이드를 정리한다.

## Goal

- 에이전트가 worktree 선택/생성/정리 명령을 일관된 방식으로 사용하도록 표준화
- 병렬 에이전트 실행 시 충돌을 줄이고, 재현 가능한 작업 로그를 남김

## Common baseline

모든 에이전트 환경에서 먼저 확인:

```sh
wt --version
wt list
```

권장 패턴:

```sh
# 작업 브랜치용 worktree 경로 확보
wt path --create <branch>

# 해당 경로에서 명령 실행
wt run <branch> -- <cmd...>

# 정리 전 미리보기
wt prune
wt cleanup
```

## Codex: skill로 등록

Codex는 `SKILL.md` 기반 스킬 구성을 사용한다.
`wt`를 자주 쓰는 팀이라면 전용 스킬을 만들어 설치해 두는 것이 가장 안정적이다.

### 1) 스킬 디렉터리 생성

예시:

```text
$CODEX_HOME/skills/wt-worktree/
  SKILL.md
```

### 2) `SKILL.md`에 `wt` 운영 규칙 작성

예시:

```md
# wt-worktree

## When to use
- worktree 생성/탐색/정리가 필요한 작업
- 병렬 에이전트 실행 시 브랜치별 격리가 필요한 작업

## Required checks
- `wt --version`
- `wt list`

## Standard flow
1. `wt path --create <branch>`로 작업 경로 확보
2. 필요한 작업 실행
3. 필요 시 `wt cleanup`/`wt prune`로 정리

## Safety
- destructive 옵션은 사용자 명시 요청 시에만 사용
- remove/prune는 기본적으로 preview 먼저 실행
```

### 3) Codex에서 사용

- 프롬프트에 스킬명을 직접 언급해 호출한다. 예: `wt-worktree 스킬을 사용해서 ...`
- 레포의 `AGENTS.md` 규칙과 함께 적용해 일관된 실행 정책을 유지한다.

## Claude: 프로젝트 가이드 + 래퍼 명령으로 연결

Claude에는 Codex의 `SKILL.md`와 동일한 표준 등록 포맷이 없을 수 있으므로,
프로젝트 지침 문서와 명시적 명령 템플릿으로 연결하는 방식을 권장한다.

### 1) 프로젝트 지침에 `wt` 표준 흐름 명시

예시 항목:

- worktree 작업은 `wt path --create <branch>`를 우선 사용
- 직접 `git worktree` 명령보다 `wt` 명령을 우선 사용
- 정리 작업은 `wt prune`/`wt cleanup` preview 후 적용

### 2) 반복 명령 템플릿 준비

에이전트 프롬프트에 아래 템플릿을 붙여 사용:

```text
- Worktree path: wt path --create <branch>
- Run command: wt run <branch> -- <cmd...>
- Cleanup preview: wt cleanup
```

### 3) 셸 helper로 사람/에이전트 공용 인터페이스 유지

```sh
eval "$(wt init zsh)"
# wtg <query>, wtr, wcd 사용 가능
```

## Recommended team policy

- 새 자동화/에이전트 작업 PR에는 사용한 `wt` 명령과 exit code를 기록한다.
- `wt upgrade`를 주기적으로 실행해 팀 에이전트 환경 버전을 맞춘다.
- 사용자-facing 변경 시 `VERSION`, `docs/release/notes.md`, 관련 UX 문서를 함께 갱신한다.
