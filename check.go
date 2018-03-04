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
	}
	if err := errs.ret(); err != nil {
		return err
	}
	grammar.CheckedRules = rules
	return nil
}

func expandTemplates(rules []Rule, errs *Errors) []*Rule {
	tmpNames := make(map[string]bool)
	for i := range rules {
		r := &rules[i]
		if len(r.Args) > 0 {
			tmpNames[r.Name.Name.String()] = true
		}
	}

	tmps := templateInvocations(rules)
	var expanded []*Rule
	for i := range rules {
		r := &rules[i]
		if len(r.Args) == 0 {
			n := r.Name.Name.String()
			if tmpNames[n] {
				errs.add(r, "rule %s redefined", n)
			} else {
				expanded = append(expanded, r)
			}
			continue
		}
		seenParams := make(map[string]bool)
		for _, param := range r.Name.Args {
			n := param.String()
			if seenParams[n] {
				errs.add(param, "parameter %s redefined", n)
			}
			seenParams[n] = true
		}
		for _, t := range tmps[r.Name.Name.String()] {
			c := *r
			if len(t.Args) != len(r.Args) {
				errs.add(t, "template %s argument count mismatch: got %d, expected %d",
					r.Name, len(t.Args), len(r.Args))
				continue
			}
			sub := make(map[string]string, len(r.Args))
			for i, arg := range t.Args {
				sub[r.Args[i].String()] = arg.String()
			}
			c.Args = t.Args
			c.Expr = r.Expr.substitute(sub)
			expanded = append(expanded, &c)
		}
	}
	return expanded
}

func templateInvocations(rules []Rule) map[string][]*Ident {
	tmps := make(map[string][]*Ident)
	for i := range rules {
		r := &rules[i]
		r.Expr.Walk(func(e Expr) bool {
			if id, ok := e.(*Ident); ok {
				if len(id.Args) > 0 {
					n := id.Name.Name.String()
					tmps[n] = append(tmps[n], id)
				}
			}
			return true
		})
	}
	return tmps
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
