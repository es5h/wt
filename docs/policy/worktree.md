# Policy (draft)

이 문서는 `wt` 내부 정책(경로/정규화 등)을 “구현 근거가 되는 규칙” 형태로 기록합니다.

## Git context
- “Git 컨텍스트(repo root)”는 현재 디렉토리에서 `git rev-parse --show-toplevel`로 결정한다.

## Default worktree root
기본 생성 경로 정책:
- 기본 생성 경로는 `<primary-root>/.wt/<branch>` 이다.
- `.wt/`는 레포의 산출물/로컬 작업 디렉토리이므로 git에서 추적하지 않는다(`/.wt/`를 `.gitignore`에 추가).
- 이 기본값은 재현성을 위해 유지하고, 다른 레이아웃은 opt-in 오버라이드로만 적용한다.

`<primary-root>` 결정 규칙:
- `git rev-parse --path-format=absolute --git-common-dir`로 “공유 git 디렉토리”를 구한다.
- 그 부모 디렉토리를 `<primary-root>`로 사용한다.

의도:
- linked worktree 내부에서 `wt create`/`wt goto --create`를 실행해도, 기본 경로가 “현재 worktree 아래”로 잡혀 중첩되는 문제(예: `.wt/a/.wt/b/...`)를 방지한다.

## Overrides (opt-in)
기본 정책은 재현성을 위해 고정하되, 사용자/팀 환경에 맞게 “명시적으로” 오버라이드할 수 있게 한다.

우선순위(높음 → 낮음):
1) CLI flag `--root`
2) Environment variable `WT_ROOT`
3) Repo-local git config `wt.root` (`git config --local wt.root ...`)
4) Default policy (`<primary-root>/.wt`)

규칙:
- 이 우선순위는 `wt create`와 `wt goto --create`에 동일하게 적용한다.
- `--path`가 지정되면 최종 생성 경로를 직접 지정한 것으로 보고 root 정책보다 우선한다.
- `--root`, `WT_ROOT`, `wt.root` 값이 상대 경로이면 모두 `<primary-root>` 기준으로 해석한다.
- `wt.root`는 repo-local config만 읽는다. global/system git config는 이 정책에 포함하지 않는다.

지원 environment variable:
- `WT_ROOT`: worktree 루트 디렉토리(절대/상대 경로)

repo-local git config 예시:
- `git config --local wt.root ../.wt`
- `git config --local wt.root .worktrees`

주의:
- 자동으로 사용자 홈/셸 설정을 변경하는 동작은 기본값으로 두지 않는다.
- 오버라이드는 “없는 경우에만 적용”이 아니라, 명시적 설정이 있는 경우에만 적용하는 것을 원칙으로 한다.

## Branch name → directory name normalization
- 브랜치명 → 폴더명 정규화 규칙을 유틸로 통일한다.
  - 현재 기본 경로는 git 브랜치의 `/`를 하위 디렉토리로 유지해 `<root>/<branch>` 형태를 만든다.
  - 절대 경로/상위 디렉토리 탈출(`..`)이 되는 브랜치명은 기본 경로 계산에 사용할 수 없다.
