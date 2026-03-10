# Release process

`wt` 릴리즈는 Go module 태그(`vX.Y.Z`)를 기준으로 배포한다.

## Rules

- `VERSION`은 `X.Y.Z` 형식(접두사 `v` 없음)을 사용한다.
- Git tag는 반드시 `v$(cat VERSION)` 형식을 사용한다.
- 사용자-facing 변경은 merge 전에 `docs/release/notes.md`의 `## Unreleased`에 기록한다.
- 릴리즈 설치 경로는 `go install github.com/es5h/wt/cmd/wt@latest`를 기준으로 유지한다.

## Release steps

1. main에 반영할 PR에서 `VERSION`을 bump 하고 `docs/release/notes.md`를 업데이트한다.
2. main 머지 후 로컬에서 최신 main을 가져온다.
3. `VERSION` 값을 확인하고 동일 버전의 태그를 생성한다.
4. 태그를 origin에 push 한다.
5. GitHub Actions `release` 워크플로가 태그 검증 후 GitHub Release를 자동 생성한다.

예시:

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
- GitHub Actions `ci`는 tag push 시 `v<semver>` 형식과 `VERSION` 파일 일치 여부를 자동 검증한다.
- GitHub Actions `release`는 같은 검증을 통과한 tag에 대해 Release를 자동 발행한다.
