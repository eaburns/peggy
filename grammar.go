// Copyright 2017 The Peggy Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

//line grammar.y:5
package main

import __yyfmt__ "fmt"

//line grammar.y:5
import "io"

//line grammar.y:10
type peggySymType struct {
	yys     int
	text    text
	cclass  *CharClass
	loc     Loc
	expr    Expr
	rule    Rule
	rules   []Rule
	grammar Grammar
}

const _ERROR = 57346
const _IDENT = 57347
const _STRING = 57348
const _CODE = 57349
const _ARROW = 57350
const _CHARCLASS = 57351

var peggyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"_ERROR",
	"_IDENT",
	"_STRING",
	"_CODE",
	"_ARROW",
	"_CHARCLASS",
	"'.'",
	"'*'",
	"'+'",
	"'?'",
	"':'",
	"'/'",
	"'!'",
	"'&'",
	"'('",
	"')'",
	"'^'",
	"'\\n'",
}
var peggyStatenames = [...]string{}

const peggyEofCode = 1
const peggyErrCode = 2
const peggyInitialStackSize = 16

//line grammar.y:153

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

//line yacctab:1
var peggyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const peggyPrivate = 57344

const peggyLast = 83

var peggyAct = [...]int{

	24, 20, 19, 41, 43, 31, 42, 22, 32, 30,
	4, 54, 35, 2, 39, 26, 25, 29, 46, 47,
	48, 13, 7, 9, 35, 33, 40, 44, 53, 37,
	45, 34, 49, 10, 1, 17, 50, 51, 18, 6,
	52, 23, 31, 38, 36, 32, 30, 16, 10, 15,
	8, 28, 26, 25, 29, 43, 31, 14, 3, 32,
	30, 27, 11, 21, 12, 5, 26, 25, 29, 23,
	31, 0, 0, 32, 30, 0, 0, 0, 0, 0,
	26, 25, 29,
}
var peggyPact = [...]int{

	-11, -1000, 43, -1000, -11, -1000, -11, -11, -1000, -1000,
	41, -1000, 28, -1000, 28, 64, 17, -11, -1000, -3,
	-1000, 36, -1000, 0, -1000, -1, -1, 7, -1000, 64,
	-1000, -1000, -1000, 64, -1000, 64, -1000, -1000, -1000, 50,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 9,
	-3, -1000, -1000, -1000, -1000,
}
var peggyPgo = [...]int{

	0, 65, 2, 1, 63, 7, 0, 61, 51, 3,
	44, 39, 23, 22, 34, 13, 57,
}
var peggyR1 = [...]int{

	0, 14, 1, 1, 11, 13, 13, 13, 12, 12,
	2, 2, 3, 3, 4, 4, 5, 5, 6, 6,
	6, 7, 7, 7, 7, 8, 8, 8, 8, 8,
	8, 8, 8, 9, 10, 16, 16, 15, 15,
}
var peggyR2 = [...]int{

	0, 2, 4, 2, 1, 3, 1, 0, 3, 4,
	3, 1, 2, 1, 2, 1, 3, 1, 2, 2,
	1, 2, 2, 2, 1, 3, 2, 2, 1, 1,
	1, 1, 3, 1, 1, 2, 1, 1, 0,
}
var peggyChk = [...]int{

	-1000, -14, -15, -16, 21, -1, -11, -13, 7, -12,
	5, -16, -16, -15, -16, 8, 6, -13, -12, -2,
	-3, -4, -5, 5, -6, 17, 16, -7, -8, 18,
	10, 6, 9, 8, -15, 15, -10, -5, 7, 14,
	-6, -9, 7, 5, -6, -9, 11, 12, 13, -2,
	-2, -3, -6, 19, 2,
}
var peggyDef = [...]int{

	38, -2, 7, 37, 36, 1, 0, 38, 4, 6,
	0, 35, 7, 3, 37, 0, 0, 38, 5, 8,
	11, 13, 15, 29, 17, 0, 0, 20, 24, 0,
	28, 30, 31, 0, 2, 0, 12, 14, 34, 0,
	18, 26, 33, 29, 19, 27, 21, 22, 23, 0,
	9, 10, 16, 25, 32,
}
var peggyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	21, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 16, 3, 3, 3, 3, 17, 3,
	18, 19, 11, 12, 3, 3, 10, 15, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 14, 3,
	3, 3, 3, 13, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 20,
}
var peggyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9,
}
var peggyTok3 = [...]int{
	0,
}

var peggyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	peggyDebug        = 0
	peggyErrorVerbose = false
)

type peggyLexer interface {
	Lex(lval *peggySymType) int
	Error(s string)
}

type peggyParser interface {
	Parse(peggyLexer) int
	Lookahead() int
}

type peggyParserImpl struct {
	lval  peggySymType
	stack [peggyInitialStackSize]peggySymType
	char  int
}

func (p *peggyParserImpl) Lookahead() int {
	return p.char
}

func peggyNewParser() peggyParser {
	return &peggyParserImpl{}
}

const peggyFlag = -1000

func peggyTokname(c int) string {
	if c >= 1 && c-1 < len(peggyToknames) {
		if peggyToknames[c-1] != "" {
			return peggyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func peggyStatname(s int) string {
	if s >= 0 && s < len(peggyStatenames) {
		if peggyStatenames[s] != "" {
			return peggyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func peggyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !peggyErrorVerbose {
		return "syntax error"
	}

	for _, e := range peggyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + peggyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := peggyPact[state]
	for tok := TOKSTART; tok-1 < len(peggyToknames); tok++ {
		if n := base + tok; n >= 0 && n < peggyLast && peggyChk[peggyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if peggyDef[state] == -2 {
		i := 0
		for peggyExca[i] != -1 || peggyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; peggyExca[i] >= 0; i += 2 {
			tok := peggyExca[i]
			if tok < TOKSTART || peggyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if peggyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += peggyTokname(tok)
	}
	return res
}

func peggylex1(lex peggyLexer, lval *peggySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = peggyTok1[0]
		goto out
	}
	if char < len(peggyTok1) {
		token = peggyTok1[char]
		goto out
	}
	if char >= peggyPrivate {
		if char < peggyPrivate+len(peggyTok2) {
			token = peggyTok2[char-peggyPrivate]
			goto out
		}
	}
	for i := 0; i < len(peggyTok3); i += 2 {
		token = peggyTok3[i+0]
		if token == char {
			token = peggyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = peggyTok2[1] /* unknown char */
	}
	if peggyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", peggyTokname(token), uint(char))
	}
	return char, token
}

func peggyParse(peggylex peggyLexer) int {
	return peggyNewParser().Parse(peggylex)
}

func (peggyrcvr *peggyParserImpl) Parse(peggylex peggyLexer) int {
	var peggyn int
	var peggyVAL peggySymType
	var peggyDollar []peggySymType
	_ = peggyDollar // silence set and not used
	peggyS := peggyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	peggystate := 0
	peggyrcvr.char = -1
	peggytoken := -1 // peggyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		peggystate = -1
		peggyrcvr.char = -1
		peggytoken = -1
	}()
	peggyp := -1
	goto peggystack

ret0:
	return 0

ret1:
	return 1

peggystack:
	/* put a state and value onto the stack */
	if peggyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", peggyTokname(peggytoken), peggyStatname(peggystate))
	}

	peggyp++
	if peggyp >= len(peggyS) {
		nyys := make([]peggySymType, len(peggyS)*2)
		copy(nyys, peggyS)
		peggyS = nyys
	}
	peggyS[peggyp] = peggyVAL
	peggyS[peggyp].yys = peggystate

peggynewstate:
	peggyn = peggyPact[peggystate]
	if peggyn <= peggyFlag {
		goto peggydefault /* simple state */
	}
	if peggyrcvr.char < 0 {
		peggyrcvr.char, peggytoken = peggylex1(peggylex, &peggyrcvr.lval)
	}
	peggyn += peggytoken
	if peggyn < 0 || peggyn >= peggyLast {
		goto peggydefault
	}
	peggyn = peggyAct[peggyn]
	if peggyChk[peggyn] == peggytoken { /* valid shift */
		peggyrcvr.char = -1
		peggytoken = -1
		peggyVAL = peggyrcvr.lval
		peggystate = peggyn
		if Errflag > 0 {
			Errflag--
		}
		goto peggystack
	}

peggydefault:
	/* default state action */
	peggyn = peggyDef[peggystate]
	if peggyn == -2 {
		if peggyrcvr.char < 0 {
			peggyrcvr.char, peggytoken = peggylex1(peggylex, &peggyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if peggyExca[xi+0] == -1 && peggyExca[xi+1] == peggystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			peggyn = peggyExca[xi+0]
			if peggyn < 0 || peggyn == peggytoken {
				break
			}
		}
		peggyn = peggyExca[xi+1]
		if peggyn < 0 {
			goto ret0
		}
	}
	if peggyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			peggylex.Error(peggyErrorMessage(peggystate, peggytoken))
			Nerrs++
			if peggyDebug >= 1 {
				__yyfmt__.Printf("%s", peggyStatname(peggystate))
				__yyfmt__.Printf(" saw %s\n", peggyTokname(peggytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for peggyp >= 0 {
				peggyn = peggyPact[peggyS[peggyp].yys] + peggyErrCode
				if peggyn >= 0 && peggyn < peggyLast {
					peggystate = peggyAct[peggyn] /* simulate a shift of "error" */
					if peggyChk[peggystate] == peggyErrCode {
						goto peggystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if peggyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", peggyS[peggyp].yys)
				}
				peggyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if peggyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", peggyTokname(peggytoken))
			}
			if peggytoken == peggyEofCode {
				goto ret1
			}
			peggyrcvr.char = -1
			peggytoken = -1
			goto peggynewstate /* try again in the same state */
		}
	}

	/* reduction by production peggyn */
	if peggyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", peggyn, peggyStatname(peggystate))
	}

	peggynt := peggyn
	peggypt := peggyp
	_ = peggypt // guard against "declared and not used"

	peggyp -= peggyR2[peggyn]
	// peggyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if peggyp+1 >= len(peggyS) {
		nyys := make([]peggySymType, len(peggyS)*2)
		copy(nyys, peggyS)
		peggyS = nyys
	}
	peggyVAL = peggyS[peggyp+1]

	/* consult goto table to find next state */
	peggyn = peggyR1[peggyn]
	peggyg := peggyPgo[peggyn]
	peggyj := peggyg + peggyS[peggyp].yys + 1

	if peggyj >= peggyLast {
		peggystate = peggyAct[peggyg]
	} else {
		peggystate = peggyAct[peggyj]
		if peggyChk[peggystate] != -peggyn {
			peggystate = peggyAct[peggyg]
		}
	}
	// dummy call; replaced with literal code
	switch peggynt {

	case 1:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:34
		{
			peggylex.(*lexer).result = peggyDollar[2].grammar
		}
	case 2:
		peggyDollar = peggyS[peggypt-4 : peggypt+1]
		//line grammar.y:37
		{
			peggyVAL.grammar = Grammar{Prelude: peggyDollar[1].text, Rules: peggyDollar[3].rules}
		}
	case 3:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:38
		{
			peggyVAL.grammar = Grammar{Rules: peggyDollar[1].rules}
		}
	case 4:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:42
		{
			loc := peggyDollar[1].text.Begin()
			loc.Col++ // skip the open {.
			err := ParseGoFile(loc, peggyDollar[1].text.String())
			if err != nil {
				peggylex.(*lexer).err = err
			}
			peggyVAL.text = peggyDollar[1].text
		}
	case 5:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:53
		{
			peggyVAL.rules = append(peggyDollar[1].rules, peggyDollar[3].rule)
		}
	case 6:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:54
		{
			peggyVAL.rules = []Rule{peggyDollar[1].rule}
		}
	case 7:
		peggyDollar = peggyS[peggypt-0 : peggypt+1]
		//line grammar.y:58
		{
			peggyVAL.rules = nil
		}
	case 8:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:61
		{
			peggyVAL.rule = Rule{Name: peggyDollar[1].text, Expr: peggyDollar[3].expr}
		}
	case 9:
		peggyDollar = peggyS[peggypt-4 : peggypt+1]
		//line grammar.y:64
		{
			peggyVAL.rule = Rule{Name: peggyDollar[1].text, ErrorName: peggyDollar[2].text, Expr: peggyDollar[4].expr}
		}
	case 10:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:70
		{
			e, ok := peggyDollar[1].expr.(*Choice)
			if !ok {
				e = &Choice{Exprs: []Expr{peggyDollar[1].expr}}
			}
			e.Exprs = append(e.Exprs, peggyDollar[3].expr)
			peggyVAL.expr = e
		}
	case 11:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:78
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 12:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:81
		{
			peggyVAL.expr = &Action{Expr: peggyDollar[1].expr, Code: peggyDollar[2].text}
		}
	case 13:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:82
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 14:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:86
		{
			e, ok := peggyDollar[1].expr.(*Sequence)
			if !ok {
				e = &Sequence{Exprs: []Expr{peggyDollar[1].expr}}
			}
			e.Exprs = append(e.Exprs, peggyDollar[2].expr)
			peggyVAL.expr = e
		}
	case 15:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:94
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 16:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:97
		{
			peggyVAL.expr = &LabelExpr{Label: peggyDollar[1].text, Expr: peggyDollar[3].expr}
		}
	case 17:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:98
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 18:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:101
		{
			peggyVAL.expr = &PredExpr{Expr: peggyDollar[2].expr, Loc: peggyDollar[1].loc}
		}
	case 19:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:102
		{
			peggyVAL.expr = &PredExpr{Neg: true, Expr: peggyDollar[2].expr, Loc: peggyDollar[1].loc}
		}
	case 20:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:103
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 21:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:106
		{
			peggyVAL.expr = &RepExpr{Op: '*', Expr: peggyDollar[1].expr, Loc: peggyDollar[2].loc}
		}
	case 22:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:107
		{
			peggyVAL.expr = &RepExpr{Op: '+', Expr: peggyDollar[1].expr, Loc: peggyDollar[2].loc}
		}
	case 23:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:108
		{
			peggyVAL.expr = &OptExpr{Expr: peggyDollar[1].expr, Loc: peggyDollar[2].loc}
		}
	case 24:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:109
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 25:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:112
		{
			peggyVAL.expr = &SubExpr{Expr: peggyDollar[2].expr, Open: peggyDollar[1].loc, Close: peggyDollar[3].loc}
		}
	case 26:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:113
		{
			peggyVAL.expr = &PredCode{Code: peggyDollar[2].text, Loc: peggyDollar[1].loc}
		}
	case 27:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:114
		{
			peggyVAL.expr = &PredCode{Neg: true, Code: peggyDollar[2].text, Loc: peggyDollar[1].loc}
		}
	case 28:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:115
		{
			peggyVAL.expr = &Any{Loc: peggyDollar[1].loc}
		}
	case 29:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:116
		{
			peggyVAL.expr = &Ident{Name: peggyDollar[1].text}
		}
	case 30:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:117
		{
			peggyVAL.expr = &Literal{Text: peggyDollar[1].text}
		}
	case 31:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:118
		{
			peggyVAL.expr = peggyDollar[1].cclass
		}
	case 32:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:119
		{
			peggylex.Error("unexpected end of file")
		}
	case 33:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:123
		{
			loc := peggyDollar[1].text.Begin()
			loc.Col++ // skip the open {.
			err := ParseGoExpr(loc, peggyDollar[1].text.String())
			if err != nil {
				peggylex.(*lexer).err = err
			}
			peggyVAL.text = peggyDollar[1].text
		}
	case 34:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:135
		{
			loc := peggyDollar[1].text.Begin()
			loc.Col++ // skip the open {.
			err := ParseGoBody(loc, peggyDollar[1].text.String())
			if err != nil {
				peggylex.(*lexer).err = err
			}
			peggyVAL.text = peggyDollar[1].text
		}
	}
	goto peggystack /* stack new state and value */
}
