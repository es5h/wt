package picker

import (
	"strings"
	"unicode"
)

type Item struct {
	ID         string
	Label      string
	Detail     string
	Meta       string
	FilterText string
}

type Key int

const (
	KeyRune Key = iota
	KeyUp
	KeyDown
	KeyPageUp
	KeyPageDown
	KeyHome
	KeyEnd
	KeyBackspace
	KeyEnter
	KeyCancel
)

type Input struct {
	Key  Key
	Rune rune
}

type Outcome int

const (
	OutcomeContinue Outcome = iota
	OutcomeAccept
	OutcomeCancel
)

type ViewItem struct {
	Item     Item
	Selected bool
}

type View struct {
	Filter     string
	Total      int
	MatchCount int
	Offset     int
	Items      []ViewItem
}

type Model struct {
	items    []Item
	filtered []int
	filter   []rune
	selected int
}

func New(items []Item, initialFilter string) *Model {
	m := &Model{
		items: append([]Item(nil), items...),
	}
	if initialFilter != "" {
		m.filter = []rune(initialFilter)
	}
	m.applyFilter()
	return m
}

func (m *Model) Filter() string {
	return string(m.filter)
}

func (m *Model) Selected() (Item, bool) {
	if len(m.filtered) == 0 {
		return Item{}, false
	}
	return m.items[m.filtered[m.selected]], true
}

func (m *Model) Update(in Input, pageSize int) Outcome {
	if pageSize < 1 {
		pageSize = 1
	}

	switch in.Key {
	case KeyRune:
		if in.Rune == 0 || unicode.IsControl(in.Rune) {
			return OutcomeContinue
		}
		m.filter = append(m.filter, in.Rune)
		m.applyFilter()
	case KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
		}
	case KeyUp:
		m.move(-1)
	case KeyDown:
		m.move(1)
	case KeyPageUp:
		m.move(-pageSize)
	case KeyPageDown:
		m.move(pageSize)
	case KeyHome:
		if len(m.filtered) > 0 {
			m.selected = 0
		}
	case KeyEnd:
		if len(m.filtered) > 0 {
			m.selected = len(m.filtered) - 1
		}
	case KeyEnter:
		if len(m.filtered) > 0 {
			return OutcomeAccept
		}
	case KeyCancel:
		return OutcomeCancel
	}

	return OutcomeContinue
}

func (m *Model) View(height int) View {
	if height < 1 {
		height = 1
	}

	offset := 0
	if len(m.filtered) > 0 {
		if m.selected < 0 {
			m.selected = 0
		}
		if m.selected >= len(m.filtered) {
			m.selected = len(m.filtered) - 1
		}

		if m.selected >= height {
			offset = m.selected - height + 1
		}
	}

	end := min(offset+height, len(m.filtered))

	items := make([]ViewItem, 0, end-offset)
	for i := offset; i < end; i++ {
		items = append(items, ViewItem{
			Item:     m.items[m.filtered[i]],
			Selected: i == m.selected,
		})
	}

	return View{
		Filter:     string(m.filter),
		Total:      len(m.items),
		MatchCount: len(m.filtered),
		Offset:     offset,
		Items:      items,
	}
}

func (m *Model) move(delta int) {
	if len(m.filtered) == 0 || delta == 0 {
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(string(m.filter)))
	filtered := make([]int, 0, len(m.items))
	for i, item := range m.items {
		if query == "" || strings.Contains(strings.ToLower(filterValue(item)), query) {
			filtered = append(filtered, i)
		}
	}
	m.filtered = filtered
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func filterValue(item Item) string {
	if item.FilterText != "" {
		return item.FilterText
	}
	return strings.TrimSpace(strings.Join([]string{item.Label, item.Detail, item.Meta}, " "))
}
