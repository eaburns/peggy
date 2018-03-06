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

	// CheckedRules are the rules successfully checked by the Check pass.
	// It contains all non-template rules and all expanded templates.
	CheckedRules []*Rule
}

// A Rule defines a production in a PEG grammar.
type Rule struct {
	Name

	// ErrorName, if non-nil, indicates that this is a named rule.
	// Errors beneath a named rule are collapsed,
	// reporting the error position as the start of the rule's parse
	// with the "want" message set to ErrorName.
	//
	// If nil, the rule is unnamed and does not collapse errors.
	ErrorName Text

	// Expr is the PEG expression matched by the rule.
	Expr Expr

	// N is the rule's unique integer within its containing Grammar.
	// It is a small integer that may be used as an array index.
	N int

	// typ is the type of the rule in the action pass.
	// typ is nil before the checkLeft pass add non-nil after.
	typ *string

	// epsilon indicates whether the rule can match the empty string.
	epsilon bool

	// Labels is the set of all label names in the rule's expression.
	Labels []Label
}

func (r *Rule) Begin() Loc  { return r.Name.Begin() }
func (r *Rule) End() Loc    { return r.Expr.End() }
func (r Rule) Type() string { return *r.typ }

// A Name is the name of a rule template.
type Name struct {
	// Name is the name of the template.
	Name Text

	// Args are the arguments or parameters of the template.
	Args []Text
}

func (n Name) Begin() Loc { return n.Name.Begin() }
func (n Name) End() Loc {
	if len(n.Args) == 0 {
		return n.Name.End()
	}
	return n.Args[len(n.Args)-1].End()
}

// A Label is the name and type of a labeled sub-expression.
type Label struct {
	// Name is the name of the label.
	Name string
	// Type is the Go type of the labeled sub-expression.
	Type string
	// N is the label's unique integer within its containing Rule.
	// It is a small integer that may be used as an array index.
	N int
}

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

	// Walk calls a function for each expression in the tree.
	// Walk stops early if the function returns false.
	Walk(func(Expr) bool)

	// substitute returns a clone of the expression
	// with all occurrences of identifiers that are keys of sub
	// substituted with the corresponding value.
	// substitute must not be called after Check,
	// because it does not update bookkeeping fields
	// that are set by the Check pass.
	substitute(sub map[string]string) Expr

	// Type returns the type of the expression in the Action Tree.
	// This is the Go type associated with the expression.
	Type() string

	// epsilon returns whether the rule can match the empty string.
	epsilon() bool

	// CanFail returns whether the node can ever fail to parse.
	// Nodes like * or ?, for example, can never fail.
	// Parents of never-fail nodes needn't emit a failure branch,
	// as it will never be called.
	CanFail() bool

	// checkLeft checks for left-recursion and sets rule types.
	checkLeft(rules map[string]*Rule, p path, errs *Errors)

	// check checks for undefined identifiers,
	// linking defined identifiers to rules;
	// and checks for type mismatches.
	check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors)
}

// A Choice is an ordered choice between expressions.
type Choice struct{ Exprs []Expr }

func (e *Choice) Begin() Loc { return e.Exprs[0].Begin() }
func (e *Choice) End() Loc   { return e.Exprs[len(e.Exprs)-1].End() }

func (e *Choice) Walk(f func(Expr) bool) {
	if !f(e) {
		return
	}
	for _, kid := range e.Exprs {
		if !f(kid) {
			return
		}
	}
}

func (e *Choice) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Exprs = make([]Expr, len(e.Exprs))
	for i, kid := range e.Exprs {
		substitute.Exprs[i] = kid.substitute(sub)
	}
	return &substitute
}

func (e *Choice) Type() string {
	t := e.Exprs[0].Type()
	for _, e := range e.Exprs[1:] {
		if e.Type() != t {
			return "interface{}"
		}
	}
	return t
}

