# TUI (vim-ish picker) spec (draft)

## Goal
여러 worktree 후보 중 하나를 빠르게 선택해 `wt goto`가 경로를 반환하게 한다.

## Entry
- `wt goto` (query 없음) → 기본적으로 picker를 띄워 전체 목록에서 선택
- `wt goto <query> --tui` → 필터링된 후보를 picker로 선택

비활성 조건(초안):
- `stdout` 또는 `stdin`이 터미널이 아니면(파이프/CI) TUI 실행 금지(에러 또는 `--no-tui` 동작)

관련 CLI 스펙은 `docs/spec/cli.md` 참고.

## UI layout (초안)
- 상단: 입력(query) / 상태 라인
- 본문: 후보 리스트(이름/브랜치, 경로, 짧은 해시, locked 표시)
- 하단: 도움말(키 바인딩)

## Keybindings (초안)
- 이동: `j/k`, `Down/Up`
- 페이지: `Ctrl+d` / `Ctrl+u`
- 선택: `Enter`
- 취소: `Esc` 또는 `Ctrl+c`
- 검색: 타이핑 시 실시간 필터
- 토글(선택): `l`로 locked만 보기, `p`로 path 표시 토글

## Selection result
선택 확정 시:
- stdout: 선택된 worktree의 path만 출력(기본 모드)
- exit code: 0

취소 시:
- stderr: 취소 메시지(선택)
- exit code: 130(CTRL+C 관례) 또는 1(정책 확정 필요)

## Implementation note
- TUI는 도메인 로직과 분리(리스트/필터/선택 상태는 순수 구조체로)
- TUI 도입시 `bubbletea` 사용