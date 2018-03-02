// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"bytes"
	"errors"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"reflect"
	"strconv"
	"text/template"
)

// Generate generates a parser for the rules,
// using a default Config:
// 	Config{Prefix: "_"}
func Generate(w io.Writer, file string, grammar *Grammar) error {
	return Config{Prefix: "_"}.Generate(w, file, grammar)
}

// A Config specifies code generation options.
type Config struct {
	Prefix string
}

// Generate generates a parser for the rules.
func (c Config) Generate(w io.Writer, file string, gr *Grammar) error {
	b := bytes.NewBuffer(nil)
	if err := writePrelude(b, file, gr); err != nil {
		return err
	}
	if err := writeDecls(b, c, gr); err != nil {
		return err
	}
	for i := range gr.Rules {
		r := &gr.Rules[i]
		if err := writeRule(b, c, r); err != nil {
			return err
		}
	}
	return gofmt(w, b.String())
}

func gofmt(w io.Writer, s string) error {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, "", s, parser.ParseComments)
	if err != nil {
		io.WriteString(os.Stderr, s)
		io.WriteString(w, s)
		return err
	}
	if err := format.Node(w, fset, root); err != nil {
		io.WriteString(w, s)
		return err
	}
	return nil
}

func writePrelude(w io.Writer, file string, gr *Grammar) error {
	if gr.Prelude == nil {
		return nil
	}
	_, err := io.WriteString(w, gr.Prelude.String())
	return err
}

func writeDecls(w io.Writer, c Config, gr *Grammar) error {
	tmp, err := template.New("Decls").Parse(declsTemplate)
	if err != nil {
		return err
	}
	return tmp.Execute(w, map[string]interface{}{
		"Config":  c,
		"Grammar": gr,
	})
}

func writeRule(w io.Writer, c Config, r *Rule) error {
	funcs := map[string]interface{}{
		"gen":   gen,
		"quote": strconv.Quote,
		"makeAcceptState": func(r *Rule) state {
			return state{
				Config:      c,
				Rule:        r,
				n:           new(int),
				AcceptsPass: true,
			}
		},
		"makeNodeState": func(r *Rule) state {
			return state{
				Config:   c,
				Rule:     r,
				n:        new(int),
				NodePass: true,
			}
		},
		"makeFailState": func(r *Rule) state {
			return state{
				Config:   c,
				Rule:     r,
				n:        new(int),
				FailPass: true,
			}
		},
		"makeActionState": func(r *Rule) state {
			return state{
				Config:     c,
				Rule:       r,
				n:          new(int),
				ActionPass: true,
			}
		},
	}
	data := map[string]interface{}{
		"Config": c,
		"Rule":   r,
	}
	tmp, err := template.New("rule").Parse(ruleTemplate)
	if err != nil {
		return err
	}
	for _, ts := range [][2]string{
		{"ruleAccepts", ruleAccepts},
		{"ruleNode", ruleNode},
		{"ruleFail", ruleFail},
		{"stringLabels", stringLabels},
		{"ruleAction", ruleAction},
	} {
		name, text := ts[0], ts[1]
		tmp, err = tmp.New(name).Funcs(funcs).Parse(text)
		if err != nil {
			return err
		}
	}
	return tmp.ExecuteTemplate(w, "rule", data)
}

type state struct {
	Config
	Rule *Rule
	Expr Expr
	Fail string
	n    *int
	// AcceptsPass indicates whether to generate the accepts pass.
	AcceptsPass bool
	// NodePass indicates whether to generate the node pass.
	NodePass bool
	// FailPass indicates whether to generate the error pass.
	FailPass bool
	// ActionPass indicates whether to generate the action pass.
	ActionPass bool
}

func (s state) id(str string) string {
	(*s.n)++
	return str + strconv.Itoa(*s.n-1)
}

var globalTemplates = [][2]string{
	{"charClassCondition", charClassCondition},
}