func (e *Choice) epsilon() bool {
	for _, e := range e.Exprs {
		if e.epsilon() {
			return true
		}
	}
	return false
}

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

	// ReturnType is the go type of the value returned by the action.
	ReturnType Text

	// Labels are the labels that are in scope of this action.
	Labels []*LabelExpr
}

func (e *Action) Begin() Loc    { return e.Expr.Begin() }
func (e *Action) End() Loc      { return e.Code.End() }
func (e *Action) Type() string  { return e.ReturnType.String() }
func (e *Action) epsilon() bool { return e.Expr.epsilon() }
func (e *Action) CanFail() bool { return e.Expr.CanFail() }

func (e *Action) Walk(f func(Expr) bool) {
	if f(e) {
		f(e.Expr)
	}
}

func (e *Action) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Expr = e.Expr.substitute(sub)
	substitute.Labels = nil
	return &substitute
}

// A Sequence is a sequence of expressions.
type Sequence struct{ Exprs []Expr }

func (e *Sequence) Begin() Loc { return e.Exprs[0].Begin() }
func (e *Sequence) End() Loc   { return e.Exprs[len(e.Exprs)-1].End() }

func (e *Sequence) Walk(f func(Expr) bool) {
	if !f(e) {
		return
	}
	for _, kid := range e.Exprs {
		if !f(kid) {
			return
		}
	}
}

func (e *Sequence) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Exprs = make([]Expr, len(e.Exprs))
	for i, kid := range e.Exprs {
		substitute.Exprs[i] = kid.substitute(sub)
	}
	return &substitute
}

func (e *Sequence) Type() string {
	t := e.Exprs[0].Type()
	for _, e := range e.Exprs[1:] {
		if e.Type() != t {
			return "[]interface{}"
		}
	}
	return "[]" + t
}

func (e *Sequence) epsilon() bool {
	for _, e := range e.Exprs {
		if !e.epsilon() {
			return false
		}
	}
	return true
}

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
	// N is a small integer assigned to this label
	// that is unique within the containing Rule.
	// It is a small integer that may be used as an array index.
	N int
}

func (e *LabelExpr) Begin() Loc    { return e.Label.Begin() }
func (e *LabelExpr) End() Loc      { return e.Expr.End() }
func (e *LabelExpr) Type() string  { return e.Expr.Type() }
func (e *LabelExpr) epsilon() bool { return e.Expr.epsilon() }
func (e *LabelExpr) CanFail() bool { return e.Expr.CanFail() }

func (e *LabelExpr) Walk(f func(Expr) bool) {
	if f(e) {
		f(e.Expr)
	}
}

func (e *LabelExpr) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Expr = e.Expr.substitute(sub)
	return &substitute
}

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
func (e *PredExpr) Type() string  { return "bool" }
func (e *PredExpr) epsilon() bool { return true }
func (e *PredExpr) CanFail() bool { return e.Expr.CanFail() }

func (e *PredExpr) Walk(f func(Expr) bool) {
	if f(e) {
		f(e.Expr)
	}
}

func (e *PredExpr) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Expr = e.Expr.substitute(sub)
	return &substitute
}

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
func (e *RepExpr) Type() string  { return "[]" + e.Expr.Type() }
func (e *RepExpr) epsilon() bool { return e.Op == '*' }
func (e *RepExpr) CanFail() bool { return e.Op == '+' && e.Expr.CanFail() }

func (e *RepExpr) Walk(f func(Expr) bool) {
	if f(e) {
		f(e.Expr)
	}
}

func (e *RepExpr) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Expr = e.Expr.substitute(sub)
	return &substitute
}

// An OptExpr is an optional expression, which may or may not be matched.
type OptExpr struct {
	Expr Expr
	// Loc is the location of the ?.
	Loc Loc
}

func (e *OptExpr) Begin() Loc    { return e.Expr.Begin() }
func (e *OptExpr) End() Loc      { return e.Loc }
func (e *OptExpr) Type() string  { return "*" + e.Expr.Type() }
func (e *OptExpr) epsilon() bool { return true }
func (e *OptExpr) CanFail() bool { return false }

