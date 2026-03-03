# Shell integration & completion (draft)

## Why
- `wt goto`는 프로세스가 직접 `cd`를 할 수 없으므로, “경로 출력 + 셸 함수” 방식이 필요하다.
- `wt goto <TAB>` 같은 completion은 UX 핵심이라 문서/스펙을 먼저 고정한다.

## `wt init <shell>`
목표: `cd "$(wt goto ...)"`를 래핑하는 함수를 제공한다.

권장 UX(초안):
- 사용자는 rc 파일에 아래 중 하나를 넣는다.

예시(컨셉, 스펙 확정 전):
- zsh/bash:
  - `wtg() { cd "$(wt goto "$@")" || return; }`
- fish:
  - `function wtg; cd (wt goto $argv); or return; end`

추가(초안):
- `wt init zsh --completion` 같은 방식으로 completion까지 같이 설치할지 여부는 추후 결정

## Completion 설계(초안)

### 커맨드 형태
- `wt completion <shell>`: 해당 셸용 completion 스크립트를 stdout으로 출력

### 후보 생성 규칙
- completion은 “빠르고 부작용이 없어야” 한다.
- 후보 목록은 기본적으로 `wt list --porcelain` 또는 `wt list --json`을 기반으로 생성한다.

### 설치(문서화만)
- zsh: `~/.zshrc`에서 `source <(wt completion zsh)` 형태를 지원 고려
- bash: `source <(wt completion bash)`
- fish: 출력 파일을 `~/.config/fish/completions/wt.fish`로 저장

## Notes
- completion은 터미널 기능이라 테스트는 “생성된 스크립트 문자열” 수준의 스냅샷 테스트로 시작하는 것을 권장한다.

