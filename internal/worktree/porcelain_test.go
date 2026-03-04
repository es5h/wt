package worktree

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsePorcelain(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []Worktree
		wantErr bool
	}{
		{
			name: "UnknownKeyAfterFirstWorktree",
			in: strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
unknown key
`),
			want: []Worktree{
				{Path: "/repo", HEAD: "0123456789abcdef0123456789abcdef01234567"},
			},
		},
		{
			name: "UnknownKeyBeforeFirstWorktree",
			in: strings.TrimSpace(`
unknown key
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
`),
			wantErr: true,
		},
		{
			name: "WithCRLF",
			in:   strings.TrimSpace("worktree /repo\r\nHEAD 0123456789abcdef0123456789abcdef01234567\r\nbranch refs/heads/main"),
			want: []Worktree{
				{Path: "/repo", HEAD: "0123456789abcdef0123456789abcdef01234567", Branch: "refs/heads/main"},
			},
		},
		{
			name: "FinalizesAtEOFWithoutTrailingNewline",
			in:   "worktree /repo\nHEAD 0123456789abcdef0123456789abcdef01234567\nbranch refs/heads/main",
			want: []Worktree{
				{Path: "/repo", HEAD: "0123456789abcdef0123456789abcdef01234567", Branch: "refs/heads/main"},
			},
		},
		{
			name: "TwoRecords",
			in: strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main

worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
detached
locked reason: manually locked
prunable reason: missing .git file
`),
			want: []Worktree{
				{Path: "/repo", HEAD: "0123456789abcdef0123456789abcdef01234567", Branch: "refs/heads/main"},
				{
					Path:        "/repo/.wt/feature-x",
					HEAD:        "abcdefabcdefabcdefabcdefabcdefabcdefabcd",
					Detached:    true,
					Locked:      true,
					LockReason:  "reason: manually locked",
					Prunable:    true,
					PruneReason: "reason: missing .git file",
				},
			},
		},
		{
			name: "PrunableNoReason",
			in: strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
prunable
`),
			want: []Worktree{
				{Path: "/repo", HEAD: "0123456789abcdef0123456789abcdef01234567", Prunable: true},
			},
		},
		{
			name: "ExtraSpacesInKeyValue",
			in: strings.TrimSpace(`
worktree   /repo
HEAD   0123456789abcdef0123456789abcdef01234567
branch    refs/heads/main
locked    because i said so
`),
			want: []Worktree{
				{
					Path:       "/repo",
					HEAD:       "0123456789abcdef0123456789abcdef01234567",
					Branch:     "refs/heads/main",
					Locked:     true,
					LockReason: "because i said so",
				},
			},
		},
		{
			name: "LockedNoReason",
			in: strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
locked
`),
			want: []Worktree{
				{Path: "/repo", HEAD: "0123456789abcdef0123456789abcdef01234567", Locked: true},
			},
		},
		{
			name: "Error_WhenNoWorktreeYet",
			in: strings.TrimSpace(`
HEAD 0123456789abcdef0123456789abcdef01234567
worktree /repo
`),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParsePorcelain(strings.NewReader(tc.in))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParsePorcelain() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePorcelain() error = %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParsePorcelain() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