func (e *OptExpr) Walk(f func(Expr) bool) {
	if f(e) {
		f(e.Expr)
	}
}

func (e *OptExpr) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Expr = e.Expr.substitute(sub)
	return &substitute
}

// An Ident is an identifier referring to the name of anothe rule,
// indicating to match that rule's expression.
type Ident struct {
	Name

	// rule is the rule referred to by this identifier.
	// It is set during check.
	rule *Rule
}

func (e *Ident) Begin() Loc             { return e.Name.Begin() }
func (e *Ident) End() Loc               { return e.Name.End() }
func (e *Ident) CanFail() bool          { return true }
func (e *Ident) Walk(f func(Expr) bool) { f(e) }

func (e *Ident) Type() string {
	if e.rule == nil {
		return ""
	}
	return e.rule.Type()
}

func (e *Ident) epsilon() bool {
	if e.rule == nil {
		return false
	}
	return e.rule.epsilon
}

func (e *Ident) substitute(sub map[string]string) Expr {
	substitute := *e
	if s, ok := sub[e.Name.String()]; ok {
		substitute.Name = Name{
			Name: text{
				str:   s,
				begin: e.Name.Begin(),
				end:   e.Name.End(),
			},
		}
	}
	return &substitute
}

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
func (e *SubExpr) Type() string  { return e.Expr.Type() }
func (e *SubExpr) epsilon() bool { return e.Expr.epsilon() }
func (e *SubExpr) CanFail() bool { return e.Expr.CanFail() }

func (e *SubExpr) Walk(f func(Expr) bool) {
	if f(e) {
		f(e.Expr)
	}
}

func (e *SubExpr) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Expr = e.Expr.substitute(sub)
	return &substitute
}

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

	// Labels are the labels that are in scope of this action.
	Labels []*LabelExpr
}

func (e *PredCode) Begin() Loc             { return e.Loc }
func (e *PredCode) End() Loc               { return e.Code.End() }
func (e *PredCode) Type() string           { return "bool" }
func (e *PredCode) epsilon() bool          { return true }
func (e *PredCode) CanFail() bool          { return true }
func (e *PredCode) Walk(f func(Expr) bool) { f(e) }

func (e *PredCode) substitute(sub map[string]string) Expr {
	substitute := *e
	substitute.Labels = nil
	return &substitute
}

// A Literal matches a literal text string.
type Literal struct {
	// Text is the text to match.
	// The Begin and End locations of Text includes the ' or " delimiters,
	// but the string does not.
	Text Text
}

func (e *Literal) Begin() Loc             { return e.Text.Begin() }
func (e *Literal) End() Loc               { return e.Text.End() }
func (e *Literal) Type() string           { return "string" }
func (e *Literal) epsilon() bool          { return false }
func (e *Literal) CanFail() bool          { return true }
func (e *Literal) Walk(f func(Expr) bool) { f(e) }

func (e *Literal) substitute(sub map[string]string) Expr {
	substitute := *e
	return &substitute
}

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

func (e *CharClass) Begin() Loc             { return e.Open }
func (e *CharClass) End() Loc               { return e.Close }
func (e *CharClass) Type() string           { return "string" }
func (e *CharClass) epsilon() bool          { return false }
func (e *CharClass) CanFail() bool          { return true }
func (e *CharClass) Walk(f func(Expr) bool) { f(e) }

func (e *CharClass) substitute(sub map[string]string) Expr {
	substitute := *e
	return &substitute
}

// Any matches any rune.
type Any struct {
	// Loc is the location of the . symbol.
	Loc Loc
}

func (e *Any) Begin() Loc             { return e.Loc }
func (e *Any) End() Loc               { return Loc{Line: e.Loc.Line, Col: e.Loc.Col + 1} }
func (e *Any) Type() string           { return "string" }
func (e *Any) epsilon() bool          { return false }
func (e *Any) CanFail() bool          { return true }
func (e *Any) Walk(f func(Expr) bool) { f(e) }

func (e *Any) substitute(sub map[string]string) Expr {
	substitute := *e
	return &substitute
}
