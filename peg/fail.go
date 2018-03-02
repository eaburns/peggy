// Copyright 2018 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package peg

import "fmt"

// SimpleError returns an error with a basic error message
// that describes what was expected at all of the leaf fails
// with the greatest position in the tree.
//
// The FilePath field of the returned Error is the empty string.
// The caller can set this field if to prefix the location
// with the path to an input file.
func SimpleError(text string, node *Fail) Error {
	leaves := LeafFails(node)

	var want string
	for i, l := range leaves {
		switch {
		case i == len(leaves)-1 && i == 1:
			want += " or "
		case i == len(leaves)-1 && len(want) > 1:
			want += ", or "
		case i > 0:
			want += ", "
		}
		want += l.Want
	}

	got := "EOF"
	pos := leaves[0].Pos
	if pos < len(text) {
		end := pos + 10
		if end > len(text) {
			end = len(text)
		}
		got = "'" + text[pos:end] + "'"
	}

	return Error{
		Loc:     Location(text, pos),
		Message: fmt.Sprintf("want %s; got %s", want, got),
	}
}

// Error implements error, prefixing an error message
// with location information for the error.
type Error struct {
	// FilePath is the path of the input file containing the error.
	FilePath string
	// Loc is the location of the error.
	Loc Loc
	// Message is the error message.
	Message string
}

func (err Error) Error() string {
	return fmt.Sprintf("%s:%d.%d: %s",
		err.FilePath, err.Loc.Line, err.Loc.Column, err.Message)
}

// LeafFails returns all fails in the tree with the greatest Pos.
func LeafFails(node *Fail) []*Fail {
	pos := -1
	var fails []*Fail
	seen := make(map[*Fail]bool)
	var walk func(*Fail)
	walk = func(n *Fail) {
		if seen[n] {
			return
		}
		seen[n] = true
		if len(n.Kids) == 0 {
			switch {
			case n.Pos > pos:
				pos = n.Pos
				fails = append(fails[:0], n)
			case n.Pos == pos:
				fails = append(fails, n)
			}
			return
		}
		for _, k := range n.Kids {
			walk(k)
		}
	}
	walk(node)
	return fails
}

// DedupFails removes duplicate fail branches from the tree,
// keeping only the first occurrence of each.
// This is useful for example before printing the Fail tree,
// because the non-deduped Fail tree can be exponential
// in the input size.
func DedupFails(node *Fail) {
	seen := make(map[*Fail]bool)
	var walk func(*Fail) bool
	walk = func(n *Fail) bool {
		if seen[n] {
			return false
		}
		seen[n] = true
		var kids []*Fail
		for _, k := range n.Kids {
			if walk(k) {
				kids = append(kids, k)
			}
		}
		n.Kids = kids
		return true
	}
	walk(node)
}
