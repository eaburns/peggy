// Copyright 2018 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package peg

import "unicode/utf8"

// A Loc is a location in the input text.
type Loc struct {
	Byte   int
	Rune   int
	Line   int
	Column int
}

// Location returns the Loc at the corresponding byte offset in the text.
func Location(text string, byte int) Loc {
	var loc Loc
	loc.Line = 1
	loc.Column = 1
	for byte > loc.Byte {
		r, w := utf8.DecodeRuneInString(text[loc.Byte:])
		loc.Byte += w
		loc.Rune++
		loc.Column++
		if r == '\n' {
			loc.Line++
			loc.Column = 1
		}
	}
	return loc
}
