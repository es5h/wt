# PR guidelines

이 문서는 `wt` 레포에서 PR을 일관되게 작성/리뷰하기 위한 가이드입니다.

## Goals
- PR 하나는 한 가지 주제(기능/리팩터링/문서)만 다룬다.
- 사용자에게 보이는 변경(명령/옵션/출력/기본값)은 스펙/릴리즈 노트에 반드시 반영한다.
- 파괴적 동작(삭제/prune 등)은 기본값으로 하지 않고 opt-in으로 둔다.
- 자동화 에이전트가 작성하는 PR 본문은 기본적으로 한글로 작성한다.
- PR 본문/예시/재현 가이드에는 로컬 절대경로, 계정명, 비공개 URL 등 민감정보를 넣지 않는다.

## Merge gate
- 머지(또는 main push) 전 `make premerge` 통과가 기준이다.
- 사용자에게 보이는 변경이 있으면:
  - 스펙: `docs/spec/cli.md` 업데이트
  - 릴리즈 노트: `docs/release/notes.md`의 `## Unreleased`에 기록
  - 버전: `VERSION` bump 필요 여부 확인

## PR body template
아래 섹션을 PR 본문에 포함하는 것을 권장합니다.

추가:
- GitHub PR template: `.github/pull_request_template.md`를 기본으로 사용합니다.
- 특히 `## E2E guide`는 PR에서 빠지기 쉬우므로, CI에서 “섹션 존재”를 체크합니다(섹션이 불필요하면 `N/A` + 이유를 적습니다).

작성 규칙:
- 기본 언어는 한글로 한다. (코드/명령/경로/식별자는 원문 유지 가능)
- 예시 경로는 실제 로컬 경로 대신 `/path/to/wt`, `<repo>`, `<query>` 같은 플레이스홀더를 사용한다.
- 사용자-facing 변경이면 `User impact`, `Behavior`, `Safety`, `Tests`, `E2E guide`를 생략하지 않는다.

### Summary
- 무엇을/왜 변경했는지 3~5줄

### User impact
- 사용자에게 보이는 변화가 있으면 bullet로 명시(없으면 “None”)
- stdout/stderr 규칙, exit code 변화가 있으면 반드시 명시

### Behavior (examples)
- 실행 예 1~2개(특히 새 플래그/출력)

### Safety
- 파괴적 동작 여부, 기본값의 안전성, opt-in/confirm/--dry-run 정책

### Tests
- `make premerge`
- 필요 시 e2e 커맨드(예: `wt list --verify --base origin/main`)

### E2E guide (추천)
PR에서 사용자-facing 동작이 추가/변경되면, 아래 중 하나로 e2e 재현 커맨드를 남기는 것을 권장합니다.

옵션 A) 현재 repo에서 실행(빠름)
- `make run ARGS="--help"`
- `make run ARGS="list --verify --base origin/main"`
- `make run ARGS="list"`로 후보를 확인한 뒤, 실제 존재하는 query로 실행: `make run ARGS="path <query>"`

주의:
- `wt path`는 “현재 Git 컨텍스트의 worktree 목록”에서만 매칭하므로, 해당 repo에 worktree가 1개뿐이면 테스트할 query가 없을 수 있습니다.
- 이런 경우에는 아래 옵션 B(임시 repo) 방식으로 재현하는 것을 권장합니다.

옵션 B) 임시 repo에서 실행(가장 확실함)
```sh
tmp="$(mktemp -d)"
wt_repo="/path/to/wt"
cd "$tmp"
git init
touch a && git add a && git commit -m init
git worktree add -b feature-x ../wt-feature-x

# wt 실행(아직 설치 전이라면 go run 사용)
go run "$wt_repo"/cmd/wt list
go run "$wt_repo"/cmd/wt path feature-x
```

민감정보 주의:
- 홈 디렉토리(`/home/...`, `/Users/...`)나 실제 checkout 경로를 그대로 넣지 않는다.
- 개인/사내 전용 URL, 토큰, 사용자명, 호스트명은 예시에 포함하지 않는다.

### Docs
- 수정한 문서 링크(스펙/정책/릴리즈 노트)

## Commit hygiene
- 기능 PR은 `feat(...)` 프리픽스를 권장한다.
- 리팩터링/정리는 `chore:` 또는 `refactor:`를 사용한다.
- “동작 변경 없는 리팩터링”과 “기능 추가”를 같은 커밋에 섞지 않는다.
