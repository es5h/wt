package picker

import "testing"

func TestFitLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		width int
		want  string
	}{
		{
			name:  "no truncation",
			value: "hello",
			width: 8,
			want:  "hello",
		},
		{
			name:  "truncation with ellipsis",
			value: "0123456789",
			width: 7,
			want:  "0123...",
		},
		{
			name:  "small width without ellipsis",
			value: "0123456789",
			width: 3,
			want:  "012",
		},
		{
			name:  "zero width",
			value: "hello",
			width: 0,
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := fitLine(tc.value, tc.width); got != tc.want {
				t.Fatalf("fitLine(%q, %d) = %q, want %q", tc.value, tc.width, got, tc.want)
			}
		})
	}
}

func TestFitPrefixedLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prefix  string
		content string
		width   int
		want    string
	}{
		{
			name:    "fits with prefix",
			prefix:  "> ",
			content: "feature",
			width:   12,
			want:    "> feature",
		},
		{
			name:    "content truncates",
			prefix:  "> ",
			content: "0123456789",
			width:   8,
			want:    "> 012...",
		},
		{
			name:    "prefix wider than width",
			prefix:  ">>>>",
			content: "ignored",
			width:   3,
			want:    ">>>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := fitPrefixedLine(tc.prefix, tc.content, tc.width); got != tc.want {
				t.Fatalf("fitPrefixedLine(%q, %q, %d) = %q, want %q", tc.prefix, tc.content, tc.width, got, tc.want)
			}
		})
	}
}