func addGlobalTemplates(tmp *template.Template) error {
	for _, p := range globalTemplates {
		var err error
		if tmp, err = tmp.New(p[0]).Parse(p[1]); err != nil {
			return err
		}
	}
	return nil
}

// A note on formatting in Expr templates
//
// gofmt properly fixes any horizontal spacing issues.
// However, while it eliminates duplicate empty lines,
// it does not eliminate empty lines.
// For example, it will convert a sequence of 2 or more empty lines
// into a single empty line, but it will not remove the empty line.
// So it's important to handle newlines propertly
// to maintain a nice, consistent formatting.
//
// There are two rules:
// 	1) Templates must end with a newline, or the codegen will be invalid.
// 	2) Templates should not begin with an newline, or the codegen will be ugly.

var declsTemplate = `
	{{$pre := $.Config.Prefix -}}

	const (
		{{range $r := $.Grammar.Rules -}}
			{{$pre}}{{$r.Name.String}} int = {{$r.N}}
		{{end}}
		{{$pre}}N int = {{len $.Grammar.Rules}}
	)

	type {{$pre}}Parser struct {
		text string
		deltaPos [][{{$pre}}N]int32
		deltaErr [][{{$pre}}N]int32
		node map[{{$pre}}key]*peg.Node
		fail map[{{$pre}}key]*peg.Fail
		act map[{{$pre}}key]interface{}
		lastFail int
		data interface{}
	}

	type {{$pre}}key struct {
		start int
		name string
	}

	func {{$pre}}NewParser(text string) *{{$pre}}Parser {
		return &{{$pre}}Parser{
			text: text,
			deltaPos: make([][{{$pre}}N]int32, len(text)+1),
			deltaErr: make([][{{$pre}}N]int32, len(text)+1),
			node: make(map[{{$pre}}key]*peg.Node),
			fail: make(map[{{$pre}}key]*peg.Fail),
			act: make(map[{{$pre}}key]interface{}),
		}
	}

	func {{$pre}}max(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	func {{$pre}}memoize(parser *{{$pre}}Parser, rule, start, pos, perr int) (int, int) {
		parser.lastFail = perr
		derr := perr - start
		parser.deltaErr[start][rule] = int32(derr+1)
		if pos >= 0 {
			dpos := pos - start
			parser.deltaPos[start][rule] = int32(dpos + 1)
			return dpos, derr
		}
		parser.deltaPos[start][rule] = -1
		return -1, derr
	}

	func {{$pre}}memo(parser *{{$pre}}Parser, rule, start int) (int, int, bool) {
		dp := parser.deltaPos[start][rule]
		if dp == 0 {
			return 0, 0, false
		}
		if dp > 0 {
			dp--
		}
		de := parser.deltaErr[start][rule] - 1
		return int(dp), int(de), true
	}

	func {{$pre}}accept(parser *{{$pre}}Parser, f func(*{{$pre}}Parser, int) (int, int), pos, perr *int) bool {
		dp, de := f(parser, *pos)
		*perr = _max(*perr, *pos+de)
		if dp < 0 {
			return false
		}
		*pos += dp
		return true
	}

	func {{$pre}}node(parser *{{$pre}}Parser, f func(*{{$pre}}Parser, int) (int, *peg.Node), node *peg.Node, pos *int) bool {
		p, kid := f(parser, *pos)
		if kid == nil {
			return false
		}
		node.Kids = append(node.Kids, kid)
		*pos = p
		return true
	}

	func {{$pre}}fail(parser *{{$pre}}Parser, f func(*{{$pre}}Parser, int, int) (int, *peg.Fail), errPos int, node *peg.Fail, pos *int) bool {
		p, kid := f(parser, *pos, errPos)
		if kid.Want != "" || len(kid.Kids) > 0 {
			node.Kids = append(node.Kids, kid)
		}
		if p < 0 {
			return false
		}
		*pos = p
		return true
	}

	func {{$pre}}next(parser *{{$pre}}Parser, pos int) (rune, int) {
		r, w := peg.DecodeRuneInString(parser.text[pos:])
		return r, w
	}

	func {{$pre}}sub(parser *{{$pre}}Parser, start, end int, kids []*peg.Node) *peg.Node {
		node := &peg.Node{
			Text: parser.text[start:end],
			Kids: make([]*peg.Node, len(kids)),
		}
		copy(node.Kids, kids)
		return node
	}

	func {{$pre}}leaf(parser *{{$pre}}Parser, start, end int) *peg.Node {
		return &peg.Node{Text: parser.text[start:end]}
	}
`

