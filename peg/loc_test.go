// Copyright 2018 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package peg

import (
	"strings"
	"testing"
)

func TestLocation(t *testing.T) {
	tests := []struct {
		in   string
		want Loc
	}{
		{
			in:   "*",
			want: Loc{Byte: 0, Rune: 0, Line: 1, Column: 1},
		},
		{
			in:   "abc*",
			want: Loc{Byte: 3, Rune: 3, Line: 1, Column: 4},
		},
		{
			in:   "ab\n*",
			want: Loc{Byte: 3, Rune: 3, Line: 2, Column: 1},
		},
		{
			in:   "ab\n*",
			want: Loc{Byte: 3, Rune: 3, Line: 2, Column: 1},
		},
		{
			in:   "ab\nabc\nxyz*",
			want: Loc{Byte: 10, Rune: 10, Line: 3, Column: 4},
		},
		{
			in:   "☺*",
			want: Loc{Byte: len("☺"), Rune: 1, Line: 1, Column: 2},
		},
		{
			in:   "☺☺☺*",
			want: Loc{Byte: 3 * len("☺"), Rune: 3, Line: 1, Column: 4},
		},
		{
			in:   "☺☺\n☺*",
			want: Loc{Byte: 3*len("☺") + 1, Rune: 4, Line: 2, Column: 2},
		},
		{
			in:   "☺☺\n☺*☹☹☹",
			want: Loc{Byte: 3*len("☺") + 1, Rune: 4, Line: 2, Column: 2},
		},
	}
	for _, test := range tests {
		b := strings.Index(test.in, "*")
		if b < 0 {
			panic("no *")
		}
		got := Location(test.in, b)
		if got != test.want {
			t.Errorf("Location(%q, %d)=%v, want %v", test.in, b, got, test.want)
		}
	}
}
