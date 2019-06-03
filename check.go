// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"sort"
)

// Check does semantic analysis of the rules,
// setting bookkeeping needed to later generate the parser,
// returning any errors encountered in order of their begin location.
func Check(grammar *Grammar) error {
	var errs Errors
	rules := expandTemplates(grammar.Rules, &errs)
	ruleMap := make(map[string]*Rule, len(rules))
	for i, r := range rules {
		r.N = i
		name := r.Name.String()
		if other := ruleMap[name]; other != nil {
			errs.add(r, "rule %s redefined", name)
		}
		ruleMap[name] = r
	}

	var p path
	for _, r := range rules {
		r.checkLeft(ruleMap, p, &errs)
	}

	var labels []map[string]*LabelExpr
	for _, r := range rules {
		ls := check(r, ruleMap, &errs)
		labels = append(labels, ls)
	}
	for i, ls := range labels {
		rule := rules[i]
		for name, expr := range ls {
			l := Label{Name: name, Type: expr.Type(), N: expr.N}
			rule.Labels = append(rule.Labels, l)
		}
		sort.Slice(rule.Labels, func(i, j int) bool {
			return rule.Labels[i].N < rule.Labels[j].N
		})
	}
	if err := errs.ret(); err != nil {
		return err
	}
	grammar.CheckedRules = rules
	return nil
}

func expandTemplates(ruleDefs []Rule, errs *Errors) []*Rule {
	var expanded, todo []*Rule
	tmplNames := make(map[string]*Rule)
	for i := range ruleDefs {
		r := &ruleDefs[i]
		if len(r.Name.Args) > 0 {
			seenParams := make(map[string]bool)
			for _, param := range r.Name.Args {
				n := param.String()
				if seenParams[n] {
					errs.add(param, "parameter %s redefined", n)
				}
				seenParams[n] = true
			}
			tmplNames[r.Name.Name.String()] = r
		} else {
			expanded = append(expanded, r)
			todo = append(todo, r)
		}
	}

	seen := make(map[string]bool)
	for i := 0; i < len(todo); i++ {
		for _, invok := range invokedTemplates(todo[i]) {
			if seen[invok.Name.String()] {
				continue
			}
			seen[invok.Name.String()] = true
			tmpl := tmplNames[invok.Name.Name.String()]
			if tmpl == nil {
				continue // undefined template, error reported elsewhere
			}
			exp := expand1(tmpl, invok, errs)
			if exp == nil {
				continue // error expanding, error reported elsewhere
			}
			todo = append(todo, exp)
			expanded = append(expanded, exp)
		}
	}
	return expanded
}

func expand1(tmpl *Rule, invok *Ident, errs *Errors) *Rule {
	if len(invok.Args) != len(tmpl.Args) {
		errs.add(invok, "template %s argument count mismatch: got %d, expected %d",
			tmpl.Name, len(invok.Args), len(tmpl.Args))
		return nil
	}
	copy := *tmpl
	sub := make(map[string]string, len(tmpl.Args))
	for i, arg := range invok.Args {
		sub[tmpl.Args[i].String()] = arg.String()
	}
	copy.Args = invok.Args
	copy.Expr = tmpl.Expr.substitute(sub)
	return &copy
}

func invokedTemplates(r *Rule) []*Ident {
	var tmpls []*Ident
	r.Expr.Walk(func(e Expr) bool {
		if id, ok := e.(*Ident); ok {
			if len(id.Args) > 0 {
				tmpls = append(tmpls, id)
			}
		}
		return true
	})
	return tmpls
}

type path struct {
	stack []*Rule
	seen  map[*Rule]bool
}

func (p *path) push(r *Rule) bool {
	if p.seen == nil {
		p.seen = make(map[*Rule]bool)
	}
	if p.seen[r] {
		return false
	}
	p.stack = append(p.stack, r)
	p.seen[r] = true
	return true
}

func (p *path) pop() {
	p.stack = p.stack[:len(p.stack)]
}

func (p *path) cycle(r *Rule) []*Rule {
	for i := len(p.stack) - 1; i >= 0; i-- {
		if p.stack[i] == r {
			return append(p.stack[i:], r)
		}
	}
	panic("no cycle")
}

func cycleString(rules []*Rule) string {
	var s string
	for _, r := range rules {
		if s != "" {
			s += ", "
		}
		s += r.Name.String()
	}
	return s
}

func (r *Rule) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	if r.typ != nil {
		return
	}
	if !p.push(r) {
		cycle := p.cycle(r)
		errs.add(cycle[0], "left-recursion: %s", cycleString(cycle))
		for _, r := range cycle {
			r.typ = new(string)
		}
		return
	}
	r.Expr.checkLeft(rules, p, errs)
	t := r.Expr.Type()
	r.typ = &t
	r.epsilon = r.Expr.epsilon()
	p.pop()
}

