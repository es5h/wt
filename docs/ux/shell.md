# Shell integration and completion

이 문서는 현재 구현된 셸 helper와 completion 사용 흐름을 설명한다.

## Recommended Flow

1. `wt`를 설치한다.
2. `wt init <shell>` 출력으로 helper를 셸에 추가한다.
3. 필요하면 `wt completion <shell>`을 설치해 자동완성을 켠다.

예시:

```sh
eval "$(wt init zsh)"
```

## `wt init <shell>`

`wt init`은 rc 파일을 직접 수정하지 않고 helper 정의만 `stdout`으로 출력한다.
출력 상단에는 셸별 opt-in 설치 가이드(즉시 적용, rc 영구 적용, completion 설치 예시)가 주석으로 포함된다.

현재 포함되는 helper:

- `wtr()`: `cd "$(wt root)" || return`
- `wtg()`: `cd "$(wt path "$@")" || return`
- `wcd()`: `cd "$(wt path "$@")" || return`

의미:

- `wtr`: 현재 Git 컨텍스트의 primary repo root로 이동
- `wtg`: query로 worktree를 찾아 이동
- `wcd`: `wtg`와 동일 동작의 별칭 helper

## zsh

helper 적용:

```sh
eval "$(wt init zsh)"
```

영구 적용:

```sh
wt init zsh >> ~/.zshrc
```

`wt init zsh`에는 completion bridge가 포함된다. `_wt` completion이 설치되어 있으면 `wtg <TAB>`와 `wcd <TAB>`가 `wt path <TAB>`처럼 동작한다.

## bash

```sh
eval "$(wt init bash)"
```

## fish

```sh
wt init fish | source
```

## Completion

`wt`는 Cobra 기본 `wt completion <shell>` 명령을 제공한다.
`wt init <shell>`은 completion을 자동 설치/자동 로드하지 않고, 아래 설치 명령을 주석으로 안내만 한다.

zsh 설치:

```sh
mkdir -p ~/.zsh/completions
wt completion zsh > ~/.zsh/completions/_wt
```

`~/.zshrc` 예시:

```sh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

bash 설치:

```sh
mkdir -p ~/.bash_completion.d
wt completion bash > ~/.bash_completion.d/wt
```

fish 설치:

```sh
mkdir -p ~/.config/fish/completions
wt completion fish > ~/.config/fish/completions/wt.fish
```

## `wt path` Query Completion

현재 동적 후보는 `git worktree list --porcelain`에 등록된 worktree 기준이다.

- linked worktree가 있는 브랜치명은 completion 후보에 포함된다.
- detached entry는 basename 기반으로 후보가 잡힐 수 있다.
- 기본값은 현재 등록된 worktree 후보만 제안한다.

원격 브랜치 후보를 추가하고 싶으면:

```sh
export WT_PATH_COMPLETE_REMOTE=1
```

이 경우에도 자동 `git fetch`는 하지 않는다. 로컬에 이미 존재하는 `refs/remotes/origin/*`만 사용한다.

## Troubleshooting

`wtg`나 `wcd`에서 탭이 파일명 완성으로만 동작하면 `_wt`가 로드되지 않은 상태일 가능성이 높다.

진단:

```sh
whence -v _wt || true
```

`_wt not found`면 completion을 먼저 설치해야 한다.
