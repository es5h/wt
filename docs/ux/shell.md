# Shell integration & completion (draft)

## Why
- `wt init`/completion 동작을 문서화한다. (관련 CLI 스펙: `docs/spec/cli.md`)

## Low-level completion (recommended)
초기에는 “가볍고 안전한” 자동완성을 권장합니다. (서브커맨드/플래그 중심)

`cobra` 기반 CLI는 `wt completion <shell>`을 기본 제공하므로, 이를 그대로 사용합니다.

### 먼저 확인(진단)
아래가 `"_wt not found"`이면 현재는 `wt` 전용 completion이 “설치/로딩”되어 있지 않은 상태입니다. (이 경우 `wt<TAB>`에서 파일명이 나오는 것은 zsh 기본 파일명 completion입니다.)

```sh
whence -v _wt || true
```

### 동적 후보(현재 구현)
`wt goto <query>`의 `<query>` 위치에서는, 현재 레포의 `git worktree list --porcelain` 결과를 기반으로 **기존 worktree 브랜치 이름**을 자동완성 후보로 제공합니다.

- 안전성: 읽기 전용(`git worktree list`)만 호출
- 성능: 짧은 git 호출 1회 수준
- 제약: “없는 브랜치 생성(--create)”은 아직 미구현이므로, 후보는 “이미 존재하는 worktree”에 한정됩니다.

### zsh 설치(옵트인)
```sh
mkdir -p ~/.zsh/completions
wt completion zsh > ~/.zsh/completions/_wt

# ~/.zshrc에 아래가 없다면 추가
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

#### oh-my-zsh 사용 시(권장)
oh-my-zsh를 쓰면 기본적으로 `~/.oh-my-zsh/custom/completions`가 `fpath`에 들어가 있으니, 아래가 가장 간단합니다.

```sh
mkdir -p ~/.oh-my-zsh/custom/completions
wt completion zsh > ~/.oh-my-zsh/custom/completions/_wt

rm -f ~/.zcompdump*
autoload -Uz compinit && compinit

whence -v _wt
```

### bash 설치(옵트인)
```sh
wt completion bash > ~/.bash_completion.d/wt
source ~/.bash_completion.d/wt
```

### fish 설치(옵트인)
```sh
mkdir -p ~/.config/fish/completions
wt completion fish > ~/.config/fish/completions/wt.fish
```

## `wt init <shell>`
목표: `cd "$(wt goto ...)"`를 래핑하는 함수를 제공한다.

권장 UX(초안):
- 사용자는 아래 중 하나로 rc 파일에 추가한다.

예시(컨셉, 스펙 확정 전):
- zsh/bash:
  - `wtg() { cd "$(wt goto "$@")" || return; }`
- fish:
  - `function wtg; cd (wt goto $argv); or return; end`

### 사용(추천)
```sh
wt init zsh
```

또는(즉시 적용, opt-in):
```sh
eval "$(wt init zsh)"
```

추가(초안):
- completion 설치 UX를 `wt init`에 포함할지 여부는 추후 결정(로드맵: `docs/roadmap/README.md`)

## Completion 설계(초안)

### 커맨드 형태
- (계획) `wt completion <shell>`: 해당 셸용 completion 스크립트를 stdout으로 출력 (로드맵: `docs/roadmap/README.md`)

### 후보 생성 규칙
- completion은 “빠르고 부작용이 없어야” 한다.
- 후보 목록은 기본적으로 `wt list --porcelain` 또는 `wt list --json`을 기반으로 생성한다.

### 설치(문서화만)
- zsh: `~/.zshrc`에서 `source <(wt completion zsh)` 형태를 지원 고려
- bash: `source <(wt completion bash)`
- fish: 출력 파일을 `~/.config/fish/completions/wt.fish`로 저장

## Notes
- completion은 터미널 기능이라 테스트는 “생성된 스크립트 문자열” 수준의 스냅샷 테스트로 시작하는 것을 권장한다.
