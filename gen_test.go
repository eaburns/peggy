// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/eaburns/peggy/peg"
	"github.com/eaburns/pretty"
)

var binaries map[string]string

type test struct {
	Name    string
	Grammar string
	Input   string
	Pos     int
	Node    *peg.Node
	Fail    *peg.Fail
}

var testCases = []test{
	/*
		// BUG: We shouldn't allow unused labels.
		// This would be caught if we did a type check
		// before gofmt on the output.
		{
			Name:    "label unused match",
			Grammar: "A <- L:'abc'",
			Input:   "abc",
		},
	*/
	{
		// "start" is an internal identifier name. There should be no conflict.
		Name:    "label name conflicts with parser internal variable",
		Grammar: `A <- start:'abc' &{ start == "abc" } 'xyz'`,
		Input:   "abcxyz",
		Pos:     len("abcxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcxyz",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label pred expr",
		Grammar: `A <- L:&'abc' &{L == ""} "abc"`,
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{Text: "abc"},
			},
		},
	},
	{
		Name:    "label rep expr none",
		Grammar: `A <- L:'abc'* &{L == ""} 'xyz'`,
		Input:   "xyz",
		Pos:     len("xyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "xyz",
			Kids: []*peg.Node{
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label rep expr one",
		Grammar: `A <- L:'abc'* &{L == "abc"} 'xyz'`,
		Input:   "abcxyz",
		Pos:     len("abcxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcxyz",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label rep expr many",
		Grammar: `A <- L:'abc'* &{L == "abcabcabc"} 'xyz'`,
		Input:   "abcabcabcxyz",
		Pos:     len("abcabcabcxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcabcabcxyz",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "abc"},
				{Text: "abc"},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label opt expr empty",
		Grammar: `A <- L:'abc'? &{L == ""} 'xyz'`,
		Input:   "xyz",
		Pos:     len("xyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "xyz",
			Kids: []*peg.Node{
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label opt expr non-empty",
		Grammar: `A <- L:'abc'? &{L == "abc"} 'xyz'`,
		Input:   "abcxyz",
		Pos:     len("abcxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcxyz",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label ident",
		Grammar: "A <- L:B &{L == `abc`} 'xyz'\nB <- 'abc'",
		Input:   "abcxyz",
		Pos:     len("abcxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcxyz",
			Kids: []*peg.Node{
				{
					Name: "B",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label subexpr",
		Grammar: "A <- L:('123' / 'abc') &{L == `abc`} 'xyz'",
		Input:   "abcxyz",
		Pos:     len("abcxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcxyz",
			Kids: []*peg.Node{
				{
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label predcode",
		Grammar: "A <- L:&{true} &{L == ``} 'xyz'",
		Input:   "xyz",
		Pos:     len("xyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "xyz",
			Kids: []*peg.Node{
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label literal",
		Grammar: "A <- L:'abc' &{L == `abc`} 'xyz'",
		Input:   "abcxyz",
		Pos:     len("abcxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcxyz",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label charclass",
		Grammar: "A <- L:[a-z] &{L == `n`} 'xyz'",
		Input:   "nxyz",
		Pos:     len("nxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "nxyz",
			Kids: []*peg.Node{
				{Text: "n"},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label any",
		Grammar: "A <- L:. &{L == `α`} 'xyz'",
		Input:   "αxyz",
		Pos:     len("αxyz"),
		Node: &peg.Node{
			Name: "A",
			Text: "αxyz",
			Kids: []*peg.Node{
				{Text: "α"},
				{Text: "xyz"},
			},
		},
	},
	{
		Name:    "label multiple",
		Grammar: "A <- one:. two:. three:. &{one == `1` && two == `2` && three == `3`}",
		Input:   "123",
		Pos:     len("123"),
		Node: &peg.Node{
			Name: "A",
			Text: "123",
			Kids: []*peg.Node{
				{Text: "1"},
				{Text: "2"},
				{Text: "3"},
			},
		},
	},
	{
		Name:    "nested labels",
		Grammar: "A <- abc:(ab:(a:'a' 'b') 'c') &{abc == `abc` && ab == `ab` && a == `a`}",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{
					Text: "abc",
					Kids: []*peg.Node{
						{
							Text: "ab",
							Kids: []*peg.Node{
								{Text: "a"},
								{Text: "b"},
							},
						},
						{Text: "c"},
					},
				},
			},
		},
	},
	{
		Name:    "predcode with label mismatch",
		Grammar: `A <- L:'abc'* &{L == ""} !.`,
		Input:   "abc",
		Pos:     len("abc"),
		Fail: &peg.Fail{
			Name: "A",
			Pos:  0,
			Kids: []*peg.Fail{
				{
					Pos:  len("abc"),
					Want: `"abc"`,
				},
				{
					Pos:  len("abc"),
					Want: `&{L == ""}`,
				},
			},
		},
	},
	{
		Name:    "predcode match",
		Grammar: "A <- &{ true }",
		Input:   "☺☹",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "predcode mismatch",
		Grammar: "A <- &{ false }",
		Input:   "☺☹",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Pos:  0,
			Kids: []*peg.Fail{
				{
					Pos:  0,
					Want: "&{ false }",
				},
			},
		},
	},
	{
		Name:    "neg predcode match",
		Grammar: "A <- !{ false }",
		Input:   "☺☹",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "neg predcode mismatch",
		Grammar: "A <- !{ true }",
		Input:   "☺☹",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Pos:  0,
			Kids: []*peg.Fail{
				{
					Pos:  0,
					Want: "!{ true }",
				},
			},
		},
	},
	{
		Name:    "literal match",
		Grammar: "A <- 'abc'",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{{Text: "abc"}},
		},
	},
	{
		Name:    "literal match non-ASCII",
		Grammar: "A <- 'αβξ'",
		Input:   "αβξ",
		Pos:     len("αβξ"),
		Node: &peg.Node{
			Name: "A",
			Text: "αβξ",
			Kids: []*peg.Node{{Text: "αβξ"}},
		},
	},
	{
		Name:    "literal mismatch",
		Grammar: "A <- 'abc'",
		Input:   "abz",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `"abc"`},
			},
		},
	},
	{
		Name:    "any match",
		Grammar: "A <- .",
		Input:   "abc",
		Pos:     len("a"),
		Node: &peg.Node{
			Name: "A",
			Text: "a",
			Kids: []*peg.Node{{Text: "a"}},
		},
	},
	{
		Name:    "any match non-ASCII",
		Grammar: "A <- .",
		Input:   "αβξ",
		Pos:     len("α"),
		Node: &peg.Node{
			Name: "A",
			Text: "α",
			Kids: []*peg.Node{{Text: "α"}},
		},
	},
	{
		Name:    "any mismatch",
		Grammar: "A <- .",
		Input:   "",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `.`},
			},
		},
	},
	{
		Name:    "charclass match rune",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "a",
		Pos:     len("a"),
		Node: &peg.Node{
			Name: "A",
			Text: "a",
			Kids: []*peg.Node{{Text: "a"}},
		},
	},
	{
		Name:    "charclass match range",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "B",
		Pos:     len("B"),
		Node: &peg.Node{
			Name: "A",
			Text: "B",
			Kids: []*peg.Node{{Text: "B"}},
		},
	},
	{
		Name:    "charclass match non-ASCII rune",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "☺",
		Pos:     len("☺"),
		Node: &peg.Node{
			Name: "A",
			Text: "☺",
			Kids: []*peg.Node{{Text: "☺"}},
		},
	},
	{
		Name:    "charclass match non-ASCII range",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "β",
		Pos:     len("β"),
		Node: &peg.Node{
			Name: "A",
			Text: "β",
			Kids: []*peg.Node{{Text: "β"}},
		},
	},
	{
		Name:    "charclass mismatch rune",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "z",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "charclass mismatch before range",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "@", // just before A
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "charclass mismatch after range",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "D", // just after C
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "charclass mismatch non-ASCII rune",
		Grammar: "A <- [abcA-C☹☺α-ξ]",
		Input:   "·",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass match rune",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "z",
		Pos:     len("z"),
		Node: &peg.Node{
			Name: "A",
			Text: "z",
			Kids: []*peg.Node{{Text: "z"}},
		},
	},
	{
		Name:    "neg charclass match before range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "@", // just before A
		Pos:     len("@"),
		Node: &peg.Node{
			Name: "A",
			Text: "@",
			Kids: []*peg.Node{{Text: "@"}},
		},
	},
	{
		Name:    "neg charclass match after range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "D", // just after C
		Pos:     len("D"),
		Node: &peg.Node{
			Name: "A",
			Text: "D",
			Kids: []*peg.Node{{Text: "D"}},
		},
	},
	{
		Name:    "neg charclass match non-ASCII rune",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "·",
		Pos:     len("·"),
		Node: &peg.Node{
			Name: "A",
			Text: "·",
			Kids: []*peg.Node{{Text: "·"}},
		},
	},
	{
		Name:    "neg charclass mismatch rune",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "a",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass mismatch begin range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "A",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass mismatch mid range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "B",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass mismatch end range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "C",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass mismatch non-ASCII rune",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "☺",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass mismatch begin non-ASCII range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "α",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass mismatch mid non-ASCII range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "β",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "neg charclass mismatch end non-ASCII range",
		Grammar: "A <- [^abcA-C☹☺α-ξ]",
		Input:   "ξ",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abcA-C☹☺α-ξ]`},
			},
		},
	},
	{
		Name:    "ident match",
		Grammar: "A <- B\nB <- 'abc'",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{
					Name: "B",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
		},
	},
	{
		Name:    "ident mismatch",
		Grammar: "A <- B\nB <- 'abc'",
		Input:   "abz",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Name: "B",
					Kids: []*peg.Fail{
						{Want: `"abc"`},
					},
				},
			},
		},
	},
	{
		Name:    "star match 0",
		Grammar: "A <- 'abc'*",
		Input:   "xyz",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "star match 0 EOF",
		Grammar: "A <- 'abc'*",
		Input:   "xyz",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "star match 1",
		Grammar: "A <- 'abc'*",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{{Text: "abc"}},
		},
	},
	{
		Name:    "star match >1",
		Grammar: "A <- 'abc'*",
		Input:   "abcabcabcxyz",
		Pos:     len("abcabcabc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcabcabc",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "abc"},
				{Text: "abc"},
			},
		},
	},
	{
		Name:    "star any",
		Grammar: "A <- .*",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{Text: "a"},
				{Text: "b"},
				{Text: "c"},
			},
		},
	},
	{
		Name:    "star charclass",
		Grammar: "A <- [abc]*",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{Text: "a"},
				{Text: "b"},
				{Text: "c"},
			},
		},
	},
	{
		Name:    "star neg charclass",
		Grammar: "A <- [^abc]*",
		Input:   "XYZ",
		Pos:     len("XYZ"),
		Node: &peg.Node{
			Name: "A",
			Text: "XYZ",
			Kids: []*peg.Node{
				{Text: "X"},
				{Text: "Y"},
				{Text: "Z"},
			},
		},
	},
	{
		Name:    "star ident",
		Grammar: "A <- B*\nB <- 'abc'",
		Input:   "abcabc",
		Pos:     len("abcabc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcabc",
			Kids: []*peg.Node{
				{
					Name: "B",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
				{
					Name: "B",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
		},
	},
	{
		Name:    "star subexpr",
		Grammar: "A <- ('a' 'b' 'c')*",
		Input:   "abcabc",
		Pos:     len("abcabc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcabc",
			Kids: []*peg.Node{
				{
					Text: "abc",
					Kids: []*peg.Node{
						{Text: "a"},
						{Text: "b"},
						{Text: "c"},
					},
				},
				{
					Text: "abc",
					Kids: []*peg.Node{
						{Text: "a"},
						{Text: "b"},
						{Text: "c"},
					},
				},
			},
		},
	},
	{
		Name:    "plus match 1",
		Grammar: "A <- 'abc'+",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{{Text: "abc"}},
		},
	},
	{
		Name:    "plus match >1",
		Grammar: "A <- 'abc'+",
		Input:   "abcabcabcxyz",
		Pos:     len("abcabcabc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcabcabc",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "abc"},
				{Text: "abc"},
			},
		},
	},
	{
		Name:    "plus mismatch",
		Grammar: "A <- 'abc'+",
		Input:   "xyz",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `"abc"`},
			},
		},
	},
	{
		Name:    "plus mismatch EOF",
		Grammar: "A <- 'abc'+",
		Input:   "",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `"abc"`},
			},
		},
	},
	{
		Name:    "plus any match",
		Grammar: "A <- .+",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{Text: "a"},
				{Text: "b"},
				{Text: "c"},
			},
		},
	},
	{
		Name:    "plus any mismatch",
		Grammar: "A <- .+",
		Input:   "",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `.`},
			},
		},
	},
	{
		Name:    "plus charclass match",
		Grammar: "A <- [abc]+",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{Text: "a"},
				{Text: "b"},
				{Text: "c"},
			},
		},
	},
	{
		Name:    "plus charclass mismatch",
		Grammar: "A <- [abc]+",
		Input:   "XYZ",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[abc]`},
			},
		},
	},
	{
		Name:    "plus neg charclass match",
		Grammar: "A <- [^abc]+",
		Input:   "XYZ",
		Pos:     len("XYZ"),
		Node: &peg.Node{
			Name: "A",
			Text: "XYZ",
			Kids: []*peg.Node{
				{Text: "X"},
				{Text: "Y"},
				{Text: "Z"},
			},
		},
	},
	{
		Name:    "plus neg charclass mismatch",
		Grammar: "A <- [^abc]+",
		Input:   "abc",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `[^abc]`},
			},
		},
	},
	{
		Name:    "plus ident match",
		Grammar: "A <- B+\nB <- 'abc'",
		Input:   "abcabc",
		Pos:     len("abcabc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcabc",
			Kids: []*peg.Node{
				{
					Name: "B",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
				{
					Name: "B",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
		},
	},
	{
		Name:    "plus ident mismatch",
		Grammar: "A <- B+\nB <- 'abc'",
		Input:   "xyz",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Name: "B",
					Kids: []*peg.Fail{
						{Want: `"abc"`},
					},
				},
			},
		},
	},
	{
		Name:    "plus subexpr match",
		Grammar: "A <- ('a' 'b' 'c')+",
		Input:   "abcabc",
		Pos:     len("abcabc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcabc",
			Kids: []*peg.Node{
				{
					Text: "abc",
					Kids: []*peg.Node{
						{Text: "a"},
						{Text: "b"},
						{Text: "c"},
					},
				},
				{
					Text: "abc",
					Kids: []*peg.Node{
						{Text: "a"},
						{Text: "b"},
						{Text: "c"},
					},
				},
			},
		},
	},
	{
		Name:    "plus subexpr mismatch",
		Grammar: "A <- ('a' 'b' 'c')+",
		Input:   "xyz",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `"a"`},
			},
		},
	},
	{
		Name:    "question match 0",
		Grammar: "A <- 'abc'?",
		Input:   "xyz",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "question match 0 EOF",
		Grammar: "A <- 'abc'?",
		Input:   "xyz",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "question match 1",
		Grammar: "A <- 'abc'?",
		Input:   "abcabc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{{Text: "abc"}},
		},
	},
	{
		Name:    "question any",
		Grammar: "A <- .?",
		Input:   "a",
		Pos:     len("a"),
		Node: &peg.Node{
			Name: "A",
			Text: "a",
			Kids: []*peg.Node{{Text: "a"}},
		},
	},
	{
		Name:    "question charclass",
		Grammar: "A <- [abc]?",
		Input:   "a",
		Pos:     len("a"),
		Node: &peg.Node{
			Name: "A",
			Text: "a",
			Kids: []*peg.Node{{Text: "a"}},
		},
	},
	{
		Name:    "question neg charclass",
		Grammar: "A <- [^abc]?",
		Input:   "X",
		Pos:     len("X"),
		Node: &peg.Node{
			Name: "A",
			Text: "X",
			Kids: []*peg.Node{{Text: "X"}},
		},
	},
	{
		Name:    "question ident",
		Grammar: "A <- B?\nB <- 'abc'",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{
					Name: "B",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
		},
	},
	{
		Name:    "question match subexpr",
		Grammar: "A <- ('a' 'b' 'c')?",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{
				{
					Text: "abc",
					Kids: []*peg.Node{
						{Text: "a"},
						{Text: "b"},
						{Text: "c"},
					},
				},
			},
		},
	},
	{
		Name:    "pos pred match",
		Grammar: "A <- &'abc'",
		Input:   "abc",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "pos pred mismatch",
		Grammar: "A <- &'abc'",
		Input:   "xyz",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `&"abc"`},
			},
		},
	},
	{
		Name:    "pos pred any match",
		Grammar: "A <- &.",
		Input:   "a",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "pos pred any mismatch",
		Grammar: "A <- &.",
		Input:   "",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `&.`},
			},
		},
	},
	{
		Name:    "pos pred charclass match",
		Grammar: "A <- &[abc]",
		Input:   "a",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "pos pred charclass mismatch",
		Grammar: "A <- &[abc]",
		Input:   "X",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `&[abc]`},
			},
		},
	},
	{
		Name:    "pos pred neg charclass match",
		Grammar: "A <- &[^abc]",
		Input:   "X",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "pos pred neg charclass mismatch",
		Grammar: "A <- &[^abc]",
		Input:   "a",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `&[^abc]`},
			},
		},
	},
	{
		Name:    "neg pred match",
		Grammar: "A <- !'abc'",
		Input:   "xyz",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "neg pred mismatch",
		Grammar: "A <- !'abc'",
		Input:   "abc",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `!"abc"`},
			},
		},
	},
	{
		Name:    "neg pred any match",
		Grammar: "A <- !.",
		Input:   "",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "neg pred any mismatch",
		Grammar: "A <- !.",
		Input:   "a",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `!.`},
			},
		},
	},
	{
		Name:    "neg pred charclass match",
		Grammar: "A <- ![abc]",
		Input:   "x",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "neg pred charclass mismatch",
		Grammar: "A <- ![abc]",
		Input:   "a",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `![abc]`},
			},
		},
	},
	{
		Name:    "neg pred neg charclass match",
		Grammar: "A <- ![^abc]",
		Input:   "a",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "neg pred neg charclass mismatch",
		Grammar: "A <- ![^abc]",
		Input:   "x",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `![^abc]`},
			},
		},
	},
	{
		Name:    "neg pred literal match",
		Grammar: "A <- !B\nB <- 'abc'",
		Input:   "xyz",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "neg pred neg charclass mismatch",
		Grammar: "A <- !B\nB <- 'abc'",
		Input:   "abc",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: "!B"},
			},
		},
	},
	{
		Name:    "sequence match",
		Grammar: "A <- 'abc' 'def' 'ghi'",
		Input:   "abcdefghi",
		Pos:     len("abcdefghi"),
		Node: &peg.Node{
			Name: "A",
			Text: "abcdefghi",
			Kids: []*peg.Node{
				{Text: "abc"},
				{Text: "def"},
				{Text: "ghi"},
			},
		},
	},
	{
		Name:    "sequence mismatch first",
		Grammar: "A <- 'abc' 'def' 'ghi'",
		Input:   "XYZdefghi",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `"abc"`},
			},
		},
	},
	{
		Name:    "sequence mismatch mid",
		Grammar: "A <- 'abc' 'def' 'ghi'",
		Input:   "abcXYZghi",
		Pos:     len("abc"), // error after abc
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Pos:  len("abc"),
					Want: `"def"`,
				},
			},
		},
	},
	{
		Name:    "sequence mismatch last",
		Grammar: "A <- 'abc' 'def' 'ghi'",
		Input:   "abcdefXYZ",
		Pos:     len("abcdef"), // error after abcdef
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Pos:  len("abcdef"),
					Want: `"ghi"`,
				},
			},
		},
	},
	{
		Name:    "choice match first",
		Grammar: "A <- 'abc' / 'def' / 'ghi'",
		Input:   "abc",
		Pos:     len("abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "abc",
			Kids: []*peg.Node{{Text: "abc"}},
		},
	},
	{
		Name:    "choice match mid",
		Grammar: "A <- 'abc' / 'def' / 'ghi'",
		Input:   "def",
		Pos:     len("def"),
		Node: &peg.Node{
			Name: "A",
			Text: "def",
			Kids: []*peg.Node{{Text: "def"}},
		},
	},
	{
		Name:    "choice match last",
		Grammar: "A <- 'abc' / 'def' / 'ghi'",
		Input:   "ghi",
		Pos:     len("ghi"),
		Node: &peg.Node{
			Name: "A",
			Text: "ghi",
			Kids: []*peg.Node{{Text: "ghi"}},
		},
	},
	{
		Name:    "choice mismatch",
		Grammar: "A <- 'abc' / 'def' / 'ghi'",
		Input:   "XYZ",
		Pos:     0,
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `"abc"`},
				{Want: `"def"`},
				{Want: `"ghi"`},
			},
		},
	},
	{
		Name:    "choice that can't fail empty",
		Grammar: "A <- 'abc' / 'def'?",
		Input:   "XYZ",
		Pos:     0,
		Node: &peg.Node{
			Name: "A",
		},
	},
	{
		Name:    "choice that can't fail non-empty",
		Grammar: "A <- 'abc' / 'def'?",
		Input:   "def",
		Pos:     len("def"),
		Node: &peg.Node{
			Name: "A",
			Text: "def",
			Kids: []*peg.Node{{Text: "def"}},
		},
	},
	{
		Name:    "choice after sequence match first",
		Grammar: "A <- '123' ('abc'/ 'αβξ')",
		Input:   "123abc",
		Pos:     len("123abc"),
		Node: &peg.Node{
			Name: "A",
			Text: "123abc",
			Kids: []*peg.Node{
				{Text: "123"},
				{
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
		},
	},
	{
		Name:    "choice after sequence match second",
		Grammar: "A <- '123' ('abc'/ 'αβξ')",
		Input:   "123αβξ",
		Pos:     len("123αβξ"),
		Node: &peg.Node{
			Name: "A",
			Text: "123αβξ",
			Kids: []*peg.Node{
				{Text: "123"},
				{
					Text: "αβξ",
					Kids: []*peg.Node{{Text: "αβξ"}},
				},
			},
		},
	},
	{
		Name:    "choice after sequence mismatch",
		Grammar: "A <- '123' ('abc'/ 'αβξ')",
		Input:   "123XYZ",
		Pos:     len("123"), // error after 123
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Pos:  len("123"),
					Want: `"abc"`,
				},
				{
					Pos:  len("123"),
					Want: `"αβξ"`,
				},
			},
		},
	},
	{
		Name:    "rule memo success",
		Grammar: "A <- 'a' B 'c' / 'a' B 'd'\nB <- 'B'",
		Input:   "aBd",
		Pos:     len("aBd"),
		Node: &peg.Node{
			Name: "A",
			Text: "aBd",
			Kids: []*peg.Node{
				{Text: "a"},
				{
					Name: "B",
					Text: "B",
					Kids: []*peg.Node{{Text: "B"}},
				},
				{Text: "d"},
			},
		},
	},
	{
		Name:    "rule memo failure",
		Grammar: "A <- 'a' B 'c' / 'a' B 'd'\nB <- 'B'",
		Input:   "aAd",
		Pos:     len("a"), // error after a
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Name: "B",
					Pos:  len("a"),
					Kids: []*peg.Fail{
						{
							Pos:  len("a"),
							Want: `"B"`,
						},
					},
				},
				{
					Name: "B",
					Pos:  len("a"),
					Kids: []*peg.Fail{
						{
							Pos:  len("a"),
							Want: `"B"`,
						},
					},
				},
			},
		},
	},
	{
		Name:    "latest error",
		Grammar: "A <- B 'x'\nB <- 'abc' 'def' / .",
		Input:   "abcxyz",
		Pos:     len("abc"), // latest error is after abc
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Name: "B",
					Pos:  0,
					Kids: []*peg.Fail{
						{
							Pos:  len("abc"),
							Want: `"def"`,
						},
					},
				},
			},
		},
	},
	{
		// Don't report the location of fails in silent exprs, & and !.
		Name:    "ignore silent fails",
		Grammar: "A <- !B 'xyz'\nB <- 'abc' 'def'",
		Input:   "abc",
		Pos:     0, // latest error is just before 'xyz', not after 'abc'
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{Want: `"xyz"`},
			},
		},
	},
	{
		// If an expr first fails in a silent expr, & or !,
		// we still report it's fail position if it fails
		// subsequently in a non-silent context.
		// Note that this is different from the behavior
		// of some other PEG parsers, which don't emit errors
		// if the cached value failed in a silent context.
		Name:    "no cache silent fails",
		Grammar: "A <- &B 'f' / B\nB <- 'a' 'b' 'c' 'd' 'e'",
		Input:   "abce",
		// The error is the missing 'd' between 'abc' and 'e'.
		// Some other PEG parsers would report the error at 0,
		// because the first time 'd' fails, it's silent, that's cached
		// and the subsequent fail uses the cached,
		// un-reported error.
		Pos: len("abc"),
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Name: "B",
					Kids: []*peg.Fail{
						{
							Pos:  len("abc"),
							Want: `"d"`,
						},
					},
				},
			},
		},
	},
	{
		Name:    "named rule fail",
		Grammar: "A <- 'abc' B 'def'\nB 'name' <- C\nC <- D\nD <- '123'",
		Input:   "abc124",
		Pos:     len("abc"),
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Name: "B",
					Pos:  len("abc"),
					Want: "name",
				},
			},
		},
	},
	{
		Name:    "ignore errors under successful named rules",
		Grammar: "A <- 'abc' B 'def'\nB 'name' <- '1' '2' '3' / .",
		// B fails after 12, backtracks and succeeds after the 1.
		// We should not report the error after abc12, but after abc1.
		Input: "abc12x",
		Pos:   len("abc1"),
		Fail: &peg.Fail{
			Name: "A",
			Kids: []*peg.Fail{
				{
					Pos:  len("abc1"),
					Want: `"def"`,
				},
			},
		},
	},
}

