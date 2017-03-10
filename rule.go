// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import "fmt"

// Grammar is a PEG grammar.
type Grammar struct {
	// Prelude is custom code added to the beginning of the generated output.
	Prelude Text

	// Rules are the rules of the grammar.
	Rules []Rule
}

// A Rule defines a production in a PEG grammar.
type Rule struct {
	// Name is the name of the rule.
	Name Text

	// ErrorName, if non-nil, indicates that this is a named rule.
	// Errors beneath a named rule are collapsed,
	// reporting the error position as the start of the rule's parse
	// with the "want" message set to ErrorName.
	//
	// If nil, the rule is unnamed and does not collapse errors.
	ErrorName Text

	// Expr is the PEG expression matched by the rule.
	Expr Expr

	// Labels is the set of all label names in the rule's expression.
	Labels []string
}

func (e Rule) Begin() Loc { return e.Name.Begin() }
func (e Rule) End() Loc   { return e.Expr.End() }

// Text is a string of text located along with its location in the input.
type Text interface {
	Located
	// String is the text string.
	String() string
}

// Loc identifies a location in a file by its line and column numbers.
type Loc struct {
	// File is the name of the input file.
	File string
	// Line is line number of the location.
	// The first line of input is line number 1.
	Line int
	// Col is the Loc's rune offset into the line.
	// Col 0 is before the first rune on the line.
	Col int
}

// Less returns whether the receiver is earlier in the input than the argument.
func (l Loc) Less(j Loc) bool {
	if l.Line == j.Line {
		return l.Col < j.Col
	}
	return l.Line < j.Line
}

// PrettyPrint implements the pretty.PrettyPrinter interface,
// returning a simpler, one-line string form of the Loc.
func (l Loc) PrettyPrint() string { return fmt.Sprintf("Loc{%d, %d}", l.Line, l.Col) }

// Begin returns the Loc.
func (l Loc) Begin() Loc { return l }

// End returns the Loc.
func (l Loc) End() Loc { return l }

// Expr is PEG expression that matches a sequence of input runes.
type Expr interface {
	Located
	String() string

	// fullString returns the fully parenthesized string representation.
	fullString() string

	// CanFail returns whether the node can ever fail to parse.
	// Nodes like * or ?, for example, can never fail.
	// Parents of never-fail nodes needn't emit a failure branch,
	// as it will never be called.
	CanFail() bool

	// check does semantic analysis of the expression,
	// setting any bookkeeping needed for later code generation,
	// and returning the first error encountered if any.
	check(rules map[string]*Rule, labels map[string]bool, errs *Errors)
}

// A Choice is an ordered choice between expressions.
type Choice struct{ Exprs []Expr }

func (e *Choice) Begin() Loc { return e.Exprs[0].Begin() }
func (e *Choice) End() Loc   { return e.Exprs[len(e.Exprs)-1].End() }

func (e *Choice) CanFail() bool {
	// A choice node can only fail if all of its branches can fail.
	// If there is a non-failing branch, it will always return accept.
	for _, s := range e.Exprs {
		if !s.CanFail() {
			return false
		}
	}
	return true
}

// An Action is an action expression:
// a subexpression and code to run if matched.
type Action struct {
	Expr Expr
	// Code is the Go code to execute if the subexpression is matched.
	// The Begin and End locations of Code includes the { } delimiters,
	// but the string does not.
	//
	// TODO: specify the environment under which the code is run.
	Code Text
}

func (e *Action) Begin() Loc    { return e.Expr.Begin() }
func (e *Action) End() Loc      { return e.Code.End() }
func (e *Action) CanFail() bool { return e.Expr.CanFail() }

// A Sequence is a sequence of expressions.
type Sequence struct{ Exprs []Expr }

func (e *Sequence) Begin() Loc { return e.Exprs[0].Begin() }
func (e *Sequence) End() Loc   { return e.Exprs[len(e.Exprs)-1].End() }

func (e *Sequence) CanFail() bool {
	for _, s := range e.Exprs {
		if s.CanFail() {
			return true
		}
	}
	return false
}

// A LabelExpr is a labeled subexpression.
// The label can be used in actions to refer to the result of the subexperssion.
type LabelExpr struct {
	// Label is the text of the label, not including the :.
	Label Text
	Expr  Expr
}

func (e *LabelExpr) Begin() Loc    { return e.Label.Begin() }
func (e *LabelExpr) End() Loc      { return e.Expr.End() }
func (e *LabelExpr) CanFail() bool { return e.Expr.CanFail() }