func gen(parentState state, expr Expr, fail string) (string, error) {
	t := reflect.TypeOf(expr)
	tmpString, ok := templates[reflect.TypeOf(expr)]
	if !ok {
		return "", errors.New("gen not found: " + t.String())
	}
	funcs := map[string]interface{}{
		"quote":     strconv.Quote,
		"quoteRune": strconv.QuoteRune,
		"id":        parentState.id,
		"gen":       gen,
		"last":      func(i int, exprs []Expr) bool { return i == len(exprs)-1 },
	}
	tmp, err := template.New(t.String()).Funcs(funcs).Parse(tmpString)
	if err != nil {
		return "", err
	}
	if err := addGlobalTemplates(tmp); err != nil {
		return "", err
	}
	b := bytes.NewBuffer(nil)
	state := parentState
	state.Expr = expr
	state.Fail = fail
	err = tmp.Execute(b, state)
	return b.String(), err
}

// templates contains a mapping from Expr types to their templates.
// These templates parse the input text and compute
// for each <rule, pos> pair encountered by the parse,
// the position immediately following the text accepted by the rule,
// or the position of the furthest error encountered by the rule.
//
// When generating the parse tree pass,
// the templates also add peg.Nodes to the kids slice.
//
// Variables for use by the templates:
// 	parser is the *Parser.
// 		parser.text is the input text.
// 	pos is the byte offset into parser.text of where to begin parsing.
// 		If the Expr fails to parse, pos must be set to the position of the error.
// 		If if the Expr succeeds to parse, pos must be set
// 		to the position just after the accepted text.
//
// On the accepts pass these variables are also defined:
// 	perr is the position of the max error position found so far.
// 		It is only defined if Rule.Expr.CanFail.
// 		It is initialized to -1 at the beginning of the parse.
// 		It is updated by Choice nodes when branches fail,
// 		and by rules when their entire parse fails.
// 	ok is a scratch boolean variable.
// 		It may be either true or false before and after each Expr template.
// 		Each template that wants to use ok must set it before using it.
//
// On the node tree pass these variables are also defined:
// 	node is the *peg.Node of the Rule being parsed.
//
// On the action tree pass these variables are also defined:
// 	node is an interface{} containing the current action tree node value.
//
// On the fail tree pass these variables are also defined:
// 	failure is the *peg.Fail of the Rule being parsed.
// 	errPos is the position before which Fail nodes are not generated.
var templates = map[reflect.Type]string{
	reflect.TypeOf(&Choice{}):    choiceTemplate,
	reflect.TypeOf(&Action{}):    actionTemplate,
	reflect.TypeOf(&Sequence{}):  sequenceTemplate,
	reflect.TypeOf(&LabelExpr{}): labelExprTemplate,
	reflect.TypeOf(&PredExpr{}):  predExprTemplate,
	reflect.TypeOf(&RepExpr{}):   repExprTemplate,
	reflect.TypeOf(&OptExpr{}):   optExprTemplate,
	reflect.TypeOf(&SubExpr{}):   subExprTemplate,
	reflect.TypeOf(&PredCode{}):  predCodeTemplate,
	reflect.TypeOf(&Ident{}):     identTemplate,
	reflect.TypeOf(&Literal{}):   literalTemplate,
	reflect.TypeOf(&Any{}):       anyTemplate,
	reflect.TypeOf(&CharClass{}): charClassTemplate,
}

var ruleTemplate = `
	{{template "ruleAccepts" $}}
	{{template "ruleNode" $}}
	{{template "ruleFail" $}}
	{{template "ruleAction" $}}
`

