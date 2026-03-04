package picker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

var (
	ErrCancelled = errors.New("tui picker: selection cancelled")
	ErrNonTTY    = errors.New("tui picker: requires a TTY on stdin and stderr")
)

type Config struct {
	Title         string
	Help          string
	Items         []Item
	InitialFilter string
}

func Run(input *os.File, screen *os.File, cfg Config) (Item, error) {
	if input == nil || screen == nil || !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(screen.Fd())) {
		return Item{}, ErrNonTTY
	}

	state, err := term.MakeRaw(int(input.Fd()))
	if err != nil {
		return Item{}, fmt.Errorf("tui picker: enable raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(int(input.Fd()), state)
	}()

	if _, err := screen.Write([]byte("\x1b[?1049h\x1b[?25l")); err != nil {
		return Item{}, err
	}
	defer func() {
		_, _ = screen.Write([]byte("\x1b[?25h\x1b[?1049l"))
	}()

	model := New(cfg.Items, cfg.InitialFilter)
	byteCh := make(chan byte, 16)
	errCh := make(chan error, 1)

	go func() {
		buf := make([]byte, 1)
		for {
			n, readErr := input.Read(buf)
			if n > 0 {
				byteCh <- buf[0]
			}
			if readErr != nil {
				errCh <- readErr
				return
			}
		}
	}()

	for {
		if err := render(screen, cfg, model); err != nil {
			return Item{}, err
		}

		key, err := readInput(byteCh, errCh)
		if err != nil {
			return Item{}, fmt.Errorf("tui picker: read input: %w", err)
		}

		_, rows, sizeErr := term.GetSize(int(screen.Fd()))
		if sizeErr != nil || rows < 6 {
			rows = 12
		}

		switch model.Update(key, rows-4) {
		case OutcomeAccept:
			selected, ok := model.Selected()
			if !ok {
				continue
			}
			return selected, nil
		case OutcomeCancel:
			return Item{}, ErrCancelled
		}
	}
}

func render(screen *os.File, cfg Config, model *Model) error {
	_, rows, err := term.GetSize(int(screen.Fd()))
	if err != nil || rows < 6 {
		rows = 12
	}
	listHeight := max(rows-4, 1)

	view := model.View(listHeight)
	var b bytes.Buffer
	b.WriteString("\x1b[H\x1b[2J")
	fmt.Fprintf(&b, "%s\n", firstNonEmpty(cfg.Title, "Select worktree"))
	fmt.Fprintf(&b, "Filter: %s\n", view.Filter)

	if view.MatchCount == 0 {
		b.WriteString("  no matches\n")
	} else {
		for _, item := range view.Items {
			prefix := "  "
			if item.Selected {
				prefix = "> "
			}
			line := item.Item.Label
			if item.Item.Meta != "" {
				line += " [" + item.Item.Meta + "]"
			}
			fmt.Fprintf(&b, "%s%s\n", prefix, line)
			if item.Item.Detail != "" {
				fmt.Fprintf(&b, "    %s\n", item.Item.Detail)
			}
		}
	}

	fmt.Fprintf(&b, "%d/%d matches", view.MatchCount, view.Total)
	help := strings.TrimSpace(cfg.Help)
	if help != "" {
		fmt.Fprintf(&b, " | %s", help)
	}
	b.WriteString("\n")

	_, err = screen.Write(b.Bytes())
	return err
}

func readInput(byteCh <-chan byte, errCh <-chan error) (Input, error) {
	select {
	case b := <-byteCh:
		return decodeInput(b, byteCh)
	case err := <-errCh:
		return Input{}, err
	}
}

func decodeInput(first byte, byteCh <-chan byte) (Input, error) {
	switch first {
	case 3:
		return Input{Key: KeyCancel}, nil
	case 10, 13:
		return Input{Key: KeyEnter}, nil
	case 8, 127:
		return Input{Key: KeyBackspace}, nil
	case 11:
		return Input{Key: KeyUp}, nil
	case 14:
		return Input{Key: KeyDown}, nil
	case 16:
		return Input{Key: KeyUp}, nil
	case 27:
		return decodeEscape(byteCh), nil
	default:
		return Input{Key: KeyRune, Rune: rune(first)}, nil
	}
}

func decodeEscape(byteCh <-chan byte) Input {
	select {
	case next := <-byteCh:
		if next != '[' {
			return Input{Key: KeyCancel}
		}

		select {
		case code := <-byteCh:
			switch code {
			case 'A':
				return Input{Key: KeyUp}
			case 'B':
				return Input{Key: KeyDown}
			case 'H':
				return Input{Key: KeyHome}
			case 'F':
				return Input{Key: KeyEnd}
			case '5':
				if discardTilde(byteCh) {
					return Input{Key: KeyPageUp}
				}
			case '6':
				if discardTilde(byteCh) {
					return Input{Key: KeyPageDown}
				}
			}
		case <-time.After(25 * time.Millisecond):
		}
	case <-time.After(25 * time.Millisecond):
		return Input{Key: KeyCancel}
	}

	return Input{Key: KeyCancel}
}

func discardTilde(byteCh <-chan byte) bool {
	select {
	case b := <-byteCh:
		return b == '~'
	case <-time.After(25 * time.Millisecond):
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
