// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"go/parser"
	"go/scanner"
	"go/token"
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
func ParseGoBody(loc Loc, code string) error {
	code = "package main; func p() {\n" + code + "}"
	_, err := parser.ParseFile(token.NewFileSet(), loc.File, code, 0)
	if err == nil {
		return nil
	}

	el, ok := err.(scanner.ErrorList)
	if !ok {
		return err
	}
	p := el[0].Pos
	loc.Line += p.Line - 2 // -2 because p.Line is 1-based and the func line.
	if p.Line > 2 {
		loc.Col = 1
	}
	loc.Col += p.Column - 1
	return Err(loc, el[0].Msg)
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