var stringLabels = `
	{{- if $.Rule.Labels -}}
		var labels [{{len $.Rule.Labels}}]string
		labels = labels
	{{- end -}}
`

var ruleAccepts = `
	{{$pre := $.Config.Prefix -}}
	{{- $name := $.Rule.Name.String -}}
	func {{$pre}}{{$name}}Accepts(parser *{{$pre}}Parser, start int) (deltaPos, deltaErr int) {
		{{- template "stringLabels" $}}
		if dp, de, ok := {{$pre}}memo(parser, {{$pre}}{{$name}}, start); ok {
			return dp, de
		}
		pos, perr := start, -1
		{{gen (makeAcceptState $.Rule) $.Rule.Expr "fail" -}}

		{{if $.Rule.ErrorName -}}
			perr = start
		{{end -}}
		return {{$pre}}memoize(parser, {{$pre}}{{$name}}, start, pos, perr)
	{{if $.Rule.Expr.CanFail -}}
	fail:
		return {{$pre}}memoize(parser, {{$pre}}{{$name}}, start, -1, perr)
	{{end -}}
	}
`

var ruleNode = `
	{{$pre := $.Config.Prefix -}}
	{{- $name := $.Rule.Name.String -}}
	func {{$pre}}{{$name}}Node(parser *{{$pre}}Parser, start int) (int, *peg.Node) {
		{{- template "stringLabels" $}}
		dp := parser.deltaPos[start][{{$pre}}{{$name}}]
		if dp < 0 {
			return -1, nil
		}
		key := {{$pre}}key{start: start, name: {{quote $name}}}
		node := parser.node[key]
		if node != nil {
			return start + int(dp - 1), node
		}
		pos := start
		node = &peg.Node{Name: {{quote $name}}}
		{{gen (makeNodeState $.Rule) $.Rule.Expr "fail" -}}

		node.Text = parser.text[start:pos]
		parser.node[key] = node
		return pos, node
	{{if $.Rule.Expr.CanFail -}}
	fail:
		return -1, nil
	{{end -}}
	}
`

var ruleFail = `
	{{$pre := $.Config.Prefix -}}
	{{- $name := $.Rule.Name.String -}}
	func {{$pre}}{{$name}}Fail(parser *{{$pre}}Parser, start, errPos int) (int, *peg.Fail) {
		{{- template "stringLabels" $}}
		if start > parser.lastFail {
			return -1, &peg.Fail{}
		}
		dp := parser.deltaPos[start][{{$pre}}{{$name}}]
		de := parser.deltaErr[start][{{$pre}}{{$name}}]
		if start + int(de - 1) < errPos {
			if dp > 0 {
				return start + int(dp - 1), &peg.Fail{}
			}
			return -1, &peg.Fail{}
		}
		key := {{$pre}}key{start: start, name: {{quote $name}}}
		failure := parser.fail[key]
		if dp < 0 && failure != nil {
			return -1, failure
		}
		if dp > 0 && failure != nil {
			return start + int(dp - 1), failure
		}
		pos := start
		failure = &peg.Fail{
			Name: {{quote $name}},
			Pos: int(start),
		}
		{{gen (makeFailState $.Rule) $.Rule.Expr "fail" -}}

		{{if $.Rule.ErrorName -}}
			failure.Kids = nil
		{{end -}}
		parser.fail[key] = failure
		return pos, failure
	{{if $.Rule.Expr.CanFail -}}
	fail:
		{{if $.Rule.ErrorName -}}
			failure.Kids = nil
			failure.Want = {{quote $.Rule.ErrorName.String}}
		{{end -}}
		parser.fail[key] = failure
		return -1, failure
	{{end -}}
	}
`

