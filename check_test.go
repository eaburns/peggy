// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name string
		in   string
		err  string
	}{
		{
			name: "empty OK",
			in:   "",
			err:  "",
		},
		{
			name: "various OK",
			in: `A <- (G/B C)*
B <- &{pred}*
C <- !{pred}* string:{ act }
D <- .* !B
E <- C*
F <- "cde"*
G <- [fgh]*`,
			err: "",
		},
		{
			name: "redefined rule",
			in:   "A <- [x]\nA <- [y]",
			err:  "^test.file:2.1,2.9: rule A redefined",
		},
		{
			name: "undefined rule",
			in:   "A <- B",
			err:  "^test.file:1.6,1.7: rule B undefined",
		},
		{
			name: "redefined label",
			in:   "A <- a:[a] a:[a]",
			err:  "^test.file:1.12,1.13: label a redefined",
		},
		{
			name: "choice first error",
			in:   "A <- Undefined / A",
			err:  ".+",
		},
		{
			name: "choice second error",
			in:   "A <- B / Undefined\nB <- [x]",
			err:  ".+",
		},
		{
			name: "seq first error",
			in:   "A <- Undefined A",
			err:  ".+",
		},
		{
			name: "sequence second error",
			in:   "A <- B Undefined\nB <- [x]",
			err:  ".+",
		},
		{
			name: "template parameter OK",
			in: `A<x> <- x
				B <- A<C>
				C <- "c"`,
			err: "",
		},
		{
			name: "template parameter redef",
			in: `A<x, x> <- x
				B <- A<C, C>
				C <- "c"`,
			err: "^test.file:1.6,1.7: parameter x redefined$",
		},
		{
			name: "template and non-template redef",
			in: `A<x> <- x
				B <- A<C>
				C <- "c"
				A <- "a"`,
			err: "^test.file:4.5,4.13: rule A redefined$",
		},
		{
			name: "template arg count mismatch",
			in: `A<x> <- x
				B <- A<C, C>
				C <- "c"`,
			err: "test.file:2.10,2.16: template A<x> argument count mismatch: got 2, expected 1",
		},
		{
			name: "multiple errors",
			in:   "A <- U1 U2\nA <- u:[x] u:[x]",
			err: "test.file:1.6,1.8: rule U1 undefined\n" +
				"test.file:1.9,1.11: rule U2 undefined\n" +
				"test.file:2.1,2.17: rule A redefined\n" +
				"test.file:2.12,2.13: label u redefined",
		},
		{
			name: "right recursion is OK",
			in: `A <- "b" B
				B <- A`,
		},
		{
			name: "direct left-recursion",
			in:   `A <- A`,
			err:  "^test.file:1.1,1.7: left-recursion: A, A$",
		},
		{
			name: "indirect left-recursion",
			in: `A <- C0
				C0 <- C1
				C1 <- C2
				C2 <- C0`,
			err: "^test.file:2.5,2.13: left-recursion: C0, C1, C2, C0$",
		},
		{
			name: "choice left-recursion",
			in: `A <- B / C / D
				B <- "b"
				C <- "c"
				D <- A`,
			err: "^test.file:1.1,1.15: left-recursion: A, D, A$",
		},
		{
			name: "sequence left-recursion",
			in: `A <- !B C D E
				B <- "b"
				C <- !"c"
				D <- C # non-consuming through C
				E <- A`,
			err: "^test.file:1.1,1.14: left-recursion: A, E, A$",
		},
		{
			name: "various expr left-recursion",
			in: `Choice <- "a" / Sequence
				Sequence <- SubExpr "b"
				SubExpr <- ( PredExpr )
				PredExpr <- &RepExpr
				RepExpr <- OptExpr+
				OptExpr <- Action?
				Action <- Choice string:{ return "" }`,
			err: "^test.file:1.1,1.25: left-recursion: Choice, Sequence, SubExpr, PredExpr, RepExpr, OptExpr, Action, Choice$",
		},
		{
			name: "template left-recursion",
			in: `A <- C0
				C0 <- C1
				C1 <- C2<C0>
				C2<X> <- X`,
			err: "^test.file:2.5,2.13: left-recursion: C0, C1, C2<C0>, C0$",
		},
		{
			name: "multiple left-recursion errors",
			in: `A <- A
				B <- C
				C <- B`,
			err: "^test.file:1.1,1.7: left-recursion: A, A\n" +
				"test.file:2.5,2.11: left-recursion: B, C, B$",
		},
		{
			name: "right-recursion is OK",
			in: `A <- B C A?
				B <- "b" B / C
				C <- "c"`,
			err: "",
		},

		{
			name: "choice type mismatch",
			in:   `A <- "a" / "b" int:{ return 5 }`,
			err:  "^test.file:1.12,1.32: type mismatch: got int, expected string",
		},
		{
			name: "sequence type mismatch",
			in:   `A <- "a" ( "b" int:{ return 5 } )`,
			err:  "^test.file:1.10,1.33: type mismatch: got int, expected string",
		},
		{
			name: "unused choice, no mismatch",
			in:   `A <- ( "a" / "b" int:{ return 5 } ) int:{ return 6 }`,
			err:  "",
		},
		{
			name: "unused sequence, no mismatch",
			in:   `A <- "a" ( "b" int:{ return 5 } ) int:{ return 6 }`,
			err:  "",
		},
		{
			name: "&-pred subexpression is unused",
			in:   `A <- "a" !( "b" int:{ return 5 } )`,
			err:  "",
		},
		{
			name: "!-pred subexpression is unused",
			in:   `A <- "a" !( "b" int:{ return 5 } )`,
			err:  "",
		},
		{
			name: "multiple type errors",
			in: `A <- B ( "c" int: { return 0 } )
				B <- "b" / ( "c" int: { return 0 } )`,
			err: "^test.file:1.8,1.32: type mismatch: got int, expected string\n" +
				"test.file:2.16,2.40: type mismatch: got int, expected string$",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			in := strings.NewReader(test.in)
			g, err := Parse(in, "test.file")
			if err != nil {
				t.Errorf("Parse(%q, _)=_, %v, want _,nil", test.in, err)
				return
			}
			err = Check(g)
			if test.err == "" {
				if err != nil {
					t.Errorf("Check(%q)=%v, want nil", test.in, err)
				}
				return
			}
			re := regexp.MustCompile(test.err)
			if err == nil || !re.MatchString(err.Error()) {
				var e string
				if err != nil {
					e = err.Error()
				}
				t.Errorf("Check(%q)=%v, but expected to match %q",
					test.in, e, test.err)
				return
			}
		})
	}
}
