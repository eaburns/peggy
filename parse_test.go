// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/eaburns/pretty"
)

// A ParserTest is a Peggy input-file parser test
// with a given input and expected string formats.
type ParserTest struct {
	Name  string
	Input string
	// FullString is the expected fully parenthesized string.
	FullString string
	// String is the expected regular String string.
	// This is the same as Input, but without
	// comments and unnecessary whitespace,
	// except for a single space, " ",
	// separating sub-exprsessions of a sequence,
	// and on either side of <-.
	String string
	// Prelude is the expected file prelude text.
	Prelude string
	// Error is a regexp string that matches an expected parse error.
	Error string
}

// ParseTests is a set of tests matching
// FullString and String outputs with expected outputs for successful parses,
// and expected parse errors for failed parses.
// If Input contains a ☹ rune, the io.RuneScanner returns an error on that rune.
var ParseTests = []ParserTest{
	{
		Name:       "empty",
		Input:      "",
		FullString: "",
		String:     "",
	},
	{
		Name:       "only whitespace",
		Input:      "  \n\n\t    ",
		FullString: "",
		String:     "",
	},
	{
		Name:       "simple rule",
		Input:      "A <- B",
		FullString: "A <- (B)",
		String:     "A <- B",
	},
	{
		Name:       "named rule",
		Input:      `A "name" <- B`,
		FullString: `A "name" <- (B)`,
		String:     `A "name" <- B`,
	},
	{
		Name:       "named rule, single quotes",
		Input:      `A 'name' <- B`,
		FullString: `A "name" <- (B)`,
		String:     `A "name" <- B`,
	},
	{
		Name:       "named rule, empty name",
		Input:      `A "" <- B`,
		FullString: `A "" <- (B)`,
		String:     `A "" <- B`,
	},
	{
		Name:       "named rule, escapes",
		Input:      `A "\t\nabc" <- B`,
		FullString: `A "\t\nabc" <- (B)`,
		String:     `A "\t\nabc" <- B`,
	},
	{
		Name: "prelude and simple rule",
		Input: `{
package main

import "fmt"

func main() { fmt.Println("Hello, World") }
}
A <- B`,
		FullString: "A <- (B)",
		String:     "A <- B",
		Prelude: `
package main

import "fmt"

func main() { fmt.Println("Hello, World") }
`,
	},
	{
		Name:       "multiple simple rules",
		Input:      "A <- B\nC <- D",
		FullString: "A <- (B)\nC <- (D)",
		String:     "A <- B\nC <- D",
	},
	{
		Name:       "multiple simple rules",
		Input:      "A <- B\nC <- D",
		FullString: "A <- (B)\nC <- (D)",
		String:     "A <- B\nC <- D",
	},
	{
		Name:       "whitespace",
		Input:      "\tA <- B\n   \n\n    C <- D\t  ",
		FullString: "A <- (B)\nC <- (D)",
		String:     "A <- B\nC <- D",
	},
	{
		Name:       "comments",
		Input:      "# comment\nA <- B # comment\n# comment",
		FullString: "A <- (B)",
		String:     "A <- B",
	},

	// Operands.
	{
		Name:       "& pred code",
		Input:      "A <- &{pred}",
		FullString: "A <- (&{pred})",
		String:     "A <- &{pred}",
	},
	{
		Name:       "! pred code",
		Input:      "A <- !{pred}",
		FullString: "A <- (!{pred})",
		String:     "A <- !{pred}",
	},
	{
		Name:       "any",
		Input:      "A <- .",
		FullString: "A <- (.)",
		String:     "A <- .",
	},
	{
		Name:       "identifier",
		Input:      "A <- BCD",
		FullString: "A <- (BCD)",
		String:     "A <- BCD",
	},
	{
		Name:       "non-ASCII identifier",
		Input:      "Â <- _αβξ",
		FullString: "Â <- (_αβξ)",
		String:     "Â <- _αβξ",
	},
	{
		Name:       "double-quote string",
		Input:      `A <- "BCD☺"`,
		FullString: `A <- ("BCD☺")`,
		String:     `A <- "BCD☺"`,
	},
	{
		Name:       "single-quote string",
		Input:      `A <- 'BCD☺'`,
		FullString: `A <- ("BCD☺")`,
		String:     `A <- "BCD☺"`,
	},
	{
		Name:       "character class",
		Input:      `A <- [abc\nxyzαβξ1-9A-Z\-]`,
		FullString: `A <- ([abc\nxyzαβξ1-9A-Z\-])`,
		String:     `A <- [abc\nxyzαβξ1-9A-Z\-]`,
	},
	{
		Name:       "^ character class",
		Input:      `A <- [^^abc\nxyzαβξ]`,
		FullString: `A <- ([^\^abc\nxyzαβξ])`,
		String:     `A <- [^\^abc\nxyzαβξ]`,
	},
	{
		Name:       "character class, delimiters",
		Input:      `A <- [[\]]`,
		FullString: `A <- ([[\]])`,
		String:     `A <- [[\]]`,
	},
	{
		// ^ should only negate the class if it's at the beginning
		Name:       "character class, non-first^",
		Input:      `A <- [abc^]`,
		FullString: `A <- ([abc\^])`,
		String:     `A <- [abc\^]`,
	},
	{
		Name:       "character class, escaping",
		Input:      `A <- [\a] [\b] [\f] [\n] [\r] [\t] [\v] [\\] [\-] [\]] [\101] [\x41] [\u0041] [\U00000041] [\aa\b] [a\ab] [\^]`,
		FullString: `A <- ((((((((((((((((([\a]) ([\b])) ([\f])) ([\n])) ([\r])) ([\t])) ([\v])) ([\\])) ([\-])) ([\]])) ([A])) ([A])) ([A])) ([A])) ([\aa\b])) ([a\ab])) ([\^]))`,
		String:     `A <- [\a] [\b] [\f] [\n] [\r] [\t] [\v] [\\] [\-] [\]] [A] [A] [A] [A] [\aa\b] [a\ab] [\^]`,
	},

	// Associativity.
	{
		Name:       "choice associativity",
		Input:      "A <- B/C/D",
		FullString: "A <- (((B)/(C))/(D))",
		String:     "A <- B/C/D",
	},
	{
		Name:       "sequence associativity",
		Input:      "A <- B C D",
		FullString: "A <- (((B) (C)) (D))",
		String:     "A <- B C D",
	},

	// Precedence.
	{
		Name:       "various precedences",
		Input:      "A <- x:B*+ C?/(!D y:&E)* T:{c}/F !{p}",
		FullString: "A <- ((((x:(((B)*)+)) ((C)?))/((((!(D)) (y:(&(E))))*) T:{c}))/((F) (!{p})))",
		String:     "A <- x:B*+ C?/(!D y:&E)* T:{c}/F !{p}",
	},
	{
		Name:       "action < choice",
		Input:      "A <- B T:{act}/C T:{act}",
		FullString: "A <- (((B) T:{act})/((C) T:{act}))",
		String:     "A <- B T:{act}/C T:{act}",
	},
	{
		Name:       "sequence < action",
		Input:      "A <- B C T:{act}",
		FullString: "A <- (((B) (C)) T:{act})",
		String:     "A <- B C T:{act}",
	},
	{
		Name:       "label < sequence",
		Input:      "A <- s:A t:B",
		FullString: "A <- ((s:(A)) (t:(B)))",
		String:     "A <- s:A t:B",
	},
	{
		Name:       "pred < label",
		Input:      "A <- s:!A t:&B",
		FullString: "A <- ((s:(!(A))) (t:(&(B))))",
		String:     "A <- s:!A t:&B",
	},
	{
		Name:       "rep < pred",
		Input:      "A <- !A* &B+ !C?",
		FullString: "A <- (((!((A)*)) (&((B)+))) (!((C)?)))",
		String:     "A <- !A* &B+ !C?",
	},
	{
		Name: "operand < rep",
		Input: `A <- (a/b c)*
B <- &{pred}*
C <- !{pred}*
D <- .*
E <- Z*
F <- "cde"*
G <- [fgh]*`,
		FullString: `A <- (((a)/((b) (c)))*)
B <- ((&{pred})*)
C <- ((!{pred})*)
D <- ((.)*)
E <- ((Z)*)
F <- (("cde")*)
G <- (([fgh])*)`,
		String: `A <- (a/b c)*
B <- &{pred}*
C <- !{pred}*
D <- .*
E <- Z*
F <- "cde"*
G <- [fgh]*`,
	},

	// Actions
	{
		Name:       "action with ident type",
		Input:      `A <- "abc" T:{act}`,
		FullString: `A <- (("abc") T:{act})`,
		String:     `A <- "abc" T:{act}`,
	},
	{
		Name:       "action with string type",
		Input:      `A <- "abc" "interface{}":{act}`,
		FullString: `A <- (("abc") "interface{}":{act})`,
		String:     `A <- "abc" "interface{}":{act}`,
	},
	{
		Name:       "action strip unnecessary quotes",
		Input:      `A <- "abc" "XYZ":{act}`,
		FullString: `A <- (("abc") XYZ:{act})`,
		String:     `A <- "abc" XYZ:{act}`,
	},

	// Rune escaping
	{
		Name:       `escape \a`,
		Input:      `A <- "\a"`,
		FullString: `A <- ("\a")`,
		String:     `A <- "\a"`,
	},
	{
		Name:       `escape \b`,
		Input:      `A <- "\b"`,
		FullString: `A <- ("\b")`,
		String:     `A <- "\b"`,
	},
	{
		Name:       `escape \f`,
		Input:      `A <- "\f"`,
		FullString: `A <- ("\f")`,
		String:     `A <- "\f"`,
	},
	{
		Name:       `escape \n`,
		Input:      `A <- "\n"`,
		FullString: `A <- ("\n")`,
		String:     `A <- "\n"`,
	},
	{
		Name:       `escape \r`,
		Input:      `A <- "\r"`,
		FullString: `A <- ("\r")`,
		String:     `A <- "\r"`,
	},
	{
		Name:       `escape \t`,
		Input:      `A <- "\t"`,
		FullString: `A <- ("\t")`,
		String:     `A <- "\t"`,
	},
	{
		Name:       `escape \v`,
		Input:      `A <- "\v"`,
		FullString: `A <- ("\v")`,
		String:     `A <- "\v"`,
	},
	{
		Name:       `escape \\`,
		Input:      `A <- "\\"`,
		FullString: `A <- ("\\")`,
		String:     `A <- "\\"`,
	},
	{
		Name:       `escape \"`,
		Input:      `A <- "\""`,
		FullString: `A <- ("\"")`,
		String:     `A <- "\""`,
	},
	{
		Name:       `escape \'`,
		Input:      `A <- '\''`,
		FullString: `A <- ("'")`,
		String:     `A <- "'"`,
	},
	{
		Name:       `escape \000`,
		Input:      `A <- "\000"`,
		FullString: `A <- ("\x00")`,
		String:     `A <- "\x00"`,
	},
	{
		Name:       `escape \101 (A)`,
		Input:      `A <- "\101"`,
		FullString: `A <- ("A")`,
		String:     `A <- "A"`,
	},
	{
		Name:       `escape \101BCD`,
		Input:      `A <- "\101BCD"`,
		FullString: `A <- ("ABCD")`,
		String:     `A <- "ABCD"`,
	},
	{
		Name:       `escape \377 (255)`,
		Input:      `A <- "\377"`,
		FullString: `A <- ("ÿ")`, // \xFF
		String:     `A <- "ÿ"`,
	},
	{
		Name:  `escape \400 (256)`,
		Input: `A <- "\400"`,
		Error: "^test.file:1.6,1.11:.*>255",
	},
	{
		Name:  `escape \400 (256)`,
		Input: `A <- "xyz\400"`,
		// TODO: report the correct error location.
		Error: "^test.file:1.6,1.14:.*>255",
	},
	{
		Name:  `escape \4`,
		Input: `A <- "\4"`,
		Error: "^test.file:1.6,1.10: unknown escape sequence",
	},
	{
		Name:  `escape \40`,
		Input: `A <- "\40"`,
		Error: "^test.file:1.6,1.11: unknown escape sequence",
	},
	{
		Name:       `escape \x00`,
		Input:      `A <- "\x00"`,
		FullString: `A <- ("\x00")`,
		String:     `A <- "\x00"`,
	},
	{
		Name:       `escape \x41 (A)`,
		Input:      `A <- "\x41"`,
		FullString: `A <- ("A")`,
		String:     `A <- "A"`,
	},
	{
		Name:       `escape \x41BCD`,
		Input:      `A <- "\x41BCD"`,
		FullString: `A <- ("ABCD")`,
		String:     `A <- "ABCD"`,
	},
	{
		Name:       `escape \xFF`,
		Input:      `A <- "\xFF"`,
		FullString: `A <- ("ÿ")`, // \xFF
		String:     `A <- "ÿ"`,
	},
	{
		Name:  `escape \xF`,
		Input: `A <- "\xF"`,
		Error: "^test.file:1.6,1.11: unknown escape sequence",
	},
	{
		Name:       `escape \u0000`,
		Input:      `A <- "\u0000"`,
		FullString: `A <- ("\x00")`,
		String:     `A <- "\x00"`,
	},
	{
		Name:       `escape \u0041 (A)`,
		Input:      `A <- "\u0041"`,
		FullString: `A <- ("A")`,
		String:     `A <- "A"`,
	},
	{
		Name:       `escape \u0041BCD`,
		Input:      `A <- "\u0041BCD"`,
		FullString: `A <- ("ABCD")`,
		String:     `A <- "ABCD"`,
	},
	{
		Name:       `escape \u263A (☺)`,
		Input:      `A <- "\u263A"`,
		FullString: `A <- ("☺")`,
		String:     `A <- "☺"`,
	},
	{
		Name:       `escape \u263a (☺)`,
		Input:      `A <- "\u263a"`,
		FullString: `A <- ("☺")`,
		String:     `A <- "☺"`,
	},
	{
		Name:  `escape \uF`,
		Input: `A <- "\xF"`,
		Error: "^test.file:1.6,1.11: unknown escape sequence",
	},
	{
		Name:  `escape \uFF`,
		Input: `A <- "\uFF"`,
		Error: "^test.file:1.6,1.12: unknown escape sequence",
	},
	{
		Name:  `escape \uFFF`,
		Input: `A <- "\uFFF"`,
		Error: "^test.file:1.6,1.13: unknown escape sequence",
	},
	{
		Name:       `escape \U00000000`,
		Input:      `A <- "\U00000000"`,
		FullString: `A <- ("\x00")`,
		String:     `A <- "\x00"`,
	},
	{
		Name:       `escape \U00000041 (A)`,
		Input:      `A <- "\U00000041"`,
		FullString: `A <- ("A")`,
		String:     `A <- "A"`,
	},
	{
		Name:       `escape \U00000041BCD`,
		Input:      `A <- "\U00000041BCD"`,
		FullString: `A <- ("ABCD")`,
		String:     `A <- "ABCD"`,
	},
	{
		Name:       `escape \U0000263A (☺)`,
		Input:      `A <- "\U0000263A"`,
		FullString: `A <- ("☺")`,
		String:     `A <- "☺"`,
	},
	{
		Name:       `escape \U0000263a (☺)`,
		Input:      `A <- "\U0000263a"`,
		FullString: `A <- ("☺")`,
		String:     `A <- "☺"`,
	},
	{
		Name:       `escape \U0010FFFF`,
		Input:      `A <- "\U0010FFFF"`,
		FullString: `A <- ("\U0010ffff")`,
		String:     `A <- "\U0010ffff"`,
	},
	{
		Name:  `escape \U00110000`,
		Input: `A <- "\U00110000"`,
		Error: "^test.file:1.6,1.17:.*>0x10FFFF",
	},
	{
		Name:  `escape \UF`,
		Input: `A <- "\UF"`,
		Error: "^test.file:1.6,1.11: unknown escape sequence",
	},
	{
		Name:  `escape \UFF`,
		Input: `A <- "\UFF"`,
		Error: "^test.file:1.6,1.12: unknown escape sequence",
	},
	{
		Name:  `escape \UFFF`,
		Input: `A <- "\UFFF"`,
		Error: "^test.file:1.6,1.13: unknown escape sequence",
	},
	{
		Name:  `escape \UFFFF`,
		Input: `A <- "\UFFFF"`,
		Error: "^test.file:1.6,1.14: unknown escape sequence",
	},
	{
		Name:  `escape \UFFFFF`,
		Input: `A <- "\UFFFFF"`,
		Error: "^test.file:1.6,1.15: unknown escape sequence",
	},
	{
		Name:  `escape \UFFFFFF`,
		Input: `A <- "\UFFFFFF"`,
		Error: "^test.file:1.6,1.16: unknown escape sequence",
	},
	{
		Name:  `escape \UFFFFFFF`,
		Input: `A <- "\UFFFFFFF"`,
		Error: "^test.file:1.6,1.17: unknown escape sequence",
	},
	{
		Name:       `string with multiple escapes`,
		Input:      `A <- "x\a\b\f\n\r\t\v\\\"\000\x00\u0000\U00000000☺"`,
		FullString: `A <- ("x\a\b\f\n\r\t\v\\\"\x00\x00\x00\x00☺")`,
		String:     `A <- "x\a\b\f\n\r\t\v\\\"\x00\x00\x00\x00☺"`,
	},
	{
		Name:  `unknown escape`,
		Input: `A <- "\z"`,
		Error: "^test.file:1.6,1.9: unknown escape sequence",
	},
	{
		Name:  `escape eof`,
		Input: `A <- "\`,
		Error: `^test.file:1.6,1.8: unclosed "`,
	},

	// Whitespace.
	// BUG: The current YACC grammar
	// doesn't allow whitespace between all tokens,
	// but only particular tokens.
	// Specifically whitespace can only appear after
	// delimiters after which a new rule cannot begin.
	// This is because, in order to remain LALR(1),
	// a newline terminates a sequence expression,
	// denoting that the next identifier is a rule name.
	{
		Name: `after <-`,
		Input: `A <-
		"a"

		B <- #comment
		"b"

		C "c" <-
		"c"

		D "d" <- #comment
		"d"`,
		FullString: `A <- ("a")
B <- ("b")
C "c" <- ("c")
D "d" <- ("d")`,
		String: `A <- "a"
B <- "b"
C "c" <- "c"
D "d" <- "d"`,
	},
	{
		Name: `after /`,
		Input: `A <- B /
		C / # comment
		D`,
		FullString: `A <- (((B)/(C))/(D))`,
		String:     `A <- B/C/D`,
	},
	{
		Name: `after : label`,
		Input: `A <- l:
		B m: #comment
		C`,
		FullString: `A <- ((l:(B)) (m:(C)))`,
		String:     `A <- l:B m:C`,
	},
	{
		Name: `after & predicate`,
		Input: `A <- &
		B & #comment
		C`,
		FullString: `A <- ((&(B)) (&(C)))`,
		String:     `A <- &B &C`,
	},
	{
		Name: `after ! predicate`,
		Input: `A <- !
		B ! #comment
		C`,
		FullString: `A <- ((!(B)) (!(C)))`,
		String:     `A <- !B !C`,
	},
	{
		Name: `after (`,
		Input: `A <- (
		B ( #comment
		C))`,
		FullString: `A <- ((B) (C))`,
		String:     `A <- (B (C))`,
	},
	{
		Name: `before )`,
		Input: `A <- (B (C
		) #comment
		)`,
		FullString: `A <- ((B) (C))`,
		String:     `A <- (B (C))`,
	},
	{
		Name: `after & code`,
		Input: `A <- &
		{code} & #comment
		{CODE}`,
		FullString: `A <- ((&{code}) (&{CODE}))`,
		String:     `A <- &{code} &{CODE}`,
	},
	{
		Name: `after ! code`,
		Input: `A <- !
		{code} ! #comment
		{CODE}`,
		FullString: `A <- ((!{code}) (!{CODE}))`,
		String:     `A <- !{code} !{CODE}`,
	},
	{
		Name: `after : type`,
		Input: `A <- A "t":
		{code} / B T: #comment
		{CODE}`,
		FullString: `A <- (((A) t:{code})/((B) T:{CODE}))`,
		String:     `A <- A t:{code}/B T:{CODE}`,
	},

	// Systax errors.
	{
		Name:  "bad rule name",
		Input: "\n\t\t&",
		Error: "^test.file:2.3,2.4:",
	},
	{
		Name:  "missing <-",
		Input: "\nA B",
		Error: "^test.file:2.3,2.4:",
	},
	{
		Name:  "bad <-",
		Input: "\nA <~ C",
		Error: "^test.file:2.4,2.5:",
	},
	{
		Name:  "missing expr",
		Input: "\nA <-",
		Error: "^test.file:2.5:",
	},
	{
		Name:  "unexpected rune",
		Input: "\nA <- C ☺",
		Error: "^test.file:2.8,2.9:",
	},
	{
		Name:  "unclosed (",
		Input: "\nA <- (B",
		Error: "^test.file:2.8:",
	},
	{
		Name:  "unclosed '",
		Input: "\nA <- 'B",
		Error: "^test.file:2.6,2.8: unclosed '",
	},
	{
		Name:  `unclosed "`,
		Input: "\nA <- \"B",
		Error: "^test.file:2.6,2.8: unclosed \"",
	},
	{
		Name:  `unclosed {`,
		Input: "\nA <- B { code",
		Error: "^test.file:2.8,2.14: unclosed {",
	},
	{
		Name:  `unclosed spans lines`,
		Input: "\nA <- \"B\n\nC",
		Error: "^test.file:2.6,4.2: unclosed \"",
	},
	{
		Name:  "unclosed [",
		Input: "\nA <- [B",
		Error: "^test.file:2.6,2.8: unclosed [[]",
	},
	{
		Name:  "character class empty",
		Input: "\nA <- []",
		Error: "^test.file:2.6,2.8: bad char class: empty",
	},
	{
		Name:  "character class starts with span",
		Input: "\nA <- [-9]",
		Error: "^test.file:2.7,2.9: bad span",
	},
	{
		Name:  "character class no span start",
		Input: "\nA <- [1-3-9]",
		Error: "^test.file:2.10,2.12: bad span",
	},
	{
		Name:  "character class ends with span",
		Input: "\nA <- [0-]",
		Error: "^test.file:2.7,2.9: bad span",
	},
	{
		Name:  "character class inverted span",
		Input: "\nA <- [9-0]",
		Error: "^test.file:2.7,2.10: bad span",
	},
	{
		Name:  "character class span after span",
		Input: "\nA <- [^0-9abcA-Zz-a]",
		Error: "^test.file:2.17,2.20: bad span",
	},
	{
		Name:  "character class bad span after rune",
		Input: "\nA <- [^0-9abcZ-A]",
		Error: "^test.file:2.14,2.17: bad span",
	},

	// Go syntax errors.
	{
		Name:  `bad prelude`,
		Input: "{ not package line }\nA <- B",
		Error: "^test.file:1.3",
	},
	{
		Name: `bad multi-line prelude`,
		Input: `{
package main

import "fmt"

// Missing open paren.
func main() { fmt.Println"Hello, World") }
}
A <- B`,
		Error: "^test.file:7.26",
	},
	{
		Name: `bad bool expression`,
		// = instead of ==.
		Input: "\nA <- &{ x = z}",
		Error: "^test.file:2.11",
	},
	{
		Name: `bad multi-line bool expression`,
		// Missing the closed paren on p(.
		Input: "\nA <- &{ x == \n p(y, z, h}",
		Error: "^test.file:3.11",
	},
	{
		Name: `bad action`,
		// = instead of ==.
		Input: "\nA <- B T:{ if ( }",
		Error: "^test.file:2.17",
	},
	{
		Name: `bad multi-line action`,
		// = instead of ==.
		Input: "\nA <- B T:{\n	if ( }",
		Error: "^test.file:3.7",
	},
	{
		Name: `bad action: invalid nested func def`,
		// = instead of ==.
		Input: "\nA <- B T:{ func f() int { return 1 } }",
		Error: "^test.file:2.17",
	},

	// I/O errors.
	{
		Name:  "only I/O error",
		Input: "☹",
		Error: testIOError,
	},
	{
		Name:  "comment I/O error",
		Input: "#☹",
		Error: testIOError,
	},
	{
		Name:  "ident I/O error",
		Input: "A☹",
		Error: testIOError,
	},
	{
		Name:  "arrow I/O error",
		Input: "A <☹",
		Error: testIOError,
	},
	{
		Name:  "code I/O error",
		Input: "A <- B { ☹",
		Error: testIOError,
	},
	{
		Name:  "char class I/O error",
		Input: "A <- [☹",
		Error: testIOError,
	},
	{
		Name:  "double-quoted string I/O error",
		Input: "A <- \"☹",
		Error: testIOError,
	},
	{
		Name:  "single-quoted string I/O error",
		Input: "A <- '☹",
		Error: testIOError,
	},
}