var ruleAction = `
	{{$pre := $.Config.Prefix -}}
	{{- $name := $.Rule.Name.String -}}
	{{- $type := $.Rule.Expr.Type -}}
	func {{$pre}}{{$name}}Action(parser *{{$pre}}Parser, start int) (int, *{{$type}}) {
		{{- template "stringLabels" $}}
		{{if $.Rule.Labels -}}
			{{range $l := $.Rule.Labels -}}
				var label{{$l.N}} {{$l.Type}}
				label{{$l.N}} = label{{$l.N}}
			{{end}}
		{{- end -}}
		dp := parser.deltaPos[start][{{$pre}}{{$name}}]
		if dp < 0 {
			return -1, nil
		}
		key := {{$pre}}key{start: start, name: {{quote $name}}}
		node := parser.act[key]
		if node != nil {
			n, _ := node.({{$type}})
			return start + int(dp - 1), &n
		}
		pos := start
		{{gen (makeActionState $.Rule) $.Rule.Expr "fail" -}}

		parser.act[key] = node
		{
			n, _ := node.({{$type}})
			return pos,  &n
		}
	{{if $.Rule.Expr.CanFail -}}
	fail:
		return -1, nil
	{{end -}}
	}
`

var choiceTemplate = `// {{$.Expr.String}}
{
	{{- $ok := id "ok" -}}
	{{- $nkids := id "nkids" -}}
	{{- $pos0 := id "pos" -}}
	{{$pos0}} := pos
	{{if $.NodePass -}}
		{{$nkids}} := len(node.Kids)
	{{end -}}
	{{- range $i, $subExpr := $.Expr.Exprs -}}
		{{- $fail := id "fail" -}}
		{{gen $ $subExpr $fail -}}

		{{if $subExpr.CanFail -}}
			goto {{$ok}}
			{{$fail}}:
				{{if $.NodePass -}}
					node.Kids = node.Kids[:{{$nkids}}]
				{{end -}}
				pos = {{$pos0}}
			{{if last $i $.Expr.Exprs -}}
				goto {{$.Fail}}
			{{end -}}
		{{end -}}
	{{end -}}
	{{$ok}}:
}
`

var actionTemplate = `// action
	{{gen $ $.Expr.Expr $.Fail -}}
	{{if $.ActionPass -}}
	{{/* TODO: don't put the func in the scope of the rule. */ -}}
	node = func(
		{{- if $.Expr.Labels -}}
			{{range $lexpr := $.Expr.Labels -}}
				{{$lexpr.Label}} {{$lexpr.Type}},
			{{- end -}}
		{{- end -}}) {{$.Expr.Type}} { {{$.Expr.Code}} }(
		{{- if $.Expr.Labels -}}
			{{range $lexpr := $.Expr.Labels -}}
				label{{$lexpr.N}},
			{{- end -}}
		{{- end -}}
	)
	{{end -}}
`

var sequenceTemplate = `// {{$.Expr.String}}
	{{$nodes := id "nodes" -}}
	{{if $.ActionPass -}}
		{
		var {{$nodes}} {{$.Expr.Type}}
	{{end -}}
	{{range $i, $subExpr := $.Expr.Exprs -}}
		{{gen $ $subExpr $.Fail -}}
		{{if $.ActionPass -}}
			{
				n, _ := node.({{$subExpr.Type}})
				{{if last $i $.Expr.Exprs -}}
					node = append({{$nodes}}, n)
				{{else -}}
					{{$nodes}} = append({{$nodes}}, n)
				{{end -}}
			}
		{{end -}}	
	{{end -}}
	{{if $.ActionPass -}}
		}
	{{end -}}
`

var labelExprTemplate = `// {{$.Expr.String}}
	{{$name := $.Expr.Label.String -}}
	{{- $pos0 := id "pos" -}}
	{{- $subExpr := $.Expr.Expr -}}
	{
		{{$pos0}} := pos
		{{gen $ $subExpr $.Fail -}}
		labels[{{$.Expr.N}}] = parser.text[{{$pos0}}:pos]
		{{if $.ActionPass -}}
			{
				n, _ := node.({{$.Expr.Type}})
				label{{$.Expr.N}} = n
			}
		{{end -}}
	}
`

