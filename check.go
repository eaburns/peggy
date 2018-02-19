// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import "sort"

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
	var labels []map[string]*LabelExpr
	for i := range grammar.Rules {
		ls := check(&grammar.Rules[i], ruleMap, &errs)
		labels = append(labels, ls)
	}
	for i, ls := range labels {
		rule := &grammar.Rules[i]
		for name, expr := range ls {
			l := Label{Name: name, Type: expr.Type(), N: expr.N}
			rule.Labels = append(rule.Labels, l)
		}
	}
	return errs.ret()
}

func check(rule *Rule, rules map[string]*Rule, errs *Errors) map[string]*LabelExpr {
	labels := make(map[string]*LabelExpr)
	rule.Expr.check(rules, labels, errs)
	return labels
}

func (e *Choice) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.check(rules, labels, errs)
	}
}

func (e *Action) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	e.Expr.check(rules, labels, errs)
	for _, l := range labels {
		e.Labels = append(e.Labels, l)
	}
	sort.Slice(e.Labels, func(i, j int) bool {
		return e.Labels[i].Label.String() < e.Labels[j].Label.String()
	})
}

func (e *Sequence) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.check(rules, labels, errs)
	}
}

func (e *LabelExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	e.Expr.check(rules, labels, errs)
	if _, ok := labels[e.Label.String()]; ok {
		errs.add(e.Label, "label %s redefined", e.Label.String())
	}
	e.N = len(labels)
	labels[e.Label.String()] = e
}

func (e *PredExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *RepExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *OptExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *SubExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, errs *Errors) {
	e.Expr.check(rules, labels, errs)
}

func (e *Ident) check(rules map[string]*Rule, _ map[string]*LabelExpr, errs *Errors) {
	r, ok := rules[e.Name.String()]
	if !ok {
		errs.add(e, "rule %s undefined", e.Name.String())
	} else {
		e.rule = r
	}
}

func (e *PredCode) check(_ map[string]*Rule, labels map[string]*LabelExpr, _ *Errors) {
	for _, l := range labels {
		e.Labels = append(e.Labels, l)
	}
	sort.Slice(e.Labels, func(i, j int) bool {
		return e.Labels[i].Label.String() < e.Labels[j].Label.String()
	})
}

func (e *Literal) check(map[string]*Rule, map[string]*LabelExpr, *Errors) {}

func (e *CharClass) check(map[string]*Rule, map[string]*LabelExpr, *Errors) {}

func (e *Any) check(map[string]*Rule, map[string]*LabelExpr, *Errors) {}
