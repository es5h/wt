# Agent integration guide

이 문서는 `wt`를 Claude/Codex 에이전트 워크플로에서 재사용하기 위한 실무 가이드다.

## Key point

Claude Code와 Codex 모두 `SKILL.md` 기반 스킬을 지원한다.
따라서 `wt` 운영 규칙은 공통 스킬 문서로 관리하고, 도구별 차이는 설치 경로/호출 방식만 분리하는 것을 권장한다.

## Common baseline

모든 에이전트 환경에서 먼저 확인:

```sh
wt --version
wt list
```

권장 실행 흐름:

```sh
# 작업 브랜치용 worktree 경로 확보
wt path --create <branch>

# 해당 worktree에서 명령 실행
wt run <branch> -- <cmd...>

# 정리 전 미리보기
wt prune
wt cleanup
```

## Shared skill template (`SKILL.md`)

아래 템플릿을 공통으로 사용하면 된다.

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

## Codex

스킬 위치 예시:

```text
$CODEX_HOME/skills/wt-worktree/
  SKILL.md
```

호출 예시:

```text
wt-worktree 스킬을 사용해서 <task>
```

## Claude Code

스킬 위치 예시(프로젝트/사용자):

```text
.claude/skills/wt-worktree/SKILL.md
~/.claude/skills/wt-worktree/SKILL.md
```

호출 예시:

```text
wt-worktree 스킬을 사용해서 <task>
```

참고:

- Claude Code는 커스텀 커맨드가 skills로 통합되었다.
- 하위 에이전트는 `.claude/agents/*.md` 또는 `~/.claude/agents/*.md`를 사용한다.

## Recommended team policy

- 새 자동화/에이전트 작업 PR에는 사용한 `wt` 명령과 exit code를 기록한다.
- `wt upgrade`를 주기적으로 실행해 팀 에이전트 환경 버전을 맞춘다.
- 사용자-facing 변경 시 `VERSION`, `docs/release/notes.md`, 관련 UX 문서를 함께 갱신한다.