var predExprTemplate = `// {{$.Expr.String}}
{
	{{- $pre := $.Config.Prefix -}}
	{{- $ok := id "ok" -}}
	{{- $subExpr := $.Expr.Expr -}}
	{{- $pos0 := id "pos" -}}
	{{- $nkids := id "nkids" -}}
	{{- $perr0 := id "perr" -}}
	{{$pos0}} := pos
	{{if $.AcceptsPass -}}
		{{$perr0}} := perr
	{{else if $.NodePass -}}
		{{$nkids}} := len(node.Kids)
	{{else if $.FailPass -}}
		{{$nkids}} := len(failure.Kids)
	{{end -}}

	{{- if $.Expr.Neg -}}
		{{gen $ $subExpr $ok -}}
		pos = {{$pos0}}
		{{if $.NodePass -}}
			node.Kids = node.Kids[:{{$nkids}}]
		{{else if $.AcceptsPass -}}
			perr = {{$pre}}max({{$perr0}}, pos)
		{{else if $.FailPass -}}
			failure.Kids = failure.Kids[:{{$nkids}}]
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos: int(pos),
					Want: {{quote $.Expr.String}},
				})
			}
		{{else if $.ActionPass -}}
			node = false
		{{end -}}
		goto {{$.Fail}}
	{{else -}}
		{{- $fail := id "fail" -}}
		{{gen $ $subExpr $fail -}}
		goto {{$ok}}
		{{$fail}}:
			pos = {{$pos0}}
			{{if $.AcceptsPass -}}
				perr = {{$pre}}max({{$perr0}}, pos)
			{{else if $.FailPass -}}
				failure.Kids = failure.Kids[:{{$nkids}}]
				if pos >= errPos {
					failure.Kids = append(failure.Kids, &peg.Fail{
						Pos: int(pos),
						Want: {{quote $.Expr.String}},
					})
				}
			{{else if $.ActionPass -}}
				node = false
			{{end -}}
			goto {{$.Fail}}
	{{end -}}

	{{$ok}}:
	pos = {{$pos0}}
	{{if $.AcceptsPass -}}
		perr = {{$perr0}}
	{{else if $.NodePass -}}
		node.Kids = node.Kids[:{{$nkids}}]
	{{else if $.FailPass -}}
		failure.Kids = failure.Kids[:{{$nkids}}]
	{{else if $.ActionPass -}}
		node = true
	{{end -}}
}
`

var repExprTemplate = `// {{$.Expr.String}}
	{{$nkids := id "nkids" -}}
	{{$pos0 := id "pos" -}}
	{{$nodes := id "nodes" -}}
	{{- $fail := id "fail" -}}
	{{- $subExpr := $.Expr.Expr -}}
	{{if $.ActionPass -}}
		{
		var {{$nodes}} {{$.Expr.Type}}
	{{end -}}
	{{if eq $.Expr.Op '+' -}}
		{{gen $ $subExpr $.Fail -}}
		{{if $.ActionPass -}}
			{
				n, _ := node.({{$subExpr.Type}})
				{{$nodes}} = append({{$nodes}}, n)
			}
		{{end -}}
	{{end -}}
	for {
		{{if $.NodePass -}}
			{{$nkids}} := len(node.Kids)
		{{end -}}
		{{$pos0}} := pos
		{{gen $ $subExpr $fail -}}
		{{if $.ActionPass -}}
			{
				n, _ := node.({{$subExpr.Type}})
				{{$nodes}} = append({{$nodes}}, n)
			}
		{{end -}}
		continue
		{{$fail}}:
			{{if $.NodePass -}}
				node.Kids = node.Kids[:{{$nkids}}]
			{{end -}}
			pos = {{$pos0}}
			break
	}
	{{if $.ActionPass -}}
		node = {{$nodes}}
		}
	{{end -}}
`

