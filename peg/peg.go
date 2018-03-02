// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package peg

import "unicode/utf8"

// A Node is a node in a Peggy parse tree.
type Node struct {
	// Name is the name of the Rule associated with the node,
	// or the empty string for anonymous Nodes
	// that are not associated with any Rule.
	Name string

	// Text is the input text of the Node's subtree.
	Text string

	// Kids are the immediate successors of this node.
	Kids []*Node
}

// A Fail is a node in a failed-parse tree.
// A failed-parse tree contains all paths in a failed parse
// that lead to the furthest error location in the input text.
// There are two types of nodes: named and unnamed.
// Named nodes represent grammar rules that failed to parse.
// Unnamed nodes represent terminal expressions that failed to parse.
type Fail struct {
	// Name is the name of the Rule associated with the node,
	// or the empty string if the Fail is a terminal expression failure.
	Name string

	// Pos is the byte offset into the input of the Fail.
	Pos int

	// Kids are the immediate succors of this Fail.
	// Kids is only non-nil for named Fail nodes.
	Kids []*Fail

	// Want is a string describing what was expected at the error position.
	// It is only non-empty for unnamed Fail nodes.
	//
	// It can be of one of the following forms:
	// 	"…" indicating a failed literal match, where the text between the quotes is the expected literal using Go escaping.
	// 	. indicating a failed . match.
	// 	[…] indicating a failed character class match, where the text between the [ and ] is the character class.
	// 	!… where the text after ! is the string representation of a failed predicate subexpression.
	// 	&… where the text after & is the string representation of a failed predicate subexpression.
	// 	… the error-name of a rule.
	// 		For example, "int" in rule: Integer "int" <- [0-9].
	Want string
}

// DecodeRuneInString is utf8.DecodeRuneInString.
// It's here so parsers can just include peg, and not also need unicode/utf8.
func DecodeRuneInString(s string) (rune, int) {
	return utf8.DecodeRuneInString(s)
}