// A PredExpr is a non-consuming predicate expression:
// If it succeeds (or fails, in the case of Neg),
// return success and consume no input.
// If it fails (or succeeds, in the case of Neg),
// return failure and consume no input.
// Predicate expressions allow a powerful form of lookahead.
type PredExpr struct {
	Expr Expr
	// Neg indicates that the result of the predicate is negated.
	Neg bool
	// Loc is the location of the operator, & or !.
	Loc Loc
}

func (e *PredExpr) Begin() Loc    { return e.Loc }
func (e *PredExpr) End() Loc      { return e.Expr.End() }
func (e *PredExpr) CanFail() bool { return e.Expr.CanFail() }

// A RepExpr is a repetition expression, sepecifying whether the sub-expression
// should be matched any number of times (*) or one or more times (+),
type RepExpr struct {
	// Op is one of * or +.
	Op   rune
	Expr Expr
	// Loc is the location of the operator, * or  +.
	Loc Loc
}

func (e *RepExpr) Begin() Loc    { return e.Expr.Begin() }
func (e *RepExpr) End() Loc      { return e.Loc }
func (e *RepExpr) CanFail() bool { return e.Op == '+' && e.Expr.CanFail() }

// An OptExpr is an optional expression, which may or may not be matched.
type OptExpr struct {
	Expr Expr
	// Loc is the location of the ?.
	Loc Loc
}

func (e *OptExpr) Begin() Loc    { return e.Expr.Begin() }
func (e *OptExpr) End() Loc      { return e.Loc }
func (e *OptExpr) CanFail() bool { return false }

// An Ident is an identifier referring to the name of anothe rule,
// indicating to match that rule's expression.
type Ident struct {
	Name Text

	// rule is the rule referred to by this identifier.
	// It is set during check.
	rule *Rule
}

func (e *Ident) Begin() Loc    { return e.Name.Begin() }
func (e *Ident) End() Loc      { return e.Name.End() }
func (e *Ident) CanFail() bool { return true }

// A SubExpr simply wraps an expression.
// It holds no extra information beyond tracking parentheses.
// It's purpose is to allow easily re-inserting the parentheses
// when stringifying an expression, whithout the need
// to compute precedence inversion for each subexpression.
type SubExpr struct {
	Expr
	// Open is the location of the open parenthesis.
	// Close is the location of the close parenthesis.
	Open, Close Loc
}

func (e *SubExpr) Begin() Loc    { return e.Open }
func (e *SubExpr) End() Loc      { return e.Close }
func (e *SubExpr) CanFail() bool { return e.Expr.CanFail() }

// A PredCode is a predicate code expression,
// allowing predication using a Go boolean expression.
//
// TODO: Specify the conditions under which the expression is evaluated.
type PredCode struct {
	// Code is a Go boolean expression.
	// The Begin and End locations of Code includes the { } delimiters,
	// but the string does not.
	Code Text
	// Neg indicates that the result of the predicate is negated.
	Neg bool
	// Loc is the location of the operator, & or !.
	Loc Loc
}

func (e PredCode) Begin() Loc    { return e.Loc }
func (e PredCode) End() Loc      { return e.Code.End() }
func (e PredCode) CanFail() bool { return true }

// A Literal matches a literal text string.
type Literal struct {
	// Text is the text to match.
	// The Begin and End locations of Text includes the ' or " delimiters,
	// but the string does not.
	Text Text
}

func (e *Literal) Begin() Loc    { return e.Text.Begin() }
func (e *Literal) End() Loc      { return e.Text.End() }
func (e *Literal) CanFail() bool { return true }

// A CharClass matches a single rune from a set of acceptable
// (or unacceptable if Neg) runes.
type CharClass struct {
	// Spans are rune spans accepted (or rejected) by the character class.
	// The 0th rune is always ≤ the 1st.
	// Single rune matches are a span of both the same rune.
	Spans [][2]rune

	// Neg indicates that the input must not match any in the set.
	Neg bool

	// Open and Close are the Loc of [ and ] respectively.
	Open, Close Loc
}

func (e *CharClass) Begin() Loc    { return e.Open }
func (e *CharClass) End() Loc      { return e.Close }
func (e *CharClass) CanFail() bool { return true }

// Any matches any rune.
type Any struct {
	// Loc is the location of the . symbol.
	Loc Loc
}

func (e *Any) Begin() Loc    { return e.Loc }
func (e *Any) End() Loc      { return Loc{Line: e.Loc.Line, Col: e.Loc.Col + 1} }
func (e *Any) CanFail() bool { return true }