var optExprTemplate = `// {{$.Expr.String}}
	{{$nkids := id "nkids" -}}
	{{$pos0 := id "pos" -}}
	{{- $fail := id "fail" -}}
	{{- $subExpr := $.Expr.Expr -}}
	{{- if $subExpr.CanFail -}}
	{
		{{if $.NodePass -}}
			{{$nkids}} := len(node.Kids)
		{{end -}}
		{{$pos0}} := pos
		{{gen $ $subExpr $fail -}}
		{{- $ok := id "ok" -}}
		{{if $.ActionPass -}}
			{
				n, _ := node.({{$subExpr.Type}})
				node = &n
			}
		{{end -}}
		goto {{$ok}}
		{{$fail}}:
			{{if $.NodePass -}}
				node.Kids = node.Kids[:{{$nkids}}]
			{{else if $.ActionPass -}}
				node = nil
			{{end -}}
			pos = {{$pos0}}
		{{$ok}}:
	}
	{{else -}}
		{{- /* TODO: disallow this case in check */ -}}
		{{gen $ $subExpr $fail -}}
	{{- end -}}
`

var subExprTemplate = `// {{$.Expr.String}}
	{{if $.NodePass -}}
	{
		{{- $pre := $.Config.Prefix -}}
		{{$nkids := id "nkids" -}}
		{{$nkids}} := len(node.Kids)
		{{$pos0 := id "pos0" -}}
		{{$pos0}} := pos
		{{gen $ $.Expr.Expr $.Fail -}}
		sub := {{$pre}}sub(parser, {{$pos0}}, pos, node.Kids[{{$nkids}}:])
		node.Kids = append(node.Kids[:{{$nkids}}], sub)
	}
	{{else -}}
		{{gen $ $.Expr.Expr $.Fail -}}
	{{end -}}
`

// TODO: instead, create a function for each predicate
// with params that are the parser followed by
// a string for each defined label.
// Predicate code shouldn't have access to the label.Kids,
// because it's undefined for the Accepts and Fail pass.
// NOTE: kids are OK for actions,
// because actions are only to be called by the Node pass
// on a successful parse.
var predCodeTemplate = `// pred code
	if ok := func(
		{{- if $.Expr.Labels -}}
			{{range $lexpr := $.Expr.Labels -}}
				{{$lexpr.Label}} string,
			{{- end -}}
		{{- end -}}) bool { return {{$.Expr.Code}} }(
		{{- if $.Expr.Labels -}}
			{{range $lexpr := $.Expr.Labels -}}
				labels[{{$lexpr.N}}],
			{{- end -}}
		{{- end -}}
	); {{if not $.Expr.Neg}}!{{end}}ok {
		{{if $.AcceptsPass -}}
			{{- $pre := $.Config.Prefix -}}
			perr = {{$pre}}max(perr, pos)
		{{else if $.FailPass -}}
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos: int(pos),
					Want:
					{{- if $.Expr.Neg}}"!{"{{else}}"&{"{{end}}+
					{{- quote $.Expr.Code.String}}+"}",
				})
			}
		{{end -}}
		{{if $.ActionPass -}}
			node = false
		{{end -}}
		goto {{$.Fail}}
	}
	{{if $.ActionPass -}}
		node = true
	{{end -}}
`

var identTemplate = `// {{$.Expr.String}}
	{{$pre := $.Config.Prefix -}}
	{{- $name := $.Expr.Name.String -}}
	{{if $.AcceptsPass -}}
		if !{{$pre}}accept(parser, {{$pre}}{{$name}}Accepts, &pos, &perr) {
			goto {{$.Fail}}
		}
	{{else if $.NodePass -}}
		if !{{$pre}}node(parser, {{$pre}}{{$name}}Node, node, &pos) {
			goto {{$.Fail}}
		}
	{{else if $.FailPass -}}
		if !{{$pre}}fail(parser, {{$pre}}{{$name}}Fail, errPos, failure, &pos) {
			goto {{$.Fail}}
		}
	{{else if $.ActionPass -}}
		if p, n := {{$pre}}{{$name}}Action(parser, pos); n == nil {
			goto {{$.Fail}}
		} else {
			node, pos = *n, p
		}
	{{end -}}
`

