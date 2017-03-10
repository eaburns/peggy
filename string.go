// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// String returns the string representation of the rules.
// The output contains no comments or whitespace,
// except for a single space, " ",
// separating sub-exprsessions of a sequence,
// and on either side of <-.
func String(rules []Rule) string {
	var s string
	for _, r := range rules {
		if s != "" {
			s += "\n"
		}
		s += r.String()
	}
	return s
}

// String returns the string representation of a rule.
// The output contains no comments or whitespace,
// except for a single space, " ",
// separating sub-exprsessions of a sequence,
// and on either side of <-.
func (r *Rule) String() string {
	var name string
	if r.ErrorName != nil {
		name = " " + strconv.Quote(r.ErrorName.String())
	}
	return r.Name.String() + name + " <- " + r.Expr.String()
}

func (e *Choice) String() string {
	s := e.Exprs[0].String()
	for _, sub := range e.Exprs[1:] {
		s += "/" + sub.String()
	}
	return s
}

func (e *Action) String() string {
	return e.Expr.String() + "{" + e.Code.String() + "}"
}

func (e *Sequence) String() string {
	s := e.Exprs[0].String()
	for _, sub := range e.Exprs[1:] {
		s += " " + sub.String()
	}
	return s
}

func (e *LabelExpr) String() string {
	return e.Label.String() + ":" + e.Expr.String()
}

func (e *PredExpr) String() string {
	s := "&"
	if e.Neg {
		s = "!"
	}
	return s + e.Expr.String()
}

func (e *RepExpr) String() string {
	return e.Expr.String() + string([]rune{e.Op})
}

func (e *OptExpr) String() string {
	return e.Expr.String() + "?"
}

func (e *SubExpr) String() string {
	return "(" + e.Expr.String() + ")"
}

func (e *Ident) String() string {
	return e.Name.String()
}

func (e *PredCode) String() string {
	s := "&{"
	if e.Neg {
		s = "!{"
	}
	return s + e.Code.String() + "}"
}

func (e *Literal) String() string {
	return fmt.Sprintf(`%q`, e.Text)
}

func (e *CharClass) String() string {
	s := "["
	if e.Neg {
		s += "^"
	}
	for _, sp := range e.Spans {
		if sp[0] == sp[1] {
			s += charClassEsc(sp[0])
		} else {
			s += charClassEsc(sp[0]) + "-" + charClassEsc(sp[1])
		}
	}
	return s + "]"
}

func charClassEsc(r rune) string {
	switch r {
	case '-':
		return `\-`
	case ']':
		return `\]`
	}
	s := strconv.QuoteRune(r)
	return strings.TrimPrefix(strings.TrimSuffix(s, "'"), "'")
}

func (e *Any) String() string { return "." }

// FullString returns the fully parenthesized string representation of the rules.
// The output contains no comments or whitespace,
// except for a single space, " ",
// separating sub-exprsessions of a sequence,
// and on either side of <-.
func FullString(rules []Rule) string {
	var s string
	for _, r := range rules {
		if s != "" {
			s += "\n"
		}

		var name string
		if r.ErrorName != nil {
			name = " " + strconv.Quote(r.ErrorName.String())
		}
		s += fmt.Sprintf("%s%s <- %s", r.Name, name, r.Expr.fullString())
	}
	return s
}

func (e *Choice) fullString() string {
	s := strings.Repeat("(", len(e.Exprs)-1) + e.Exprs[0].fullString()
	for _, sub := range e.Exprs[1:] {
		s += "/" + sub.fullString() + ")"
	}
	return s
}

func (e *Action) fullString() string {
	return fmt.Sprintf("(%s{%s})", e.Expr.fullString(), e.Code)
}

func (e *Sequence) fullString() string {
	s := strings.Repeat("(", len(e.Exprs)-1) + e.Exprs[0].fullString()
	for _, sub := range e.Exprs[1:] {
		s += " " + sub.fullString() + ")"
	}
	return s
}

func (e *LabelExpr) fullString() string {
	return fmt.Sprintf("(%s:%s)", e.Label.String(), e.Expr.fullString())
}

func (e *PredExpr) fullString() string {
	if e.Neg {
		return fmt.Sprintf("(!%s)", e.Expr.fullString())
	}
	return fmt.Sprintf("(&%s)", e.Expr.fullString())
}

func (e *RepExpr) fullString() string {
	return fmt.Sprintf("(%s%c)", e.Expr.fullString(), e.Op)
}

func (e *OptExpr) fullString() string {
	return "(" + e.Expr.fullString() + "?)"
}

func (e *Ident) fullString() string { return "(" + e.String() + ")" }

func (e *PredCode) fullString() string { return "(" + e.String() + ")" }

func (e *Literal) fullString() string { return "(" + e.String() + ")" }

func (e *CharClass) fullString() string { return "(" + e.String() + ")" }

func (e *Any) fullString() string { return "(" + e.String() + ")" }
