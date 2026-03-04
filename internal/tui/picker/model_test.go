package picker

import "testing"

func TestModelFiltersAndSelectsFirstMatch(t *testing.T) {
	t.Parallel()

	model := New([]Item{
		{ID: "1", Label: "feature-one", Detail: "/repo/.wt/feature-one"},
		{ID: "2", Label: "bugfix-two", Detail: "/repo/.wt/bugfix-two"},
	}, "bug")

	view := model.View(5)
	if view.MatchCount != 1 {
		t.Fatalf("MatchCount = %d, want 1", view.MatchCount)
	}

	selected, ok := model.Selected()
	if !ok {
		t.Fatal("Selected() = false, want true")
	}
	if selected.ID != "2" {
		t.Fatalf("Selected().ID = %q, want %q", selected.ID, "2")
	}
}

func TestModelMovesAndPages(t *testing.T) {
	t.Parallel()

	model := New([]Item{
		{ID: "1", Label: "one"},
		{ID: "2", Label: "two"},
		{ID: "3", Label: "three"},
		{ID: "4", Label: "four"},
	}, "")

	model.Update(Input{Key: KeyDown}, 2)
	model.Update(Input{Key: KeyPageDown}, 2)

	selected, ok := model.Selected()
	if !ok {
		t.Fatal("Selected() = false, want true")
	}
	if selected.ID != "4" {
		t.Fatalf("Selected().ID = %q, want %q", selected.ID, "4")
	}

	model.Update(Input{Key: KeyPageUp}, 2)
	selected, _ = model.Selected()
	if selected.ID != "2" {
		t.Fatalf("Selected().ID after page up = %q, want %q", selected.ID, "2")
	}
}

func TestModelEditsFilter(t *testing.T) {
	t.Parallel()

	model := New([]Item{
		{ID: "1", Label: "feature-one"},
		{ID: "2", Label: "feature-two"},
	}, "")

	model.Update(Input{Key: KeyRune, Rune: 't'}, 5)
	model.Update(Input{Key: KeyRune, Rune: 'w'}, 5)

	if got := model.Filter(); got != "tw" {
		t.Fatalf("Filter() = %q, want %q", got, "tw")
	}

	model.Update(Input{Key: KeyBackspace}, 5)
	if got := model.Filter(); got != "t" {
		t.Fatalf("Filter() after backspace = %q, want %q", got, "t")
	}
}