var literalTemplate = `// {{$.Expr.String}}
	{{$want := quote $.Expr.Text.String -}}
	{{- $n := len $.Expr.Text.String -}}
	if len(parser.text[pos:]) < {{$n}} || parser.text[pos:pos+{{$n}}] != {{$want}} {
		{{if $.AcceptsPass -}}
			{{- $pre := $.Config.Prefix -}}
			perr = {{$pre}}max(perr, pos)
		{{else if $.FailPass -}}
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos: int(pos),
					Want: {{quote $.Expr.String}},
				})
			}
		{{end -}}
		goto {{$.Fail}}
	}
	{{$pre := $.Config.Prefix -}}
	{{if $.NodePass -}}
		node.Kids = append(node.Kids, {{$pre}}leaf(parser, pos, pos + {{$n}}))
	{{else if $.ActionPass -}}
		node = parser.text[pos:pos+{{$n}}]
	{{end -}}
	{{if eq $n 1 -}}
		pos++
	{{- else -}}
		pos += {{$n}}
	{{- end}}
`

var anyTemplate = `// {{$.Expr.String}}
	{{$pre := $.Config.Prefix -}}
	{{- /* \uFFFD is utf8.RuneError */ -}}
	if r, w := {{$pre}}next(parser, pos); w == 0 || r == '\uFFFD' {
		{{if $.AcceptsPass -}}
			{{- $pre := $.Config.Prefix -}}
			perr = {{$pre}}max(perr, pos)
		{{else if $.FailPass -}}
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos: int(pos),
					Want: ".",
				})
			}
		{{end -}}
		goto {{$.Fail}}
	} else {
		{{$pre := $.Config.Prefix -}}
		{{if $.NodePass -}}
			node.Kids = append(node.Kids, {{$pre}}leaf(parser, pos, pos + w))
		{{else if $.ActionPass -}}
			node = parser.text[pos:pos+w]
		{{end -}}
		pos += w
	}
`

// charClassCondition emits the if-condition for a character class,
// assuming that r and w are the rune and its width respectively.
var charClassCondition = `
	{{- /* \uFFFD is utf8.RuneError */ -}}
	{{- if $.Expr.Neg -}}w == 0 || r == '\uFFFD' ||{{end}}
	{{- range $i, $span := $.Expr.Spans -}}
		{{- $first := index $span 0 -}}
		{{- $second := index $span 1 -}}
		{{- if $.Expr.Neg -}}
			{{- if gt $i 0 -}} || {{- end -}}
			{{- if eq $first $second -}}
				r == {{quoteRune $first}}
			{{- else -}}
				(r >= {{quoteRune $first}} && r <= {{quoteRune $second}})
			{{- end -}}
		{{- else -}}
			{{- if gt $i 0}} && {{end -}}
			{{- if eq $first $second -}}
				r != {{quoteRune $first}}
			{{- else -}}
				(r < {{quoteRune $first}} ||  r > {{quoteRune $second}})
			{{- end -}}
		{{- end -}}
	{{- end -}}
`

var charClassTemplate = `// {{$.Expr.String}}
	{{$pre := $.Config.Prefix -}}
	if r, w := {{$pre}}next(parser, pos);
		{{template "charClassCondition" $}} {
		{{if $.AcceptsPass -}}
			{{- $pre := $.Config.Prefix -}}
			perr = {{$pre}}max(perr, pos)
		{{else if $.FailPass -}}
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos: int(pos),
					Want: {{quote $.Expr.String}},
				})
			}
		{{end -}}
		goto {{$.Fail}}
	} else {
		{{$pre := $.Config.Prefix -}}
		{{if $.NodePass -}}
			{{$pre := $.Config.Prefix -}}
			node.Kids = append(node.Kids, {{$pre}}leaf(parser, pos, pos + w))
		{{else if $.ActionPass -}}
			node = parser.text[pos:pos+w]
		{{end -}}
		pos += w
	}
`
