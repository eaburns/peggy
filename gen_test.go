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
	"testing"

	"github.com/eaburns/peggy/peg"
	"github.com/eaburns/pretty"
)

type genTest struct {
	grammar string
	cases   []genTestCase
}

type genTestCase struct {
	name  string
	input string
	pos   int
	node  *peg.Node
	fail  *peg.Fail
}

// TODO: add the bug case.
var genTests = []genTest{
	{
		// "start" is an internal identifier name. There should be no conflict.
		grammar: `A <- start:'abc' &{ start == "abc" } 'xyz'`,
		cases: []genTestCase{
			{
				name:  "label name conflicts with parser internal variable",
				input: "abcxyz",
				pos:   len("abcxyz"),
				node: &peg.Node{
					Name: "A",
					Text: "abcxyz",
					Kids: []*peg.Node{
						{Text: "abc"},
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: `A <- L:&'abc' &{L == ""} "abc"`,
		cases: []genTestCase{
			{
				name:  "label pred expr",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
					Text: "abc",
					Kids: []*peg.Node{
						{Text: "abc"},
					},
				},
			},
		},
	},
	{
		grammar: `A <- L:'abc'* &{L == ""} 'xyz'`,
		cases: []genTestCase{
			{
				name:  "label rep expr none",
				input: "xyz",
				pos:   len("xyz"),
				node: &peg.Node{
					Name: "A",
					Text: "xyz",
					Kids: []*peg.Node{
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: `A <- L:'abc'* &{L == "abc"} 'xyz'`,
		cases: []genTestCase{
			{
				name:  "label rep expr one",
				input: "abcxyz",
				pos:   len("abcxyz"),
				node: &peg.Node{
					Name: "A",
					Text: "abcxyz",
					Kids: []*peg.Node{
						{Text: "abc"},
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: `A <- L:'abc'* &{L == "abcabcabc"} 'xyz'`,
		cases: []genTestCase{
			{
				name:  "label rep expr many",
				input: "abcabcabcxyz",
				pos:   len("abcabcabcxyz"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: `A <- L:'abc'? &{L == ""} 'xyz'`,
		cases: []genTestCase{
			{
				name:  "label opt expr empty",
				input: "xyz",
				pos:   len("xyz"),
				node: &peg.Node{
					Name: "A",
					Text: "xyz",
					Kids: []*peg.Node{
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: `A <- L:'abc'? &{L == "abc"} 'xyz'`,
		cases: []genTestCase{
			{
				name:  "label opt expr non-empty",
				input: "abcxyz",
				pos:   len("abcxyz"),
				node: &peg.Node{
					Name: "A",
					Text: "abcxyz",
					Kids: []*peg.Node{
						{Text: "abc"},
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- L:B &{L == `abc`} 'xyz'\nB <- 'abc'",
		cases: []genTestCase{
			{
				name:  "label ident",
				input: "abcxyz",
				pos:   len("abcxyz"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: "A <- L:('123' / 'abc') &{L == `abc`} 'xyz'",
		cases: []genTestCase{
			{
				name:  "label subexpr",
				input: "abcxyz",
				pos:   len("abcxyz"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: "A <- L:&{true} &{L == ``} 'xyz'",
		cases: []genTestCase{
			{
				name:  "label predcode",
				input: "xyz",
				pos:   len("xyz"),
				node: &peg.Node{
					Name: "A",
					Text: "xyz",
					Kids: []*peg.Node{
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- L:'abc' &{L == `abc`} 'xyz'",
		cases: []genTestCase{
			{
				name:  "label literal",
				input: "abcxyz",
				pos:   len("abcxyz"),
				node: &peg.Node{
					Name: "A",
					Text: "abcxyz",
					Kids: []*peg.Node{
						{Text: "abc"},
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- L:[a-z] &{L == `n`} 'xyz'",
		cases: []genTestCase{
			{
				name:  "label charclass",
				input: "nxyz",
				pos:   len("nxyz"),
				node: &peg.Node{
					Name: "A",
					Text: "nxyz",
					Kids: []*peg.Node{
						{Text: "n"},
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- L:. &{L == `α`} 'xyz'",
		cases: []genTestCase{
			{
				name:  "label any",
				input: "αxyz",
				pos:   len("αxyz"),
				node: &peg.Node{
					Name: "A",
					Text: "αxyz",
					Kids: []*peg.Node{
						{Text: "α"},
						{Text: "xyz"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- one:. two:. three:. &{one == `1` && two == `2` && three == `3`}",
		cases: []genTestCase{
			{
				name:  "label multiple",
				input: "123",
				pos:   len("123"),
				node: &peg.Node{
					Name: "A",
					Text: "123",
					Kids: []*peg.Node{
						{Text: "1"},
						{Text: "2"},
						{Text: "3"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- abc:(ab:(a:'a' 'b') 'c') &{abc == `abc` && ab == `ab` && a == `a`}",
		cases: []genTestCase{
			{
				name:  "nested labels",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: `A <- L:'abc'* &{L == ""} !.`,
		cases: []genTestCase{
			{
				name:  "predcode with label mismatch",
				input: "abc",
				pos:   len("abc"),
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- &{ true }",
		cases: []genTestCase{
			{
				name:  "predcode match",
				input: "☺☹",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
		},
	},
	{
		grammar: "A <- &{ false }",
		cases: []genTestCase{
			{
				name:  "predcode mismatch",
				input: "☺☹",
				pos:   0,
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- !{ false }",
		cases: []genTestCase{
			{
				name:  "neg predcode match",
				input: "☺☹",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
		},
	},
	{
		grammar: "A <- !{ true }",
		cases: []genTestCase{
			{
				name:  "neg predcode mismatch",
				input: "☺☹",
				pos:   0,
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- 'abc'",
		cases: []genTestCase{
			{
				name:  "literal match",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
			{
				name:  "literal mismatch",
				input: "abz",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `"abc"`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- 'αβξ'",
		cases: []genTestCase{
			{
				name:  "literal match non-ASCII",
				input: "αβξ",
				pos:   len("αβξ"),
				node: &peg.Node{
					Name: "A",
					Text: "αβξ",
					Kids: []*peg.Node{{Text: "αβξ"}},
				},
			},
		},
	},
	{
		grammar: "A <- .",
		cases: []genTestCase{
			{
				name:  "any match",
				input: "abc",
				pos:   len("a"),
				node: &peg.Node{
					Name: "A",
					Text: "a",
					Kids: []*peg.Node{{Text: "a"}},
				},
			},
			{
				name:  "any match non-ASCII",
				input: "αβξ",
				pos:   len("α"),
				node: &peg.Node{
					Name: "A",
					Text: "α",
					Kids: []*peg.Node{{Text: "α"}},
				},
			},
			{
				name:  "any mismatch",
				input: "",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `.`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- [abcA-C☹☺α-ξ]",
		cases: []genTestCase{
			{
				name:  "charclass match rune",
				input: "a",
				pos:   len("a"),
				node: &peg.Node{
					Name: "A",
					Text: "a",
					Kids: []*peg.Node{{Text: "a"}},
				},
			},
			{
				name:  "charclass match range",
				input: "B",
				pos:   len("B"),
				node: &peg.Node{
					Name: "A",
					Text: "B",
					Kids: []*peg.Node{{Text: "B"}},
				},
			},
			{
				name:  "charclass match non-ASCII rune",
				input: "☺",
				pos:   len("☺"),
				node: &peg.Node{
					Name: "A",
					Text: "☺",
					Kids: []*peg.Node{{Text: "☺"}},
				},
			},
			{
				name:  "charclass match non-ASCII range",
				input: "β",
				pos:   len("β"),
				node: &peg.Node{
					Name: "A",
					Text: "β",
					Kids: []*peg.Node{{Text: "β"}},
				},
			},
			{
				name:  "charclass mismatch rune",
				input: "z",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "charclass mismatch before range",
				input: "@", // just before A
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "charclass mismatch after range",
				input: "D", // just after C
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "charclass mismatch non-ASCII rune",
				input: "·",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[abcA-C☹☺α-ξ]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- [^abcA-C☹☺α-ξ]",
		cases: []genTestCase{
			{
				name:  "neg charclass match rune",
				input: "z",
				pos:   len("z"),
				node: &peg.Node{
					Name: "A",
					Text: "z",
					Kids: []*peg.Node{{Text: "z"}},
				},
			},
			{
				name:  "neg charclass match before range",
				input: "@", // just before A
				pos:   len("@"),
				node: &peg.Node{
					Name: "A",
					Text: "@",
					Kids: []*peg.Node{{Text: "@"}},
				},
			},
			{
				name:  "neg charclass match after range",
				input: "D", // just after C
				pos:   len("D"),
				node: &peg.Node{
					Name: "A",
					Text: "D",
					Kids: []*peg.Node{{Text: "D"}},
				},
			},
			{
				name:  "neg charclass match non-ASCII rune",
				input: "·",
				pos:   len("·"),
				node: &peg.Node{
					Name: "A",
					Text: "·",
					Kids: []*peg.Node{{Text: "·"}},
				},
			},
			{
				name:  "neg charclass mismatch rune",
				input: "a",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "neg charclass mismatch begin range",
				input: "A",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "neg charclass mismatch mid range",
				input: "B",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "neg charclass mismatch end range",
				input: "C",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "neg charclass mismatch non-ASCII rune",
				input: "☺",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "neg charclass mismatch begin non-ASCII range",
				input: "α",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "neg charclass mismatch mid non-ASCII range",
				input: "β",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
			{
				name:  "neg charclass mismatch end non-ASCII range",
				input: "ξ",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abcA-C☹☺α-ξ]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- B\nB <- 'abc'",
		cases: []genTestCase{
			{
				name:  "ident match",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
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
				name:  "ident mismatch",
				input: "abz",
				pos:   0,
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- 'abc'*",
		cases: []genTestCase{
			{
				name:  "star match 0",
				input: "xyz",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "star match 0 EOF",
				input: "xyz",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "star match 1",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
			{
				name:  "star match >1",
				input: "abcabcabcxyz",
				pos:   len("abcabcabc"),
				node: &peg.Node{
					Name: "A",
					Text: "abcabcabc",
					Kids: []*peg.Node{
						{Text: "abc"},
						{Text: "abc"},
						{Text: "abc"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- .*",
		cases: []genTestCase{
			{
				name:  "star any",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
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
		grammar: "A <- [abc]*",
		cases: []genTestCase{
			{
				name:  "star charclass",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
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
		grammar: "A <- [^abc]*",
		cases: []genTestCase{
			{
				name:  "star neg charclass",
				input: "XYZ",
				pos:   len("XYZ"),
				node: &peg.Node{
					Name: "A",
					Text: "XYZ",
					Kids: []*peg.Node{
						{Text: "X"},
						{Text: "Y"},
						{Text: "Z"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- B*\nB <- 'abc'",
		cases: []genTestCase{
			{
				name:  "star ident",
				input: "abcabc",
				pos:   len("abcabc"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: "A <- ('a' 'b' 'c')*",
		cases: []genTestCase{
			{
				name:  "star subexpr",
				input: "abcabc",
				pos:   len("abcabc"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: "A <- 'abc'+",
		cases: []genTestCase{
			{
				name:  "plus match 1",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
			{
				name:  "plus match >1",
				input: "abcabcabcxyz",
				pos:   len("abcabcabc"),
				node: &peg.Node{
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
				name:  "plus mismatch",
				input: "xyz",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `"abc"`},
					},
				},
			},
			{
				name:  "plus mismatch EOF",
				input: "",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `"abc"`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- .+",
		cases: []genTestCase{
			{
				name:  "plus any match",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
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
				name:  "plus any mismatch",
				input: "",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `.`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- [abc]+",
		cases: []genTestCase{
			{
				name:  "plus charclass match",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
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
				name:  "plus charclass mismatch",
				input: "XYZ",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[abc]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- [^abc]+",
		cases: []genTestCase{
			{
				name:  "plus neg charclass match",
				input: "XYZ",
				pos:   len("XYZ"),
				node: &peg.Node{
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
				name:  "plus neg charclass mismatch",
				input: "abc",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `[^abc]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- B+\nB <- 'abc'",
		cases: []genTestCase{
			{
				name:  "plus ident match",
				input: "abcabc",
				pos:   len("abcabc"),
				node: &peg.Node{
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
				name:  "plus ident mismatch",
				input: "xyz",
				pos:   0,
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- ('a' 'b' 'c')+",
		cases: []genTestCase{
			{
				name:  "plus subexpr match",
				input: "abcabc",
				pos:   len("abcabc"),
				node: &peg.Node{
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
				name:  "plus subexpr mismatch",
				input: "xyz",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `"a"`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- 'abc'?",
		cases: []genTestCase{
			{
				name:  "question match 0",
				input: "xyz",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "question match 0 EOF",
				input: "xyz",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "question match 1",
				input: "abcabc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
		},
	},
	{
		grammar: "A <- .?",
		cases: []genTestCase{
			{
				name:  "question any",
				input: "a",
				pos:   len("a"),
				node: &peg.Node{
					Name: "A",
					Text: "a",
					Kids: []*peg.Node{{Text: "a"}},
				},
			},
		},
	},
	{
		grammar: "A <- [abc]?",
		cases: []genTestCase{
			{
				name:  "question charclass",
				input: "a",
				pos:   len("a"),
				node: &peg.Node{
					Name: "A",
					Text: "a",
					Kids: []*peg.Node{{Text: "a"}},
				},
			},
		},
	},
	{
		grammar: "A <- [^abc]?",
		cases: []genTestCase{
			{
				name:  "question neg charclass",
				input: "X",
				pos:   len("X"),
				node: &peg.Node{
					Name: "A",
					Text: "X",
					Kids: []*peg.Node{{Text: "X"}},
				},
			},
		},
	},
	{
		grammar: "A <- B?\nB <- 'abc'",
		cases: []genTestCase{
			{
				name:  "question ident",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: "A <- ('a' 'b' 'c')?",
		cases: []genTestCase{
			{
				name:  "question match subexpr",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
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
		},
	},
	{
		grammar: "A <- &'abc'",
		cases: []genTestCase{
			{
				name:  "pos pred match",
				input: "abc",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "pos pred mismatch",
				input: "xyz",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `&"abc"`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- &.",
		cases: []genTestCase{
			{
				name:  "pos pred any match",
				input: "a",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "pos pred any mismatch",
				input: "",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `&.`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- &[abc]",
		cases: []genTestCase{
			{
				name:  "pos pred charclass match",
				input: "a",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "pos pred charclass mismatch",
				input: "X",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `&[abc]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- &[^abc]",
		cases: []genTestCase{
			{
				name:  "pos pred neg charclass match",
				input: "X",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "pos pred neg charclass mismatch",
				input: "a",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `&[^abc]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- !'abc'",
		cases: []genTestCase{
			{
				name:  "neg pred match",
				input: "xyz",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "neg pred mismatch",
				input: "abc",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `!"abc"`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- !.",
		cases: []genTestCase{
			{
				name:  "neg pred any match",
				input: "",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "neg pred any mismatch",
				input: "a",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `!.`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- ![abc]",
		cases: []genTestCase{
			{
				name:  "neg pred charclass match",
				input: "x",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "neg pred charclass mismatch",
				input: "a",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `![abc]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- ![^abc]",
		cases: []genTestCase{
			{
				name:  "neg pred neg charclass match",
				input: "a",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "neg pred neg charclass mismatch",
				input: "x",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `![^abc]`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- !B\nB <- 'abc'",
		cases: []genTestCase{
			{
				name:  "neg pred literal match",
				input: "xyz",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "neg pred neg charclass mismatch",
				input: "abc",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: "!B"},
					},
				},
			},
		},
	},
	{
		grammar: "A <- 'abc' 'def' 'ghi'",
		cases: []genTestCase{
			{
				name:  "sequence match",
				input: "abcdefghi",
				pos:   len("abcdefghi"),
				node: &peg.Node{
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
				name:  "sequence mismatch first",
				input: "XYZdefghi",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `"abc"`},
					},
				},
			},
			{
				name:  "sequence mismatch mid",
				input: "abcXYZghi",
				pos:   len("abc"), // error after abc
				fail: &peg.Fail{
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
				name:  "sequence mismatch last",
				input: "abcdefXYZ",
				pos:   len("abcdef"), // error after abcdef
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{
							Pos:  len("abcdef"),
							Want: `"ghi"`,
						},
					},
				},
			},
		},
	},
	{
		grammar: "A <- 'abc' / 'def' / 'ghi'",
		cases: []genTestCase{
			{
				name:  "choice match first",
				input: "abc",
				pos:   len("abc"),
				node: &peg.Node{
					Name: "A",
					Text: "abc",
					Kids: []*peg.Node{{Text: "abc"}},
				},
			},
			{
				name:  "choice match mid",
				input: "def",
				pos:   len("def"),
				node: &peg.Node{
					Name: "A",
					Text: "def",
					Kids: []*peg.Node{{Text: "def"}},
				},
			},
			{
				name:  "choice match last",
				input: "ghi",
				pos:   len("ghi"),
				node: &peg.Node{
					Name: "A",
					Text: "ghi",
					Kids: []*peg.Node{{Text: "ghi"}},
				},
			},
			{
				name:  "choice mismatch",
				input: "XYZ",
				pos:   0,
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `"abc"`},
						{Want: `"def"`},
						{Want: `"ghi"`},
					},
				},
			},
		},
	},
	{
		grammar: "A <- 'abc' / 'def'?",
		cases: []genTestCase{
			{
				name:  "choice that can't fail empty",
				input: "XYZ",
				pos:   0,
				node: &peg.Node{
					Name: "A",
				},
			},
			{
				name:  "choice that can't fail non-empty",
				input: "def",
				pos:   len("def"),
				node: &peg.Node{
					Name: "A",
					Text: "def",
					Kids: []*peg.Node{{Text: "def"}},
				},
			},
		},
	},
	{
		grammar: "A <- '123' ('abc'/ 'αβξ')",
		cases: []genTestCase{
			{
				name:  "choice after sequence match first",
				input: "123abc",
				pos:   len("123abc"),
				node: &peg.Node{
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
				name:  "choice after sequence match second",
				input: "123αβξ",
				pos:   len("123αβξ"),
				node: &peg.Node{
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
				name:  "choice after sequence mismatch",
				input: "123XYZ",
				pos:   len("123"), // error after 123
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- 'a' B 'c' / 'a' B 'd'\nB <- 'B'",
		cases: []genTestCase{
			{
				name:  "rule memo success",
				input: "aBd",
				pos:   len("aBd"),
				node: &peg.Node{
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
				name:  "rule memo failure",
				input: "aAd",
				pos:   len("a"), // error after a
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- B 'x'\nB <- 'abc' 'def' / .",
		cases: []genTestCase{
			{
				name:  "latest error",
				input: "abcxyz",
				pos:   len("abc"), // latest error is after abc
				fail: &peg.Fail{
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
		},
	},
	{
		// Don't report the location of fails in silent exprs, & and !.
		grammar: "A <- !B 'xyz'\nB <- 'abc' 'def'",
		cases: []genTestCase{
			{
				name:  "ignore silent fails",
				input: "abc",
				pos:   0, // latest error is just before 'xyz', not after 'abc'
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{Want: `"xyz"`},
					},
				},
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
		grammar: "A <- &B 'f' / B\nB <- 'a' 'b' 'c' 'd' 'e'",
		cases: []genTestCase{
			{
				name:  "no cache silent fails",
				input: "abce",
				// The error is the missing 'd' between 'abc' and 'e'.
				// Some other PEG parsers would report the error at 0,
				// because the first time 'd' fails, it's silent, that's cached
				// and the subsequent fail uses the cached,
				// un-reported error.
				pos: len("abc"),
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- 'abc' B 'def'\nB 'name' <- C\nC <- D\nD <- '123'",
		cases: []genTestCase{
			{
				name:  "named rule fail",
				input: "abc124",
				pos:   len("abc"),
				fail: &peg.Fail{
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
		},
	},
	{
		grammar: "A <- 'abc' B 'def'\nB 'name' <- '1' '2' '3' / .",
		cases: []genTestCase{
			{
				name: "ignore errors under successful named rules",
				// B fails after 12, backtracks and succeeds after the 1.
				// We should not report the error after abc12, but after abc1.
				input: "abc12x",
				pos:   len("abc1"),
				fail: &peg.Fail{
					Name: "A",
					Kids: []*peg.Fail{
						{
							Pos:  len("abc1"),
							Want: `"def"`,
						},
					},
				},
			},
		},
	},
	{
		grammar: `
			A <- List<B> List<C>
			B <- "b"
			C <- "c"
			List<x> <- x*`,
		cases: []genTestCase{
			{
				name:  "unary template",
				input: "bbbccc",
				pos:   len("bbbccc"),
				node: &peg.Node{
					Name: "A",
					Text: "bbbccc",
					Kids: []*peg.Node{
						{
							Name: "List<B>",
							Text: "bbb",
							Kids: []*peg.Node{
								{
									Name: "B",
									Text: "b",
									Kids: []*peg.Node{{Text: "b"}},
								},
								{
									Name: "B",
									Text: "b",
									Kids: []*peg.Node{{Text: "b"}},
								},
								{
									Name: "B",
									Text: "b",
									Kids: []*peg.Node{{Text: "b"}},
								},
							},
						},
						{
							Name: "List<C>",
							Text: "ccc",
							Kids: []*peg.Node{
								{
									Name: "C",
									Text: "c",
									Kids: []*peg.Node{{Text: "c"}},
								},
								{
									Name: "C",
									Text: "c",
									Kids: []*peg.Node{{Text: "c"}},
								},
								{
									Name: "C",
									Text: "c",
									Kids: []*peg.Node{{Text: "c"}},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		grammar: `
			A <- Three<X, Y, Z>
			X <- "x"
			Y <- "y"
			Z <- "z"
			Three<x, y, z> <- x y z`,
		cases: []genTestCase{
			{
				name:  "3-ary template",
				input: "xyz",
				pos:   len("xyz"),
				node: &peg.Node{
					Name: "A",
					Text: "xyz",
					Kids: []*peg.Node{
						{
							Name: "Three<X, Y, Z>",
							Text: "xyz",
							Kids: []*peg.Node{
								{
									Name: "X",
									Text: "x",
									Kids: []*peg.Node{{Text: "x"}},
								},
								{
									Name: "Y",
									Text: "y",
									Kids: []*peg.Node{{Text: "y"}},
								},
								{
									Name: "Z",
									Text: "z",
									Kids: []*peg.Node{{Text: "z"}},
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestGen(t *testing.T) {
	for _, test := range genTests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			source := generateTest(prelude, test.grammar)
			binary := build(source)
			defer rm(binary)
			go rm(source)

			for _, c := range test.cases {
				test, c := test, c
				t.Run(c.name, func(t *testing.T) {
					t.Logf("%q\n", test.grammar)
					var result struct {
						Pos  int
						Perr int
						Node *peg.Node
						Fail *peg.Fail
					}
					parseGob(binary, c.input, &result)
					pos := result.Pos
					if result.Fail != nil {
						pos = result.Perr
					}
					t.Logf("result: %+v\n", result)
					if pos != c.pos {
						t.Errorf("parse(%q)=%d, want %d", c.input, pos, c.pos)
					}
					var got interface{}
					if result.Node != nil {
						got = result.Node
					} else {
						got = result.Fail
					}
					var want interface{}
					if c.node != nil {
						want = c.node
					} else {
						want = c.fail
					}
					if !reflect.DeepEqual(want, got) {
						t.Errorf("parse(%q)=\n%s\nwant\n%s",
							c.input, pretty.String(got), pretty.String(want))
					}
				})
			}
		})
	}
}

// generateTest generates Go source code for a Peggy
func generateTest(prelude string, input string) string {
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

func rm(file string) {
	if err := os.Remove(file); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove %s: %s", file, err)
	}
}

// parseGob parses an input using the given binary
// and returns the position of either the parse or error
// along with whether the parse succeeded.
func parseGob(binary, input string, result interface{}) {
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
