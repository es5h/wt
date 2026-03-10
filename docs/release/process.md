# Release process

`wt` 릴리즈는 Go module 태그(`vX.Y.Z`)를 기준으로 배포한다.

## Rules

- `VERSION`은 `X.Y.Z` 형식(접두사 `v` 없음)을 사용한다.
- Git tag는 반드시 `v$(cat VERSION)` 형식을 사용한다.
- 사용자-facing 변경은 merge 전에 `docs/release/notes.md`의 `## Unreleased`에 기록한다.
- 릴리즈 설치 경로는 `go install github.com/es5h/wt/cmd/wt@latest`를 기준으로 유지한다.
- `VERSION`을 변경한 PR은 같은 PR에서 `docs/release/notes.md`도 반드시 함께 갱신한다.
- `VERSION`은 base 대비 증가해야 하며, 이미 존재하는 태그와 충돌하면 안 된다.

## Release steps

1. main에 반영할 PR에서 `VERSION`을 bump 하고 `docs/release/notes.md`를 업데이트한다.
2. main 머지 후 `auto-tag` 워크플로가 push 범위에서 `VERSION` 변경을 감지하면 `v$(cat VERSION)` 태그를 자동 생성/푸시한다.
3. 태그 push 이벤트로 `release` 워크플로가 태그 검증 후 GitHub Release를 자동 생성한다.

수동 태깅(예외 상황에서만):

```sh
git switch main
git pull --ff-only
VERSION="$(cat VERSION)"
git tag -a "v${VERSION}" -m "release: v${VERSION}"
git push origin "v${VERSION}"
```

## Verification

- `go list -m github.com/es5h/wt@latest`가 방금 배포한 태그를 가리키는지 확인한다.
- `go install github.com/es5h/wt/cmd/wt@latest` 후 `wt --version` 출력이 태그 버전과 일치하는지 확인한다.
- GitHub Actions `ci`는 PR/main push에서 `VERSION`/`docs/release/notes.md` 정책과 semver 증가를 검증한다.
- GitHub Actions `ci`는 tag push 시 `v<semver>` 형식과 `VERSION` 파일 일치 여부를 자동 검증한다.
- GitHub Actions `auto-tag`는 main push에서 `VERSION` 변경 시에만 태그를 생성하고, 태그 중복 시 실패한다.
- GitHub Actions `release`는 같은 검증을 통과한 tag에 대해 Release를 자동 발행한다.

## Auto-tag failure response

- `auto-tag` 실패 시 먼저 실패 원인을 확인한다: `VERSION`/`docs/release/notes.md` 동반 변경 누락, semver 형식 오류, 버전 증가 규칙 위반, 기존 tag 충돌.
- 정책 위반이면 fix PR로 `VERSION`과 `docs/release/notes.md`를 함께 수정하고 main에 다시 머지한다(강제 push/태그 덮어쓰기 금지).
- 기존 tag 충돌이면 이미 릴리즈된 버전으로 판단하고 `VERSION`을 다음 semver로 올린 PR을 만든다.
- 예외적으로 수동 태깅이 필요하면 PR/이슈에 사유와 실행 로그를 남기고, 태그는 반드시 `v$(cat VERSION)` 규칙을 지킨다.
