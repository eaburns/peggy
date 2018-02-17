// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

%{
package main

import "io"
%}

%union{
	text text
	cclass *CharClass
	loc Loc
	expr Expr
	action *Action
	rule Rule
	rules []Rule
	grammar Grammar
}

%type <grammar> Grammar
%type <expr> Expr, ActExpr, SeqExpr, LabelExpr, PredExpr, RepExpr, Operand
%type <action> GoAction
%type <text> GoPred GoType Prelude
%type <rule> Rule
%type <rules> Rules

%token _ERROR
%token <text> _IDENT _STRING _CODE _ARROW
%token <cclass> _CHARCLASS
%token <loc> '.', '*', '+', '?', ':', '/', '!', '&', '(', ')', '^'

%%

Top:
	OptionalNewLine Grammar { peggylex.(*lexer).result = $2 }

Grammar:
	Prelude NewLine Rules OptionalNewLine { $$ = Grammar{ Prelude: $1, Rules: $3 } }
|	Rules OptionalNewLine { $$ = Grammar{ Rules: $1 } }

Prelude:
	_CODE
	{
		loc := $1.Begin()
		loc.Col++ // skip the open {.
		err := ParseGoFile(loc, $1.String())
		if err != nil {
			peggylex.(*lexer).err = err
		}
		$$ = $1
	}

Rules:
	Rules NewLine Rule { $$ = append($1, $3) }
|	Rule { $$ = []Rule{ $1 } }
// The following production adds a shift/reduce conflict:
// 	reduce the empty string or shift into a Rule?
// Yacc always prefers shift in the case of both, which is the desired behavior.
|	{ $$ = nil }

Rule:
	_IDENT _ARROW Expr {
		$$ = Rule{ Name: $1, Expr: $3 }
	}
|	_IDENT _STRING _ARROW Expr {
		$$ = Rule{ Name: $1, ErrorName: $2, Expr: $4 }
	}

Expr:
	Expr '/' ActExpr
	{
		e, ok := $1.(*Choice)
		if !ok {
			e = &Choice{ Exprs: []Expr{$1} }
		}
		e.Exprs = append(e.Exprs, $3)
		$$ = e
	}
|	ActExpr { $$ = $1 }

ActExpr:
	SeqExpr GoAction
	{
		$2.Expr = $1
		$$ = $2
	}
|	SeqExpr { $$ = $1 }

SeqExpr:
	SeqExpr LabelExpr
	{
		e, ok := $1.(*Sequence)
		if !ok {
			e = &Sequence{ Exprs: []Expr{$1} }
		}
		e.Exprs = append(e.Exprs, $2)
		$$ = e
	}
|	LabelExpr { $$ = $1 }

LabelExpr:
	_IDENT ':' PredExpr { $$ = &LabelExpr{ Label: $1, Expr: $3 } }
|	PredExpr { $$ = $1 }

PredExpr:
	'&' PredExpr { $$ = &PredExpr{ Expr: $2, Loc: $1 } }
|	'!' PredExpr { $$ = &PredExpr{ Neg: true, Expr: $2, Loc: $1 } }
|	RepExpr { $$ = $1 }

RepExpr:
	RepExpr '*' { $$ = &RepExpr{ Op: '*', Expr: $1, Loc: $2 } }
|	RepExpr '+' { $$ = &RepExpr{ Op: '+', Expr: $1, Loc: $2 } }
|	RepExpr '?' { $$ = &OptExpr{ Expr: $1, Loc: $2 } }
|	Operand { $$ = $1 }

Operand:
	'(' Expr ')' { $$ = &SubExpr{ Expr: $2, Open: $1, Close: $3 } }
|	'&' GoPred { $$ = &PredCode{ Code: $2, Loc: $1 } }
|	'!' GoPred { $$ = &PredCode{ Neg: true, Code: $2, Loc: $1 } }
|	'.' { $$ = &Any{ Loc: $1 } }
|	_IDENT { $$ = &Ident{ Name: $1 } }
|	_STRING { $$ = &Literal{ Text: $1 } }
|	_CHARCLASS { $$ =$1 }
|	'(' Expr error { peggylex.Error("unexpected end of file") }

GoPred:
	_CODE
	{
		loc := $1.Begin()
		loc.Col++ // skip the open {.
		err := ParseGoExpr(loc, $1.String())
		if err != nil {
			peggylex.(*lexer).err = err
		}
		$$ = $1
	}

GoAction:
	GoType _CODE
	{
		loc := $2.Begin()
		loc.Col++ // skip the open {.
		err := ParseGoBody(loc, $2.String(), $1.String() )
		if err != nil {
			peggylex.(*lexer).err = err
		}
		$$ = &Action{ Code: $2, ReturnType: $1 }
	}

GoType:
	_IDENT ':' { $$ = $1 }
|	_STRING ':' { $$ = $1 }

NewLine:
	'\n' NewLine
|	'\n'

OptionalNewLine:
	NewLine
|

%%

// Parse parses a Peggy input file, and returns the Grammar.
func Parse(in io.RuneScanner, fileName string) (*Grammar, error) {
	x := &lexer{
		in:   in,
		file: fileName,
		line: 1,
	}
	peggyParse(x)
	if x.err != nil {
		return nil, x.err
	}
	return &x.result, nil
}