func TestGen(t *testing.T) {
	t.Parallel()
	for _, test := range testCases {
		test := test // for goroutine closure
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			binary := binaries[test.Grammar]
			var result struct {
				Pos  int
				Perr int
				Node *peg.Node
				Fail *peg.Fail
			}
			parse(binary, test.Input, &result)
			pos := result.Pos
			if result.Fail != nil {
				pos = result.Perr
			}
			t.Logf("result: %+v\n", result)
			if pos != test.Pos {
				t.Errorf("parse(%q)=%d, want %d", test.Grammar, pos, test.Pos)
			}
			var got interface{}
			if result.Node != nil {
				got = result.Node
			} else {
				got = result.Fail
			}
			var want interface{}
			if test.Node != nil {
				want = test.Node
			} else {
				want = test.Fail
			}
			if !reflect.DeepEqual(want, got) {
				t.Errorf("parse(%q)=\n%s\nwant\n%s",
					test.Grammar, pretty.String(got), pretty.String(want))
			}
		})
	}
}

func TestMain(m *testing.M) {
	binaries = buildAll(testCases)
	r := m.Run()
	rmAll(binaries)
	os.Exit(r)
}

// buildAll compiles the parser binaries for every grammar in tests
// and returns a map from the Peggy grammar string to the binary path.
func buildAll(tests []test) map[string]string {
	grammars := make(map[string]bool)
	for _, t := range tests {
		if g := t.Grammar; g != "" {
			grammars[g] = true
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(grammars))
	var mu sync.Mutex
	binaries := make(map[string]string, len(grammars))
	for grammar := range grammars {
		go func(grammar string) {
			defer wg.Done()
			source := genTest(prelude, grammar)
			binary := build(source)
			go rm(source)
			mu.Lock()
			defer mu.Unlock()
			binaries[grammar] = binary
		}(grammar)
	}
	wg.Wait()
	return binaries
}

// genTest generates Go source code for a Peggy
func genTest(prelude string, input string) string {
	f, err := ioutil.TempFile(os.TempDir(), "peggy_test")
	if err != nil {
		panic(err.Error())
	}
	input = prelude + input
	g, err := Parse(strings.NewReader(input), "")
	if err != nil {
		fmt.Printf("%s\n", input)
		panic(err.Error())
	}
	if err := Check(g); err != nil {
		fmt.Printf("%s\n", input)
		panic(err.Error())
	}
	if _, err := io.WriteString(f, "/*\n"+String(g.Rules)+"\n*/\n"); err != nil {
		panic(err.Error())
	}
	if err := Generate(f, "", g); err != nil {
		panic(err.Error())
	}
	fileName := f.Name()
	if err := f.Close(); err != nil {
		panic(err.Error())
	}
	goName := fileName + ".go"
	if err := os.Rename(fileName, goName); err != nil {
		panic(err.Error())
	}
	return goName
}

// build compiles a Go source and returns the path to the binary.
func build(source string) string {
	cmd := exec.Command("go", "build", source)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		c := cmd.Path + " " + strings.Join(cmd.Args[1:], " ")
		panic("failed to run [" + c + "]: " + err.Error())
	}
	return "./" + filepath.Base(strings.TrimSuffix(source, ".go"))
}