func (e *Choice) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.checkLeft(rules, p, errs)
	}
}

func (e *Action) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	e.Expr.checkLeft(rules, p, errs)
}

func (e *Sequence) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.checkLeft(rules, p, errs)
		if !sub.epsilon() {
			break
		}
	}
}

func (e *LabelExpr) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	e.Expr.checkLeft(rules, p, errs)
}

func (e *PredExpr) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	e.Expr.checkLeft(rules, p, errs)
}

func (e *RepExpr) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	e.Expr.checkLeft(rules, p, errs)
}

func (e *OptExpr) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	e.Expr.checkLeft(rules, p, errs)
}

func (e *Ident) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	if e.rule = rules[e.Name.String()]; e.rule != nil {
		e.rule.checkLeft(rules, p, errs)
	}
}

func (e *SubExpr) checkLeft(rules map[string]*Rule, p path, errs *Errors) {
	e.Expr.checkLeft(rules, p, errs)
}

func (e *PredCode) checkLeft(rules map[string]*Rule, p path, errs *Errors) {}

func (e *Literal) checkLeft(rules map[string]*Rule, p path, errs *Errors) {}

func (e *CharClass) checkLeft(rules map[string]*Rule, p path, errs *Errors) {}

func (e *Any) checkLeft(rules map[string]*Rule, p path, errs *Errors) {}

func check(rule *Rule, rules map[string]*Rule, errs *Errors) map[string]*LabelExpr {
	labels := make(map[string]*LabelExpr)
	rule.Expr.check(rules, labels, true, errs)
	return labels
}

func (e *Choice) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.check(rules, labels, valueUsed, errs)
	}

	t := e.Type()
	for _, sub := range e.Exprs {
		// Check types, but if either type is "",
		// it's from a previous error; don't report again.
		if got := sub.Type(); got != "" && *genActions && valueUsed && got != t && got != "" && t != "" {
			errs.add(sub, "type mismatch: got %s, expected %s", got, t)
		}
	}
}

func (e *Action) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	e.Expr.check(rules, labels, false, errs)
	for _, l := range labels {
		e.Labels = append(e.Labels, l)
	}
	sort.Slice(e.Labels, func(i, j int) bool {
		return e.Labels[i].Label.String() < e.Labels[j].Label.String()
	})
}

// BUG: figure out what to do about sequence types.
func (e *Sequence) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	for _, sub := range e.Exprs {
		sub.check(rules, labels, valueUsed, errs)
	}
	t := e.Exprs[0].Type()
	for _, sub := range e.Exprs {
		if got := sub.Type(); *genActions && valueUsed && got != t && got != "" && t != "" {
			errs.add(sub, "type mismatch: got %s, expected %s", got, t)
		}
	}
}

func (e *LabelExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	e.Expr.check(rules, labels, true, errs)
	if _, ok := labels[e.Label.String()]; ok {
		errs.add(e.Label, "label %s redefined", e.Label.String())
	}
	e.N = len(labels)
	labels[e.Label.String()] = e
}

func (e *PredExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	e.Expr.check(rules, labels, false, errs)
}

func (e *RepExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	e.Expr.check(rules, labels, valueUsed, errs)
}

func (e *OptExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	e.Expr.check(rules, labels, valueUsed, errs)
}

func (e *SubExpr) check(rules map[string]*Rule, labels map[string]*LabelExpr, valueUsed bool, errs *Errors) {
	e.Expr.check(rules, labels, valueUsed, errs)
}

func (e *Ident) check(rules map[string]*Rule, _ map[string]*LabelExpr, _ bool, errs *Errors) {
	r, ok := rules[e.Name.String()]
	if !ok {
		errs.add(e, "rule %s undefined", e.Name.String())
	} else {
		e.rule = r
	}
}

func (e *PredCode) check(_ map[string]*Rule, labels map[string]*LabelExpr, _ bool, _ *Errors) {
	for _, l := range labels {
		e.Labels = append(e.Labels, l)
	}
	sort.Slice(e.Labels, func(i, j int) bool {
		return e.Labels[i].Label.String() < e.Labels[j].Label.String()
	})
}

func (e *Literal) check(map[string]*Rule, map[string]*LabelExpr, bool, *Errors) {}

func (e *CharClass) check(map[string]*Rule, map[string]*LabelExpr, bool, *Errors) {}

func (e *Any) check(map[string]*Rule, map[string]*LabelExpr, bool, *Errors) {}
