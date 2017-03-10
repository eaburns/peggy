// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"errors"
	"fmt"
	"io"
	"unicode"
)

const eof = -1

type text struct {
	str        string
	begin, end Loc
}

func (t text) PrettyPrint() string {
	return fmt.Sprintf(`Text{%d:%d-%d:%d: "%s"}`,
		t.begin.Line, t.begin.Col,
		t.end.Line, t.end.Col,
		t.str)
}

func (t text) String() string { return t.str }
func (t text) Begin() Loc     { return t.begin }
func (t text) End() Loc       { return t.end }

type lexer struct {
	in                                io.RuneScanner
	file                              string
	n, line, lineStart, prevLineStart int
	eof                               bool

	// prevBegin is the beginning of the most-recently scanned token.
	// prevEnd is the end of the most-recently scanned token.
	// These are used for error reporting.
	prevBegin, prevEnd Loc

	// err is non-nil if there was an error during parsing.
	err error
	// result contains the Grammar resulting from a successful parse.
	result Grammar
}

// Begin returns the begin location of the last returned token.
func (x *lexer) Begin() Loc { return x.prevBegin }

// End returns the end location of the last returned token.
func (x *lexer) End() Loc { return x.prevEnd }

func (x *lexer) loc() Loc {
	return Loc{
		File: x.file,
		Line: x.line,
		Col:  x.n - x.lineStart + 1,
	}
}

func (x *lexer) next() (rune, error) {
	if x.eof {
		return eof, nil
	}
	r, _, err := x.in.ReadRune()
	if err == io.EOF {
		x.eof = true
		return eof, nil
	}
	x.n++
	if r == '\n' {
		x.prevLineStart = x.lineStart
		x.lineStart = x.n
		x.line++
	}
	return r, err
}

func (x *lexer) back() error {
	if x.eof {
		return nil
	}
	if x.lineStart == x.n {
		x.lineStart = x.prevLineStart
		x.line--
	}
	x.n--
	return x.in.UnreadRune()
}

func (x *lexer) Error(s string) {
	if x.err != nil {
		return
	}
	x.err = Err(x, s)
}

func (x *lexer) Lex(lval *peggySymType) (v int) {
	defer func() { x.prevEnd = x.loc() }()
	for {
		x.prevBegin = x.loc()
		lval.text.begin = x.loc()
		lval.loc = x.loc()
		r, err := x.next()

		switch {
		case err != nil:
			break

		case r == '#':
			if err = comment(x); err != nil {
				break
			}
			return '\n'

		case unicode.IsLetter(r) || r == '_':
			if lval.text.str, err = ident(x); err != nil {
				break
			}
			lval.text.str = string([]rune{r}) + lval.text.str
			lval.text.end = x.loc()
			return _IDENT

		case r == '<':
			b := x.loc()
			if r, err = x.next(); err != nil {
				break
			}
			lval.text.str = string([]rune{'<', r})
			lval.text.end = x.loc()
			if r != '-' {
				x.prevBegin = b
				return int(r)
			}
			return _ARROW

		case r == '{':
			if lval.text.str, err = code(x); err != nil {
				break
			}
			lval.text.end = x.loc()
			return _CODE

		case r == '[':
			if err = x.back(); err != nil {
				break
			}
			if lval.cclass, err = charClass(x); err != nil {
				x.err = err
				return _ERROR
			}
			return _CHARCLASS

		case r == '\'' || r == '"':
			if lval.text.str, err = delimited(x, r); err != nil {
				break
			}
			lval.text.end = x.loc()
			return _STRING

		case unicode.IsSpace(r) && r != '\n':
			continue

		default:
			return int(r)
		}
		x.prevEnd = x.loc()
		x.Error(err.Error())
		return _ERROR
	}
}

func delimited(x *lexer, d rune) (string, error) {
	var rs []rune
	for {
		r, esc, err := x.nextUnesc(d)
		switch {
		case err != nil:
			return "", err
		case r == eof:
			return "", errors.New("unclosed " + string([]rune{d}))
		case r == d && !esc:
			return string(rs), nil
		}
		rs = append(rs, r)
	}
}

func ident(x *lexer) (string, error) {
	var rs []rune
	for {
		r, err := x.next()
		if err != nil {
			return "", err
		}
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' {
			return string(rs), x.back()
		}
		rs = append(rs, r)
	}
}

func code(x *lexer) (string, error) {
	var rs []rune
	var n int
	for {
		r, err := x.next()
		if err != nil {
			return "", err
		}
		if r == eof {
			return "", errors.New("unclosed {")
		}
		if r == '{' {
			n++
		}
		if r == '}' {
			if n == 0 {
				break
			}
			n--
		}
		rs = append(rs, r)
	}
	return string(rs), nil
}

