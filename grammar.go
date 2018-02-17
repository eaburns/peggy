//line grammar.y:8
package main

import __yyfmt__ "fmt"

//line grammar.y:8
import "io"

//line grammar.y:13
type peggySymType struct {
	yys     int
	text    text
	cclass  *CharClass
	loc     Loc
	expr    Expr
	action  *Action
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

//line grammar.y:166

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

const peggyNprod = 41
const peggyPrivate = 57344

var peggyTokenNames []string
var peggyStates []string

const peggyLast = 80

var peggyAct = [...]int{

	24, 20, 19, 43, 45, 31, 44, 22, 32, 30,
	4, 59, 35, 2, 56, 26, 25, 29, 48, 49,
	50, 13, 55, 33, 35, 9, 42, 46, 58, 37,
	47, 34, 51, 41, 45, 31, 52, 53, 32, 30,
	18, 7, 57, 54, 10, 26, 25, 29, 23, 31,
	1, 6, 32, 30, 17, 16, 57, 15, 38, 26,
	25, 29, 39, 40, 14, 3, 32, 30, 10, 11,
	8, 12, 36, 26, 25, 29, 28, 27, 21, 5,
}
var peggyPact = [...]int{

	-11, -1000, 63, -1000, -11, -1000, -11, -11, -1000, -1000,
	49, -1000, 39, -1000, 39, 43, 15, -11, -1000, -3,
	-1000, 57, -1000, 19, -1000, -1, -1, 7, -1000, 43,
	-1000, -1000, -1000, 43, -1000, 43, -1000, -1000, 36, 8,
	0, 29, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 9, -3, -1000, -1000, 29, -1000, -1000, -1000, -1000,
}
var peggyPgo = [...]int{

	0, 79, 2, 1, 78, 7, 0, 77, 76, 72,
	3, 58, 51, 25, 41, 50, 13, 64,
}
var peggyR1 = [...]int{

	0, 15, 1, 1, 12, 14, 14, 14, 13, 13,
	2, 2, 3, 3, 4, 4, 5, 5, 6, 6,
	6, 7, 7, 7, 7, 8, 8, 8, 8, 8,
	8, 8, 8, 10, 9, 11, 11, 17, 17, 16,
	16,
}
var peggyR2 = [...]int{

	0, 2, 4, 2, 1, 3, 1, 0, 3, 4,
	3, 1, 2, 1, 2, 1, 3, 1, 2, 2,
	1, 2, 2, 2, 1, 3, 2, 2, 1, 1,
	1, 1, 3, 1, 2, 2, 2, 2, 1, 1,
	0,
}
var peggyChk = [...]int{

	-1000, -15, -16, -17, 21, -1, -12, -14, 7, -13,
	5, -17, -17, -16, -17, 8, 6, -14, -13, -2,
	-3, -4, -5, 5, -6, 17, 16, -7, -8, 18,
	10, 6, 9, 8, -16, 15, -9, -5, -11, 5,
	6, 14, -6, -10, 7, 5, -6, -10, 11, 12,
	13, -2, -2, -3, 7, 14, 14, -6, 19, 2,
}
var peggyDef = [...]int{

	40, -2, 7, 39, 38, 1, 0, 40, 4, 6,
	0, 37, 7, 3, 39, 0, 0, 40, 5, 8,
	11, 13, 15, 29, 17, 0, 0, 20, 24, 0,
	28, 30, 31, 0, 2, 0, 12, 14, 0, 29,
	30, 0, 18, 26, 33, 29, 19, 27, 21, 22,
	23, 0, 9, 10, 34, 35, 36, 16, 25, 32,
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
		//line grammar.y:39
		{
			peggylex.(*lexer).result = peggyDollar[2].grammar
		}
	case 2:
		peggyDollar = peggyS[peggypt-4 : peggypt+1]
		//line grammar.y:42
		{
			peggyVAL.grammar = Grammar{Prelude: peggyDollar[1].text, Rules: peggyDollar[3].rules}
		}
	case 3:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:43
		{
			peggyVAL.grammar = Grammar{Rules: peggyDollar[1].rules}
		}
	case 4:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:47
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
		//line grammar.y:58
		{
			peggyVAL.rules = append(peggyDollar[1].rules, peggyDollar[3].rule)
		}
	case 6:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:59
		{
			peggyVAL.rules = []Rule{peggyDollar[1].rule}
		}
	case 7:
		peggyDollar = peggyS[peggypt-0 : peggypt+1]
		//line grammar.y:63
		{
			peggyVAL.rules = nil
		}
	case 8:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:66
		{
			peggyVAL.rule = Rule{Name: peggyDollar[1].text, Expr: peggyDollar[3].expr}
		}
	case 9:
		peggyDollar = peggyS[peggypt-4 : peggypt+1]
		//line grammar.y:69
		{
			peggyVAL.rule = Rule{Name: peggyDollar[1].text, ErrorName: peggyDollar[2].text, Expr: peggyDollar[4].expr}
		}
	case 10:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:75
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
		//line grammar.y:83
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 12:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:87
		{
			peggyDollar[2].action.Expr = peggyDollar[1].expr
			peggyVAL.expr = peggyDollar[2].action
		}
	case 13:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:91
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 14:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:95
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
		//line grammar.y:103
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 16:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:106
		{
			peggyVAL.expr = &LabelExpr{Label: peggyDollar[1].text, Expr: peggyDollar[3].expr}
		}
	case 17:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:107
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 18:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:110
		{
			peggyVAL.expr = &PredExpr{Expr: peggyDollar[2].expr, Loc: peggyDollar[1].loc}
		}
	case 19:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:111
		{
			peggyVAL.expr = &PredExpr{Neg: true, Expr: peggyDollar[2].expr, Loc: peggyDollar[1].loc}
		}
	case 20:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:112
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 21:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:115
		{
			peggyVAL.expr = &RepExpr{Op: '*', Expr: peggyDollar[1].expr, Loc: peggyDollar[2].loc}
		}
	case 22:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:116
		{
			peggyVAL.expr = &RepExpr{Op: '+', Expr: peggyDollar[1].expr, Loc: peggyDollar[2].loc}
		}
	case 23:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:117
		{
			peggyVAL.expr = &OptExpr{Expr: peggyDollar[1].expr, Loc: peggyDollar[2].loc}
		}
	case 24:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:118
		{
			peggyVAL.expr = peggyDollar[1].expr
		}
	case 25:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:121
		{
			peggyVAL.expr = &SubExpr{Expr: peggyDollar[2].expr, Open: peggyDollar[1].loc, Close: peggyDollar[3].loc}
		}
	case 26:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:122
		{
			peggyVAL.expr = &PredCode{Code: peggyDollar[2].text, Loc: peggyDollar[1].loc}
		}
	case 27:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:123
		{
			peggyVAL.expr = &PredCode{Neg: true, Code: peggyDollar[2].text, Loc: peggyDollar[1].loc}
		}
	case 28:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:124
		{
			peggyVAL.expr = &Any{Loc: peggyDollar[1].loc}
		}
	case 29:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:125
		{
			peggyVAL.expr = &Ident{Name: peggyDollar[1].text}
		}
	case 30:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:126
		{
			peggyVAL.expr = &Literal{Text: peggyDollar[1].text}
		}
	case 31:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:127
		{
			peggyVAL.expr = peggyDollar[1].cclass
		}
	case 32:
		peggyDollar = peggyS[peggypt-3 : peggypt+1]
		//line grammar.y:128
		{
			peggylex.Error("unexpected end of file")
		}
	case 33:
		peggyDollar = peggyS[peggypt-1 : peggypt+1]
		//line grammar.y:132
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
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:144
		{
			loc := peggyDollar[2].text.Begin()
			loc.Col++ // skip the open {.
			err := ParseGoBody(loc, peggyDollar[2].text.String(), peggyDollar[1].text.String())
			if err != nil {
				peggylex.(*lexer).err = err
			}
			peggyVAL.action = &Action{Code: peggyDollar[2].text, ReturnType: peggyDollar[1].text}
		}
	case 35:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:155
		{
			peggyVAL.text = peggyDollar[1].text
		}
	case 36:
		peggyDollar = peggyS[peggypt-2 : peggypt+1]
		//line grammar.y:156
		{
			peggyVAL.text = peggyDollar[1].text
		}
	}
	goto peggystack /* stack new state and value */
}
