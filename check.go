// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

// Check does semantic analysis of the rules,
// setting bookkeeping needed to later generate the parser,
// returning any errors encountered in order of their begin location.
func Check(grammar *Grammar) error {
	var errs Errors
	ruleMap := make(map[string]*Rule, len(grammar.Rules))
	for i := range grammar.Rules {
		r := &grammar.Rules[i]
		name := r.Name.String()
		if other := ruleMap[name]; other != nil {
			errs.add(r, "rule %s redefined", name)
		}
		ruleMap[name] = r
	}
	for i := range grammar.Rules {
		check(&grammar.Rules[i], ruleMap, &errs)
	}
	return errs.ret()
}

func check(rule *Rule, rules map[string]*Rule, errs *Errors) {
	labels := make(map[string]bool)
	rule.Expr.check(rules, labels, errs)
	for label := range labels {
		rule.Labels = append(rule.Labels, label)
	}
}

func (e *Choice) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.check(rules, labels, errs)
	}
}

func (e *Action) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *Sequence) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.check(rules, labels, errs)
	}
}

func (e *LabelExpr) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	e.Expr.check(rules, labels, errs)
	if labels[e.Label.String()] {
		errs.add(e.Label, "label %s redefined", e.Label.String())
	}
	labels[e.Label.String()] = true
}

func (e *PredExpr) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *RepExpr) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *OptExpr) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *SubExpr) check(rules map[string]*Rule, labels map[string]bool, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *Ident) check(rules map[string]*Rule, _ map[string]bool, errs *Errors) {
	r, ok := rules[e.Name.String()]
	if !ok {
		errs.add(e, "rule %s undefined", e.Name.String())
	} else {
		e.rule = r
	}
}

func (e *PredCode) check(map[string]*Rule, map[string]bool, *Errors) {}

func (e *Literal) check(map[string]*Rule, map[string]bool, *Errors) {}

func (e *CharClass) check(map[string]*Rule, map[string]bool, *Errors) {}

func (e *Any) check(map[string]*Rule, map[string]bool, *Errors) {}