func comment(x *lexer) error {
	for {
		r, err := x.next()
		if err != nil {
			return err
		}
		if r == '\n' || r == eof {
			return nil
		}
	}
}

func charClass(x *lexer) (*CharClass, error) {
	c := &CharClass{Open: x.loc()}
	if r, err := x.next(); err != nil {
		return nil, Err(c.Open, err.Error())
	} else if r != '[' {
		panic("impossible, no [")
	}

	var prev rune
	var hasPrev, span bool

	// last is the Loc just before last read rune.
	var last Loc

	// spanLoc is the location of the current span.
	// (We use type text to borrow that it implements Located.
	// However we ignore the str field.)
	var spanLoc text
loop:
	for {
		last = x.loc()
		if !span && !hasPrev {
			spanLoc.begin = x.loc()
		}
		r, esc, err := x.nextUnesc(']')
		switch {
		case err != nil:
			return nil, err

		case r == eof:
			c.Close = x.loc()
			return nil, Err(c, "unclosed [")

		case r == ']' && !esc:
			c.Close = x.loc()
			break loop

		case span:
			spanLoc.end = x.loc()
			if !hasPrev {
				return nil, Err(spanLoc, "bad span")
			}
			if prev >= r {
				return nil, Err(spanLoc, "bad span")
			}
			c.Spans = append(c.Spans, [2]rune{prev, r})
			hasPrev, span = false, false
			spanLoc.begin = spanLoc.end

		case r == '-' && !esc:
			span = true

		default:
			if !c.Neg && len(c.Spans) == 0 && r == '^' && !esc {
				c.Neg = true
				continue
			}
			if hasPrev {
				c.Spans = append(c.Spans, [2]rune{prev, prev})
				spanLoc.begin = last // in case current rune starts a span.
			}
			prev, hasPrev = r, true
		}
	}
	if span {
		spanLoc.end = last // just before closing ]
		return nil, Err(spanLoc, "bad span")
	}
	if hasPrev {
		c.Spans = append(c.Spans, [2]rune{prev, prev})
	}
	if len(c.Spans) == 0 {
		return nil, Err(c, "bad char class: empty")
	}
	return c, nil
}

var errUnknownEsc = errors.New("unknown escape sequence")

// Like next, but unescapes an escapes a rune according to Go's unescaping rules.
// The second return value is whether the rune was escaped.
func (x *lexer) nextUnesc(delim rune) (rune, bool, error) {
	switch r, err := x.next(); {
	case err != nil:
		return 0, false, err
	case r == delim:
		return r, false, nil
	case r == '\\':
		r, err = x.next()
		if err != nil {
			return 0, true, err
		}
		switch r {
		case eof:
			return eof, true, nil
		case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\':
			switch r {
			case 'a':
				r = '\a'
			case 'b':
				r = '\b'
			case 'f':
				r = '\f'
			case 'n':
				r = '\n'
			case 'r':
				r = '\r'
			case 't':
				r = '\t'
			case 'v':
				r = '\v'
			case '\\':
				r = '\\'
			}
			return r, true, nil
		case '0', '1', '2', '3', '4', '5', '6', '7':
			v, _ := oct(r)
			for i := 1; i < 3; i++ {
				r, err := x.next()
				if err != nil {
					return 0, false, err
				}
				d, ok := oct(r)
				if !ok {
					return 0, false, errUnknownEsc
				}
				v = (v << 3) | d
			}
			if v > 255 {
				return 0, false, errors.New("octal escape >255")
			}
			return v, true, nil
		case 'x', 'u', 'U':
			var n int
			switch r {
			case 'x':
				n = 2
			case 'u':
				n = 4
			case 'U':
				n = 8
			}
			var v int32
			for i := 0; i < n; i++ {
				r, err := x.next()
				if err != nil {
					return 0, false, err
				}
				d, ok := hex(r)
				if !ok {
					return 0, false, errUnknownEsc
				}
				v = (v << 4) | d
			}
			// TODO: surrogate halves are also illegal — whatever that is.
			if v > 0x10FFFF {
				return 0, false, errors.New("hex escape >0x10FFFF")
			}
			return v, true, nil
		default:
			if r == delim {
				return r, true, nil
			}
			// For character classes, allow \- as -.
			if delim == ']' && r == '-' {
				return r, true, nil
			}
			return 0, false, errUnknownEsc
		}
	default:
		return r, false, nil
	}
}

func oct(r rune) (int32, bool) {
	if '0' <= r && r <= '7' {
		return int32(r) - '0', true
	}
	return 0, false
}

func hex(r rune) (int32, bool) {
	if '0' <= r && r <= '9' {
		return int32(r) - '0', true
	}
	if 'a' <= r && r <= 'f' {
		return int32(r) - 'a' + 10, true
	}
	if 'A' <= r && r <= 'F' {
		return int32(r) - 'A' + 10, true
	}
	return 0, false
}