func rmAll(binaries map[string]string) {
	for _, binary := range binaries {
		rm(binary)
	}
}

func rm(file string) {
	if err := os.Remove(file); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove %s: %s", file, err)
	}
}

// parse parses an input using the given binary
// and returns the position of either the parse or error
// along with whether the parse succeeded.
func parse(binary, input string, result interface{}) {
	cmd := exec.Command(binary)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err.Error())
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err.Error())
	}
	if err := cmd.Start(); err != nil {
		panic(err.Error())
	}
	go func() {
		if _, err := io.WriteString(stdin, input); err != nil {
			panic(err.Error())
		}
		if err := stdin.Close(); err != nil {
			panic(err.Error())
		}
	}()
	if err := gob.NewDecoder(stdout).Decode(result); err != nil {
		panic(err.Error())
	}
	if err := cmd.Wait(); err != nil {
		panic(err.Error())
	}
}

var prelude = `{
package main

import (
	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/eaburns/peggy/peg"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	p := _NewParser(string(data))
	var result struct {
		Pos int
		Perr int
		Node       *peg.Node
		Fail       *peg.Fail
	}
	if result.Pos, result.Perr = _AAccepts(p, 0); result.Pos >= 0 {
		_, result.Node = _ANode(p, 0)
	} else {
		_, result.Fail = _AFail(p, 0, result.Perr)
	}
	if err := gob.NewEncoder(os.Stdout).Encode(&result); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
}
`
