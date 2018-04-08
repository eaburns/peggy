// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"strings"
)

// ParseGoFile parses go function body statements, returning any syntax errors.
// The errors contain location information starting from the given Loc.
func ParseGoFile(loc Loc, code string) error {
	_, err := parser.ParseFile(token.NewFileSet(), loc.File, code, 0)
	if err == nil {
		return nil
	}

	el, ok := err.(scanner.ErrorList)
	if !ok {
		return err
	}
	p := el[0].Pos
	loc.Line += p.Line - 1 // -1 because p.Line is 1-based.
	if p.Line > 1 {
		loc.Col = 1
	}
	loc.Col += p.Column - 1
	return Err(loc, el[0].Msg)
}

// ParseGoBody parses go function body statements, returning any syntax errors.
// The errors contain location information starting from the given Loc.
func ParseGoBody(loc Loc, code string) (string, error) {
	code = "package main; func p() interface{} {\n" + code + "}"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, loc.File, code, 0)
	if err == nil {
		return inferType(loc, fset, file)
	}

	el, ok := err.(scanner.ErrorList)
	if !ok {
		return "", err
	}
	p := el[0].Pos
	loc.Line += p.Line - 2 // -2 because p.Line is 1-based and the func line.
	if p.Line > 2 {
		loc.Col = 1
	}
	loc.Col += p.Column - 1
	return "", Err(loc, el[0].Msg)
}

// inferType infers the type of a function by considering its first return statement.
// If the returned expression is:
// 	* a type conversion, the type is returned.
// 	* a type assertion, the type is returned.
// 	* a function literal, the type is returned.
// 	* a composite literal, the type is returned.
// 	* an &-composite literal, the type is returned.
// 	* an int literal, int is returned.
// 	* a float literal, float64 is returned.
// 	* a character literal, rune is returned.
// 	* a string literal, string is returned.
//
// If the file does not have exactly one top-level funciton, inferType panics.
// If the function has no return statement, an error is returned.
// If the return statement does not have exactly one returned value, an error is returned.
// If the returned value is not an expression in the list above, an error is returned.
func inferType(loc Loc, fset *token.FileSet, file *ast.File) (string, error) {
	var funcDecl *ast.FuncDecl
	for _, decl := range file.Decls {
		if d, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl != nil {
				panic("multiple function declarations")
			}
			funcDecl = d
		}
	}
	if funcDecl == nil {
		panic("no function declarations")
	}

	var v findReturnVisitor
	ast.Walk(&v, funcDecl)
	if v.retStmt == nil {
		return "", Err(loc, "no return statement")
	}
	if len(v.retStmt.Results) != 1 {
		return "", Err(loc, "must return exactly one value")
	}

	var typ interface{}
	switch e := v.retStmt.Results[0].(type) {
	case *ast.CallExpr:
		if len(e.Args) != 1 {
			var s strings.Builder
			printer.Fprint(&s, fset, e)
			return "", Err(loc, "cannot infer type from a function call: "+s.String())
		}
		typ = e.Fun
	case *ast.TypeAssertExpr:
		typ = e.Type
	case *ast.FuncLit:
		typ = e.Type
	case *ast.CompositeLit:
		typ = e.Type
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			return "int", nil
		case token.FLOAT:
			return "float64", nil
		case token.CHAR:
			return "rune", nil
		case token.STRING:
			return "string", nil
		}
	case *ast.UnaryExpr:
		lit, ok := e.X.(*ast.CompositeLit)
		if !ok || e.Op != token.AND {
			return "", Err(loc, "cannot infer type")
		}
		var s strings.Builder
		printer.Fprint(&s, fset, lit.Type)
		return "*" + s.String(), nil
	default:
		return "", Err(loc, "cannot infer type")
	}
	var s strings.Builder
	printer.Fprint(&s, fset, typ)
	return s.String(), nil
}

type findReturnVisitor struct {
	retStmt *ast.ReturnStmt
}

func (v *findReturnVisitor) Visit(n ast.Node) ast.Visitor {
	if r, ok := n.(*ast.ReturnStmt); ok {
		v.retStmt = r
		return nil
	}
	return v
}

// ParseGoExpr parses a go expression, returning any syntax errors.
// The errors contain location information starting from the given Loc.
func ParseGoExpr(loc Loc, code string) error {
	_, err := parser.ParseExprFrom(token.NewFileSet(), loc.File, code, 0)
	if err == nil {
		return nil
	}

	el, ok := err.(scanner.ErrorList)
	if !ok {
		return err
	}
	p := el[0].Pos
	loc.Line += p.Line - 1 // -1 because p.Line is 1-based.
	if p.Line > 1 {
		loc.Col = 1
	}
	loc.Col += p.Column - 1
	return Err(loc, el[0].Msg)
}