func TestParse(t *testing.T) {
	for _, test := range ParseTests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			in := testRuneScanner{strings.NewReader(test.Input)}
			g, err := Parse(in, "test.file")

			if test.Error != "" {
				if err == nil {
					t.Log(pretty.String(g.Rules))
					t.Errorf("Parse(%q) ok, but expected error matching %q",
						test.Input, test.Error)
					return
				}
				re := regexp.MustCompile(test.Error)
				if !re.MatchString(err.Error()) {
					t.Errorf("Parse(%q) err=%q, but expected to match %q",
						test.Input, err.Error(), test.Error)
					return
				}
				return
			}

			if err != nil {
				t.Errorf("Parse(%q) failed: %s", test.Input, err)
				return
			}
			var pre string
			if g.Prelude != nil {
				pre = g.Prelude.String()
			}
			if pre != test.Prelude {
				t.Errorf("Parse(%q).Prelude=\n%s\nwant:\n%s",
					test.Input, pre, test.Prelude)
				return
			}
			if s := FullString(g.Rules); s != test.FullString {
				t.Errorf("Parse(%q)\nfull string:\n%q\nwant:\n%q",
					test.Input, s, test.FullString)
				return
			}
			if s := String(g.Rules); s != test.String {
				t.Errorf("Parse(%q)\nstring:\n%q\nwant:\n%q",
					test.Input, s, test.String)
				return
			}
		})
	}
}

// testRuneScanner implements io.RuneScanner, wrapping another RuneScanner,
// however, whenever the original scanner would've returned a ☹ rune,
// testRuneScanner instead returns an error.
type testRuneScanner struct {
	io.RuneScanner
}

const testIOError = "test I/O error"

func (rs testRuneScanner) ReadRune() (rune, int, error) {
	r, n, err := rs.RuneScanner.ReadRune()
	if r == '☹' {
		return 0, 0, errors.New(testIOError)
	}
	return r, n, err
}
