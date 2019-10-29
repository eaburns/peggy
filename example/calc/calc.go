// Calc is an example calculator program.
// You can build it from calc.peggy with
// 	peggy -o calc.go calc.peggy
package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/eaburns/peggy/peg"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		p := _NewParser(line)
		if pos, perr := _ExprAccepts(p, 0); pos < 0 {
			_, fail := _ExprFail(p, 0, perr)
			fmt.Println(peg.SimpleError(line, fail))
			continue
		}
		_, result := _ExprAction(p, 0)
		fmt.Println((*result).String())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type op func(*big.Float, *big.Float, *big.Float) *big.Float

type tail struct {
	op op
	r  *big.Float
}

func evalTail(l big.Float, tail []tail) big.Float {
	for _, t := range tail {
		t.op(&l, &l, t.r)
	}
	return l
}

func isSpace(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsSpace(r)
}

const (
	_Expr        int = 0
	_Sum         int = 1
	_SumTail     int = 2
	_AddOp       int = 3
	_Product     int = 4
	_ProductTail int = 5
	_MulOp       int = 6
	_Value       int = 7
	_Num         int = 8
	__           int = 9
	_EOF         int = 10

	_N int = 11
)

type _Parser struct {
	text     string
	deltaPos [][_N]int32
	deltaErr [][_N]int32
	node     map[_key]*peg.Node
	fail     map[_key]*peg.Fail
	act      map[_key]interface{}
	lastFail int
	data     interface{}
}

type _key struct {
	start int
	rule  int
}

func _NewParser(text string) *_Parser {
	return &_Parser{
		text:     text,
		deltaPos: make([][_N]int32, len(text)+1),
		deltaErr: make([][_N]int32, len(text)+1),
		node:     make(map[_key]*peg.Node),
		fail:     make(map[_key]*peg.Fail),
		act:      make(map[_key]interface{}),
	}
}

func _max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func _memoize(parser *_Parser, rule, start, pos, perr int) (int, int) {
	parser.lastFail = perr
	derr := perr - start
	parser.deltaErr[start][rule] = int32(derr + 1)
	if pos >= 0 {
		dpos := pos - start
		parser.deltaPos[start][rule] = int32(dpos + 1)
		return dpos, derr
	}
	parser.deltaPos[start][rule] = -1
	return -1, derr
}

func _memo(parser *_Parser, rule, start int) (int, int, bool) {
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

func _failMemo(parser *_Parser, rule, start, errPos int) (int, *peg.Fail) {
	if start > parser.lastFail {
		return -1, &peg.Fail{}
	}
	dp := parser.deltaPos[start][rule]
	de := parser.deltaErr[start][rule]
	if start+int(de-1) < errPos {
		if dp > 0 {
			return start + int(dp-1), &peg.Fail{}
		}
		return -1, &peg.Fail{}
	}
	f := parser.fail[_key{start: start, rule: rule}]
	if dp < 0 && f != nil {
		return -1, f
	}
	if dp > 0 && f != nil {
		return start + int(dp-1), f
	}
	return start, nil
}

func _accept(parser *_Parser, f func(*_Parser, int) (int, int), pos, perr *int) bool {
	dp, de := f(parser, *pos)
	*perr = _max(*perr, *pos+de)
	if dp < 0 {
		return false
	}
	*pos += dp
	return true
}

func _node(parser *_Parser, f func(*_Parser, int) (int, *peg.Node), node *peg.Node, pos *int) bool {
	p, kid := f(parser, *pos)
	if kid == nil {
		return false
	}
	node.Kids = append(node.Kids, kid)
	*pos = p
	return true
}

func _fail(parser *_Parser, f func(*_Parser, int, int) (int, *peg.Fail), errPos int, node *peg.Fail, pos *int) bool {
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

func _next(parser *_Parser, pos int) (rune, int) {
	r, w := peg.DecodeRuneInString(parser.text[pos:])
	return r, w
}

func _sub(parser *_Parser, start, end int, kids []*peg.Node) *peg.Node {
	node := &peg.Node{
		Text: parser.text[start:end],
		Kids: make([]*peg.Node, len(kids)),
	}
	copy(node.Kids, kids)
	return node
}

func _leaf(parser *_Parser, start, end int) *peg.Node {
	return &peg.Node{Text: parser.text[start:end]}
}

// A no-op function to mark a variable as used.
func use(interface{}) {}

func _ExprAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [1]string
	use(labels)
	if dp, de, ok := _memo(parser, _Expr, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// action
	// s:Sum EOF
	// s:Sum
	{
		pos1 := pos
		// Sum
		if !_accept(parser, _SumAccepts, &pos, &perr) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// EOF
	if !_accept(parser, _EOFAccepts, &pos, &perr) {
		goto fail
	}
	return _memoize(parser, _Expr, start, pos, perr)
fail:
	return _memoize(parser, _Expr, start, -1, perr)
}

func _ExprNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [1]string
	use(labels)
	dp := parser.deltaPos[start][_Expr]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Expr}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "Expr"}
	// action
	// s:Sum EOF
	// s:Sum
	{
		pos1 := pos
		// Sum
		if !_node(parser, _SumNode, node, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// EOF
	if !_node(parser, _EOFNode, node, &pos) {
		goto fail
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _ExprFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [1]string
	use(labels)
	pos, failure := _failMemo(parser, _Expr, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "Expr",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _Expr}
	// action
	// s:Sum EOF
	// s:Sum
	{
		pos1 := pos
		// Sum
		if !_fail(parser, _SumFail, errPos, failure, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// EOF
	if !_fail(parser, _EOFFail, errPos, failure, &pos) {
		goto fail
	}
	parser.fail[key] = failure
	return pos, failure
fail:
	parser.fail[key] = failure
	return -1, failure
}

func _ExprAction(parser *_Parser, start int) (int, *(*big.Float)) {
	var labels [1]string
	use(labels)
	var label0 (big.Float)
	dp := parser.deltaPos[start][_Expr]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Expr}
	n := parser.act[key]
	if n != nil {
		n := n.((*big.Float))
		return start + int(dp-1), &n
	}
	var node (*big.Float)
	pos := start
	// action
	{
		start0 := pos
		// s:Sum EOF
		// s:Sum
		{
			pos2 := pos
			// Sum
			if p, n := _SumAction(parser, pos); n == nil {
				goto fail
			} else {
				label0 = *n
				pos = p
			}
			labels[0] = parser.text[pos2:pos]
		}
		// EOF
		if p, n := _EOFAction(parser, pos); n == nil {
			goto fail
		} else {
			pos = p
		}
		node = func(
			start, end int, s big.Float) *big.Float {
			return (*big.Float)(&s)
		}(
			start0, pos, label0)
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _SumAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [2]string
	use(labels)
	if dp, de, ok := _memo(parser, _Sum, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// action
	// l:Product tail:SumTail*
	// l:Product
	{
		pos1 := pos
		// Product
		if !_accept(parser, _ProductAccepts, &pos, &perr) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// tail:SumTail*
	{
		pos2 := pos
		// SumTail*
		for {
			pos4 := pos
			// SumTail
			if !_accept(parser, _SumTailAccepts, &pos, &perr) {
				goto fail6
			}
			continue
		fail6:
			pos = pos4
			break
		}
		labels[1] = parser.text[pos2:pos]
	}
	return _memoize(parser, _Sum, start, pos, perr)
fail:
	return _memoize(parser, _Sum, start, -1, perr)
}

func _SumNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [2]string
	use(labels)
	dp := parser.deltaPos[start][_Sum]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Sum}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "Sum"}
	// action
	// l:Product tail:SumTail*
	// l:Product
	{
		pos1 := pos
		// Product
		if !_node(parser, _ProductNode, node, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// tail:SumTail*
	{
		pos2 := pos
		// SumTail*
		for {
			nkids3 := len(node.Kids)
			pos4 := pos
			// SumTail
			if !_node(parser, _SumTailNode, node, &pos) {
				goto fail6
			}
			continue
		fail6:
			node.Kids = node.Kids[:nkids3]
			pos = pos4
			break
		}
		labels[1] = parser.text[pos2:pos]
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _SumFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [2]string
	use(labels)
	pos, failure := _failMemo(parser, _Sum, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "Sum",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _Sum}
	// action
	// l:Product tail:SumTail*
	// l:Product
	{
		pos1 := pos
		// Product
		if !_fail(parser, _ProductFail, errPos, failure, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// tail:SumTail*
	{
		pos2 := pos
		// SumTail*
		for {
			pos4 := pos
			// SumTail
			if !_fail(parser, _SumTailFail, errPos, failure, &pos) {
				goto fail6
			}
			continue
		fail6:
			pos = pos4
			break
		}
		labels[1] = parser.text[pos2:pos]
	}
	parser.fail[key] = failure
	return pos, failure
fail:
	parser.fail[key] = failure
	return -1, failure
}

func _SumAction(parser *_Parser, start int) (int, *(big.Float)) {
	var labels [2]string
	use(labels)
	var label0 (big.Float)
	var label1 []tail
	dp := parser.deltaPos[start][_Sum]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Sum}
	n := parser.act[key]
	if n != nil {
		n := n.((big.Float))
		return start + int(dp-1), &n
	}
	var node (big.Float)
	pos := start
	// action
	{
		start0 := pos
		// l:Product tail:SumTail*
		// l:Product
		{
			pos2 := pos
			// Product
			if p, n := _ProductAction(parser, pos); n == nil {
				goto fail
			} else {
				label0 = *n
				pos = p
			}
			labels[0] = parser.text[pos2:pos]
		}
		// tail:SumTail*
		{
			pos3 := pos
			// SumTail*
			for {
				pos5 := pos
				var node6 tail
				// SumTail
				if p, n := _SumTailAction(parser, pos); n == nil {
					goto fail7
				} else {
					node6 = *n
					pos = p
				}
				label1 = append(label1, node6)
				continue
			fail7:
				pos = pos5
				break
			}
			labels[1] = parser.text[pos3:pos]
		}
		node = func(
			start, end int, l big.Float, tail []tail) big.Float {
			return (big.Float)(evalTail(l, tail))
		}(
			start0, pos, label0, label1)
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _SumTailAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [2]string
	use(labels)
	if dp, de, ok := _memo(parser, _SumTail, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// action
	// op:AddOp r:Product
	// op:AddOp
	{
		pos1 := pos
		// AddOp
		if !_accept(parser, _AddOpAccepts, &pos, &perr) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// r:Product
	{
		pos2 := pos
		// Product
		if !_accept(parser, _ProductAccepts, &pos, &perr) {
			goto fail
		}
		labels[1] = parser.text[pos2:pos]
	}
	return _memoize(parser, _SumTail, start, pos, perr)
fail:
	return _memoize(parser, _SumTail, start, -1, perr)
}

func _SumTailNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [2]string
	use(labels)
	dp := parser.deltaPos[start][_SumTail]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _SumTail}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "SumTail"}
	// action
	// op:AddOp r:Product
	// op:AddOp
	{
		pos1 := pos
		// AddOp
		if !_node(parser, _AddOpNode, node, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// r:Product
	{
		pos2 := pos
		// Product
		if !_node(parser, _ProductNode, node, &pos) {
			goto fail
		}
		labels[1] = parser.text[pos2:pos]
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _SumTailFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [2]string
	use(labels)
	pos, failure := _failMemo(parser, _SumTail, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "SumTail",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _SumTail}
	// action
	// op:AddOp r:Product
	// op:AddOp
	{
		pos1 := pos
		// AddOp
		if !_fail(parser, _AddOpFail, errPos, failure, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// r:Product
	{
		pos2 := pos
		// Product
		if !_fail(parser, _ProductFail, errPos, failure, &pos) {
			goto fail
		}
		labels[1] = parser.text[pos2:pos]
	}
	parser.fail[key] = failure
	return pos, failure
fail:
	parser.fail[key] = failure
	return -1, failure
}

func _SumTailAction(parser *_Parser, start int) (int, *tail) {
	var labels [2]string
	use(labels)
	var label0 op
	var label1 (big.Float)
	dp := parser.deltaPos[start][_SumTail]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _SumTail}
	n := parser.act[key]
	if n != nil {
		n := n.(tail)
		return start + int(dp-1), &n
	}
	var node tail
	pos := start
	// action
	{
		start0 := pos
		// op:AddOp r:Product
		// op:AddOp
		{
			pos2 := pos
			// AddOp
			if p, n := _AddOpAction(parser, pos); n == nil {
				goto fail
			} else {
				label0 = *n
				pos = p
			}
			labels[0] = parser.text[pos2:pos]
		}
		// r:Product
		{
			pos3 := pos
			// Product
			if p, n := _ProductAction(parser, pos); n == nil {
				goto fail
			} else {
				label1 = *n
				pos = p
			}
			labels[1] = parser.text[pos3:pos]
		}
		node = func(
			start, end int, op op, r big.Float) tail {
			return tail{op, &r}
		}(
			start0, pos, label0, label1)
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _AddOpAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	if dp, de, ok := _memo(parser, _AddOp, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// _ "+" {…}/_ "-" {…}
	{
		pos3 := pos
		// action
		// _ "+"
		// _
		if !_accept(parser, __Accepts, &pos, &perr) {
			goto fail4
		}
		// "+"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "+" {
			perr = _max(perr, pos)
			goto fail4
		}
		pos++
		goto ok0
	fail4:
		pos = pos3
		// action
		// _ "-"
		// _
		if !_accept(parser, __Accepts, &pos, &perr) {
			goto fail6
		}
		// "-"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "-" {
			perr = _max(perr, pos)
			goto fail6
		}
		pos++
		goto ok0
	fail6:
		pos = pos3
		goto fail
	ok0:
	}
	perr = start
	return _memoize(parser, _AddOp, start, pos, perr)
fail:
	return _memoize(parser, _AddOp, start, -1, perr)
}

func _AddOpNode(parser *_Parser, start int) (int, *peg.Node) {
	dp := parser.deltaPos[start][_AddOp]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _AddOp}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "AddOp"}
	// _ "+" {…}/_ "-" {…}
	{
		pos3 := pos
		nkids1 := len(node.Kids)
		// action
		// _ "+"
		// _
		if !_node(parser, __Node, node, &pos) {
			goto fail4
		}
		// "+"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "+" {
			goto fail4
		}
		node.Kids = append(node.Kids, _leaf(parser, pos, pos+1))
		pos++
		goto ok0
	fail4:
		node.Kids = node.Kids[:nkids1]
		pos = pos3
		// action
		// _ "-"
		// _
		if !_node(parser, __Node, node, &pos) {
			goto fail6
		}
		// "-"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "-" {
			goto fail6
		}
		node.Kids = append(node.Kids, _leaf(parser, pos, pos+1))
		pos++
		goto ok0
	fail6:
		node.Kids = node.Kids[:nkids1]
		pos = pos3
		goto fail
	ok0:
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _AddOpFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	pos, failure := _failMemo(parser, _AddOp, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "AddOp",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _AddOp}
	// _ "+" {…}/_ "-" {…}
	{
		pos3 := pos
		// action
		// _ "+"
		// _
		if !_fail(parser, __Fail, errPos, failure, &pos) {
			goto fail4
		}
		// "+"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "+" {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "\"+\"",
				})
			}
			goto fail4
		}
		pos++
		goto ok0
	fail4:
		pos = pos3
		// action
		// _ "-"
		// _
		if !_fail(parser, __Fail, errPos, failure, &pos) {
			goto fail6
		}
		// "-"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "-" {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "\"-\"",
				})
			}
			goto fail6
		}
		pos++
		goto ok0
	fail6:
		pos = pos3
		goto fail
	ok0:
	}
	failure.Kids = nil
	parser.fail[key] = failure
	return pos, failure
fail:
	failure.Kids = nil
	failure.Want = "operator"
	parser.fail[key] = failure
	return -1, failure
}

func _AddOpAction(parser *_Parser, start int) (int, *op) {
	dp := parser.deltaPos[start][_AddOp]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _AddOp}
	n := parser.act[key]
	if n != nil {
		n := n.(op)
		return start + int(dp-1), &n
	}
	var node op
	pos := start
	// _ "+" {…}/_ "-" {…}
	{
		pos3 := pos
		var node2 op
		// action
		{
			start5 := pos
			// _ "+"
			// _
			if p, n := __Action(parser, pos); n == nil {
				goto fail4
			} else {
				pos = p
			}
			// "+"
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "+" {
				goto fail4
			}
			pos++
			node = func(
				start, end int) op {
				return op((*big.Float).Add)
			}(
				start5, pos)
		}
		goto ok0
	fail4:
		node = node2
		pos = pos3
		// action
		{
			start8 := pos
			// _ "-"
			// _
			if p, n := __Action(parser, pos); n == nil {
				goto fail7
			} else {
				pos = p
			}
			// "-"
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "-" {
				goto fail7
			}
			pos++
			node = func(
				start, end int) op {
				return op((*big.Float).Sub)
			}(
				start8, pos)
		}
		goto ok0
	fail7:
		node = node2
		pos = pos3
		goto fail
	ok0:
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _ProductAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [2]string
	use(labels)
	if dp, de, ok := _memo(parser, _Product, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// action
	// l:Value tail:ProductTail*
	// l:Value
	{
		pos1 := pos
		// Value
		if !_accept(parser, _ValueAccepts, &pos, &perr) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// tail:ProductTail*
	{
		pos2 := pos
		// ProductTail*
		for {
			pos4 := pos
			// ProductTail
			if !_accept(parser, _ProductTailAccepts, &pos, &perr) {
				goto fail6
			}
			continue
		fail6:
			pos = pos4
			break
		}
		labels[1] = parser.text[pos2:pos]
	}
	return _memoize(parser, _Product, start, pos, perr)
fail:
	return _memoize(parser, _Product, start, -1, perr)
}

func _ProductNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [2]string
	use(labels)
	dp := parser.deltaPos[start][_Product]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Product}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "Product"}
	// action
	// l:Value tail:ProductTail*
	// l:Value
	{
		pos1 := pos
		// Value
		if !_node(parser, _ValueNode, node, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// tail:ProductTail*
	{
		pos2 := pos
		// ProductTail*
		for {
			nkids3 := len(node.Kids)
			pos4 := pos
			// ProductTail
			if !_node(parser, _ProductTailNode, node, &pos) {
				goto fail6
			}
			continue
		fail6:
			node.Kids = node.Kids[:nkids3]
			pos = pos4
			break
		}
		labels[1] = parser.text[pos2:pos]
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _ProductFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [2]string
	use(labels)
	pos, failure := _failMemo(parser, _Product, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "Product",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _Product}
	// action
	// l:Value tail:ProductTail*
	// l:Value
	{
		pos1 := pos
		// Value
		if !_fail(parser, _ValueFail, errPos, failure, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// tail:ProductTail*
	{
		pos2 := pos
		// ProductTail*
		for {
			pos4 := pos
			// ProductTail
			if !_fail(parser, _ProductTailFail, errPos, failure, &pos) {
				goto fail6
			}
			continue
		fail6:
			pos = pos4
			break
		}
		labels[1] = parser.text[pos2:pos]
	}
	parser.fail[key] = failure
	return pos, failure
fail:
	parser.fail[key] = failure
	return -1, failure
}

func _ProductAction(parser *_Parser, start int) (int, *(big.Float)) {
	var labels [2]string
	use(labels)
	var label0 (big.Float)
	var label1 []tail
	dp := parser.deltaPos[start][_Product]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Product}
	n := parser.act[key]
	if n != nil {
		n := n.((big.Float))
		return start + int(dp-1), &n
	}
	var node (big.Float)
	pos := start
	// action
	{
		start0 := pos
		// l:Value tail:ProductTail*
		// l:Value
		{
			pos2 := pos
			// Value
			if p, n := _ValueAction(parser, pos); n == nil {
				goto fail
			} else {
				label0 = *n
				pos = p
			}
			labels[0] = parser.text[pos2:pos]
		}
		// tail:ProductTail*
		{
			pos3 := pos
			// ProductTail*
			for {
				pos5 := pos
				var node6 tail
				// ProductTail
				if p, n := _ProductTailAction(parser, pos); n == nil {
					goto fail7
				} else {
					node6 = *n
					pos = p
				}
				label1 = append(label1, node6)
				continue
			fail7:
				pos = pos5
				break
			}
			labels[1] = parser.text[pos3:pos]
		}
		node = func(
			start, end int, l big.Float, tail []tail) big.Float {
			return (big.Float)(evalTail(l, tail))
		}(
			start0, pos, label0, label1)
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _ProductTailAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [2]string
	use(labels)
	if dp, de, ok := _memo(parser, _ProductTail, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// action
	// op:MulOp r:Value
	// op:MulOp
	{
		pos1 := pos
		// MulOp
		if !_accept(parser, _MulOpAccepts, &pos, &perr) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// r:Value
	{
		pos2 := pos
		// Value
		if !_accept(parser, _ValueAccepts, &pos, &perr) {
			goto fail
		}
		labels[1] = parser.text[pos2:pos]
	}
	return _memoize(parser, _ProductTail, start, pos, perr)
fail:
	return _memoize(parser, _ProductTail, start, -1, perr)
}

func _ProductTailNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [2]string
	use(labels)
	dp := parser.deltaPos[start][_ProductTail]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _ProductTail}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "ProductTail"}
	// action
	// op:MulOp r:Value
	// op:MulOp
	{
		pos1 := pos
		// MulOp
		if !_node(parser, _MulOpNode, node, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// r:Value
	{
		pos2 := pos
		// Value
		if !_node(parser, _ValueNode, node, &pos) {
			goto fail
		}
		labels[1] = parser.text[pos2:pos]
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _ProductTailFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [2]string
	use(labels)
	pos, failure := _failMemo(parser, _ProductTail, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "ProductTail",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _ProductTail}
	// action
	// op:MulOp r:Value
	// op:MulOp
	{
		pos1 := pos
		// MulOp
		if !_fail(parser, _MulOpFail, errPos, failure, &pos) {
			goto fail
		}
		labels[0] = parser.text[pos1:pos]
	}
	// r:Value
	{
		pos2 := pos
		// Value
		if !_fail(parser, _ValueFail, errPos, failure, &pos) {
			goto fail
		}
		labels[1] = parser.text[pos2:pos]
	}
	parser.fail[key] = failure
	return pos, failure
fail:
	parser.fail[key] = failure
	return -1, failure
}

func _ProductTailAction(parser *_Parser, start int) (int, *tail) {
	var labels [2]string
	use(labels)
	var label0 op
	var label1 (big.Float)
	dp := parser.deltaPos[start][_ProductTail]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _ProductTail}
	n := parser.act[key]
	if n != nil {
		n := n.(tail)
		return start + int(dp-1), &n
	}
	var node tail
	pos := start
	// action
	{
		start0 := pos
		// op:MulOp r:Value
		// op:MulOp
		{
			pos2 := pos
			// MulOp
			if p, n := _MulOpAction(parser, pos); n == nil {
				goto fail
			} else {
				label0 = *n
				pos = p
			}
			labels[0] = parser.text[pos2:pos]
		}
		// r:Value
		{
			pos3 := pos
			// Value
			if p, n := _ValueAction(parser, pos); n == nil {
				goto fail
			} else {
				label1 = *n
				pos = p
			}
			labels[1] = parser.text[pos3:pos]
		}
		node = func(
			start, end int, op op, r big.Float) tail {
			return tail{op, &r}
		}(
			start0, pos, label0, label1)
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _MulOpAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	if dp, de, ok := _memo(parser, _MulOp, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// _ "*" {…}/_ "/" {…}
	{
		pos3 := pos
		// action
		// _ "*"
		// _
		if !_accept(parser, __Accepts, &pos, &perr) {
			goto fail4
		}
		// "*"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "*" {
			perr = _max(perr, pos)
			goto fail4
		}
		pos++
		goto ok0
	fail4:
		pos = pos3
		// action
		// _ "/"
		// _
		if !_accept(parser, __Accepts, &pos, &perr) {
			goto fail6
		}
		// "/"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "/" {
			perr = _max(perr, pos)
			goto fail6
		}
		pos++
		goto ok0
	fail6:
		pos = pos3
		goto fail
	ok0:
	}
	perr = start
	return _memoize(parser, _MulOp, start, pos, perr)
fail:
	return _memoize(parser, _MulOp, start, -1, perr)
}

func _MulOpNode(parser *_Parser, start int) (int, *peg.Node) {
	dp := parser.deltaPos[start][_MulOp]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _MulOp}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "MulOp"}
	// _ "*" {…}/_ "/" {…}
	{
		pos3 := pos
		nkids1 := len(node.Kids)
		// action
		// _ "*"
		// _
		if !_node(parser, __Node, node, &pos) {
			goto fail4
		}
		// "*"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "*" {
			goto fail4
		}
		node.Kids = append(node.Kids, _leaf(parser, pos, pos+1))
		pos++
		goto ok0
	fail4:
		node.Kids = node.Kids[:nkids1]
		pos = pos3
		// action
		// _ "/"
		// _
		if !_node(parser, __Node, node, &pos) {
			goto fail6
		}
		// "/"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "/" {
			goto fail6
		}
		node.Kids = append(node.Kids, _leaf(parser, pos, pos+1))
		pos++
		goto ok0
	fail6:
		node.Kids = node.Kids[:nkids1]
		pos = pos3
		goto fail
	ok0:
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _MulOpFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	pos, failure := _failMemo(parser, _MulOp, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "MulOp",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _MulOp}
	// _ "*" {…}/_ "/" {…}
	{
		pos3 := pos
		// action
		// _ "*"
		// _
		if !_fail(parser, __Fail, errPos, failure, &pos) {
			goto fail4
		}
		// "*"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "*" {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "\"*\"",
				})
			}
			goto fail4
		}
		pos++
		goto ok0
	fail4:
		pos = pos3
		// action
		// _ "/"
		// _
		if !_fail(parser, __Fail, errPos, failure, &pos) {
			goto fail6
		}
		// "/"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "/" {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "\"/\"",
				})
			}
			goto fail6
		}
		pos++
		goto ok0
	fail6:
		pos = pos3
		goto fail
	ok0:
	}
	failure.Kids = nil
	parser.fail[key] = failure
	return pos, failure
fail:
	failure.Kids = nil
	failure.Want = "operator"
	parser.fail[key] = failure
	return -1, failure
}

func _MulOpAction(parser *_Parser, start int) (int, *op) {
	dp := parser.deltaPos[start][_MulOp]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _MulOp}
	n := parser.act[key]
	if n != nil {
		n := n.(op)
		return start + int(dp-1), &n
	}
	var node op
	pos := start
	// _ "*" {…}/_ "/" {…}
	{
		pos3 := pos
		var node2 op
		// action
		{
			start5 := pos
			// _ "*"
			// _
			if p, n := __Action(parser, pos); n == nil {
				goto fail4
			} else {
				pos = p
			}
			// "*"
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "*" {
				goto fail4
			}
			pos++
			node = func(
				start, end int) op {
				return op((*big.Float).Mul)
			}(
				start5, pos)
		}
		goto ok0
	fail4:
		node = node2
		pos = pos3
		// action
		{
			start8 := pos
			// _ "/"
			// _
			if p, n := __Action(parser, pos); n == nil {
				goto fail7
			} else {
				pos = p
			}
			// "/"
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "/" {
				goto fail7
			}
			pos++
			node = func(
				start, end int) op {
				return op((*big.Float).Quo)
			}(
				start8, pos)
		}
		goto ok0
	fail7:
		node = node2
		pos = pos3
		goto fail
	ok0:
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _ValueAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [1]string
	use(labels)
	if dp, de, ok := _memo(parser, _Value, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// Num/_ "(" e:Sum _ ")" {…}
	{
		pos3 := pos
		// Num
		if !_accept(parser, _NumAccepts, &pos, &perr) {
			goto fail4
		}
		goto ok0
	fail4:
		pos = pos3
		// action
		// _ "(" e:Sum _ ")"
		// _
		if !_accept(parser, __Accepts, &pos, &perr) {
			goto fail5
		}
		// "("
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "(" {
			perr = _max(perr, pos)
			goto fail5
		}
		pos++
		// e:Sum
		{
			pos7 := pos
			// Sum
			if !_accept(parser, _SumAccepts, &pos, &perr) {
				goto fail5
			}
			labels[0] = parser.text[pos7:pos]
		}
		// _
		if !_accept(parser, __Accepts, &pos, &perr) {
			goto fail5
		}
		// ")"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != ")" {
			perr = _max(perr, pos)
			goto fail5
		}
		pos++
		goto ok0
	fail5:
		pos = pos3
		goto fail
	ok0:
	}
	return _memoize(parser, _Value, start, pos, perr)
fail:
	return _memoize(parser, _Value, start, -1, perr)
}

func _ValueNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [1]string
	use(labels)
	dp := parser.deltaPos[start][_Value]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Value}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "Value"}
	// Num/_ "(" e:Sum _ ")" {…}
	{
		pos3 := pos
		nkids1 := len(node.Kids)
		// Num
		if !_node(parser, _NumNode, node, &pos) {
			goto fail4
		}
		goto ok0
	fail4:
		node.Kids = node.Kids[:nkids1]
		pos = pos3
		// action
		// _ "(" e:Sum _ ")"
		// _
		if !_node(parser, __Node, node, &pos) {
			goto fail5
		}
		// "("
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "(" {
			goto fail5
		}
		node.Kids = append(node.Kids, _leaf(parser, pos, pos+1))
		pos++
		// e:Sum
		{
			pos7 := pos
			// Sum
			if !_node(parser, _SumNode, node, &pos) {
				goto fail5
			}
			labels[0] = parser.text[pos7:pos]
		}
		// _
		if !_node(parser, __Node, node, &pos) {
			goto fail5
		}
		// ")"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != ")" {
			goto fail5
		}
		node.Kids = append(node.Kids, _leaf(parser, pos, pos+1))
		pos++
		goto ok0
	fail5:
		node.Kids = node.Kids[:nkids1]
		pos = pos3
		goto fail
	ok0:
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _ValueFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [1]string
	use(labels)
	pos, failure := _failMemo(parser, _Value, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "Value",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _Value}
	// Num/_ "(" e:Sum _ ")" {…}
	{
		pos3 := pos
		// Num
		if !_fail(parser, _NumFail, errPos, failure, &pos) {
			goto fail4
		}
		goto ok0
	fail4:
		pos = pos3
		// action
		// _ "(" e:Sum _ ")"
		// _
		if !_fail(parser, __Fail, errPos, failure, &pos) {
			goto fail5
		}
		// "("
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "(" {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "\"(\"",
				})
			}
			goto fail5
		}
		pos++
		// e:Sum
		{
			pos7 := pos
			// Sum
			if !_fail(parser, _SumFail, errPos, failure, &pos) {
				goto fail5
			}
			labels[0] = parser.text[pos7:pos]
		}
		// _
		if !_fail(parser, __Fail, errPos, failure, &pos) {
			goto fail5
		}
		// ")"
		if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != ")" {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "\")\"",
				})
			}
			goto fail5
		}
		pos++
		goto ok0
	fail5:
		pos = pos3
		goto fail
	ok0:
	}
	parser.fail[key] = failure
	return pos, failure
fail:
	parser.fail[key] = failure
	return -1, failure
}

func _ValueAction(parser *_Parser, start int) (int, *(big.Float)) {
	var labels [1]string
	use(labels)
	var label0 (big.Float)
	dp := parser.deltaPos[start][_Value]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Value}
	n := parser.act[key]
	if n != nil {
		n := n.((big.Float))
		return start + int(dp-1), &n
	}
	var node (big.Float)
	pos := start
	// Num/_ "(" e:Sum _ ")" {…}
	{
		pos3 := pos
		var node2 (big.Float)
		// Num
		if p, n := _NumAction(parser, pos); n == nil {
			goto fail4
		} else {
			node = *n
			pos = p
		}
		goto ok0
	fail4:
		node = node2
		pos = pos3
		// action
		{
			start6 := pos
			// _ "(" e:Sum _ ")"
			// _
			if p, n := __Action(parser, pos); n == nil {
				goto fail5
			} else {
				pos = p
			}
			// "("
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "(" {
				goto fail5
			}
			pos++
			// e:Sum
			{
				pos8 := pos
				// Sum
				if p, n := _SumAction(parser, pos); n == nil {
					goto fail5
				} else {
					label0 = *n
					pos = p
				}
				labels[0] = parser.text[pos8:pos]
			}
			// _
			if p, n := __Action(parser, pos); n == nil {
				goto fail5
			} else {
				pos = p
			}
			// ")"
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != ")" {
				goto fail5
			}
			pos++
			node = func(
				start, end int, e big.Float) big.Float {
				return (big.Float)(e)
			}(
				start6, pos, label0)
		}
		goto ok0
	fail5:
		node = node2
		pos = pos3
		goto fail
	ok0:
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func _NumAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [1]string
	use(labels)
	if dp, de, ok := _memo(parser, _Num, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// action
	// _ n:([0-9]+ ("." [0-9]+)?)
	// _
	if !_accept(parser, __Accepts, &pos, &perr) {
		goto fail
	}
	// n:([0-9]+ ("." [0-9]+)?)
	{
		pos1 := pos
		// ([0-9]+ ("." [0-9]+)?)
		// [0-9]+ ("." [0-9]+)?
		// [0-9]+
		// [0-9]
		if r, w := _next(parser, pos); r < '0' || r > '9' {
			perr = _max(perr, pos)
			goto fail
		} else {
			pos += w
		}
		for {
			pos4 := pos
			// [0-9]
			if r, w := _next(parser, pos); r < '0' || r > '9' {
				perr = _max(perr, pos)
				goto fail6
			} else {
				pos += w
			}
			continue
		fail6:
			pos = pos4
			break
		}
		// ("." [0-9]+)?
		{
			pos8 := pos
			// ("." [0-9]+)
			// "." [0-9]+
			// "."
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "." {
				perr = _max(perr, pos)
				goto fail9
			}
			pos++
			// [0-9]+
			// [0-9]
			if r, w := _next(parser, pos); r < '0' || r > '9' {
				perr = _max(perr, pos)
				goto fail9
			} else {
				pos += w
			}
			for {
				pos12 := pos
				// [0-9]
				if r, w := _next(parser, pos); r < '0' || r > '9' {
					perr = _max(perr, pos)
					goto fail14
				} else {
					pos += w
				}
				continue
			fail14:
				pos = pos12
				break
			}
			goto ok15
		fail9:
			pos = pos8
		ok15:
		}
		labels[0] = parser.text[pos1:pos]
	}
	perr = start
	return _memoize(parser, _Num, start, pos, perr)
fail:
	return _memoize(parser, _Num, start, -1, perr)
}

func _NumNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [1]string
	use(labels)
	dp := parser.deltaPos[start][_Num]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Num}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "Num"}
	// action
	// _ n:([0-9]+ ("." [0-9]+)?)
	// _
	if !_node(parser, __Node, node, &pos) {
		goto fail
	}
	// n:([0-9]+ ("." [0-9]+)?)
	{
		pos1 := pos
		// ([0-9]+ ("." [0-9]+)?)
		{
			nkids2 := len(node.Kids)
			pos03 := pos
			// [0-9]+ ("." [0-9]+)?
			// [0-9]+
			// [0-9]
			if r, w := _next(parser, pos); r < '0' || r > '9' {
				goto fail
			} else {
				node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
				pos += w
			}
			for {
				nkids5 := len(node.Kids)
				pos6 := pos
				// [0-9]
				if r, w := _next(parser, pos); r < '0' || r > '9' {
					goto fail8
				} else {
					node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
					pos += w
				}
				continue
			fail8:
				node.Kids = node.Kids[:nkids5]
				pos = pos6
				break
			}
			// ("." [0-9]+)?
			{
				nkids9 := len(node.Kids)
				pos10 := pos
				// ("." [0-9]+)
				{
					nkids12 := len(node.Kids)
					pos013 := pos
					// "." [0-9]+
					// "."
					if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "." {
						goto fail11
					}
					node.Kids = append(node.Kids, _leaf(parser, pos, pos+1))
					pos++
					// [0-9]+
					// [0-9]
					if r, w := _next(parser, pos); r < '0' || r > '9' {
						goto fail11
					} else {
						node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
						pos += w
					}
					for {
						nkids15 := len(node.Kids)
						pos16 := pos
						// [0-9]
						if r, w := _next(parser, pos); r < '0' || r > '9' {
							goto fail18
						} else {
							node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
							pos += w
						}
						continue
					fail18:
						node.Kids = node.Kids[:nkids15]
						pos = pos16
						break
					}
					sub := _sub(parser, pos013, pos, node.Kids[nkids12:])
					node.Kids = append(node.Kids[:nkids12], sub)
				}
				goto ok19
			fail11:
				node.Kids = node.Kids[:nkids9]
				pos = pos10
			ok19:
			}
			sub := _sub(parser, pos03, pos, node.Kids[nkids2:])
			node.Kids = append(node.Kids[:nkids2], sub)
		}
		labels[0] = parser.text[pos1:pos]
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _NumFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [1]string
	use(labels)
	pos, failure := _failMemo(parser, _Num, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "Num",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _Num}
	// action
	// _ n:([0-9]+ ("." [0-9]+)?)
	// _
	if !_fail(parser, __Fail, errPos, failure, &pos) {
		goto fail
	}
	// n:([0-9]+ ("." [0-9]+)?)
	{
		pos1 := pos
		// ([0-9]+ ("." [0-9]+)?)
		// [0-9]+ ("." [0-9]+)?
		// [0-9]+
		// [0-9]
		if r, w := _next(parser, pos); r < '0' || r > '9' {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "[0-9]",
				})
			}
			goto fail
		} else {
			pos += w
		}
		for {
			pos4 := pos
			// [0-9]
			if r, w := _next(parser, pos); r < '0' || r > '9' {
				if pos >= errPos {
					failure.Kids = append(failure.Kids, &peg.Fail{
						Pos:  int(pos),
						Want: "[0-9]",
					})
				}
				goto fail6
			} else {
				pos += w
			}
			continue
		fail6:
			pos = pos4
			break
		}
		// ("." [0-9]+)?
		{
			pos8 := pos
			// ("." [0-9]+)
			// "." [0-9]+
			// "."
			if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "." {
				if pos >= errPos {
					failure.Kids = append(failure.Kids, &peg.Fail{
						Pos:  int(pos),
						Want: "\".\"",
					})
				}
				goto fail9
			}
			pos++
			// [0-9]+
			// [0-9]
			if r, w := _next(parser, pos); r < '0' || r > '9' {
				if pos >= errPos {
					failure.Kids = append(failure.Kids, &peg.Fail{
						Pos:  int(pos),
						Want: "[0-9]",
					})
				}
				goto fail9
			} else {
				pos += w
			}
			for {
				pos12 := pos
				// [0-9]
				if r, w := _next(parser, pos); r < '0' || r > '9' {
					if pos >= errPos {
						failure.Kids = append(failure.Kids, &peg.Fail{
							Pos:  int(pos),
							Want: "[0-9]",
						})
					}
					goto fail14
				} else {
					pos += w
				}
				continue
			fail14:
				pos = pos12
				break
			}
			goto ok15
		fail9:
			pos = pos8
		ok15:
		}
		labels[0] = parser.text[pos1:pos]
	}
	failure.Kids = nil
	parser.fail[key] = failure
	return pos, failure
fail:
	failure.Kids = nil
	failure.Want = "number"
	parser.fail[key] = failure
	return -1, failure
}

func _NumAction(parser *_Parser, start int) (int, *(big.Float)) {
	var labels [1]string
	use(labels)
	var label0 string
	dp := parser.deltaPos[start][_Num]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Num}
	n := parser.act[key]
	if n != nil {
		n := n.((big.Float))
		return start + int(dp-1), &n
	}
	var node (big.Float)
	pos := start
	// action
	{
		start0 := pos
		// _ n:([0-9]+ ("." [0-9]+)?)
		// _
		if p, n := __Action(parser, pos); n == nil {
			goto fail
		} else {
			pos = p
		}
		// n:([0-9]+ ("." [0-9]+)?)
		{
			pos2 := pos
			// ([0-9]+ ("." [0-9]+)?)
			// [0-9]+ ("." [0-9]+)?
			{
				var node3 string
				// [0-9]+
				{
					var node6 string
					// [0-9]
					if r, w := _next(parser, pos); r < '0' || r > '9' {
						goto fail
					} else {
						node6 = parser.text[pos : pos+w]
						pos += w
					}
					node3 += node6
				}
				for {
					pos5 := pos
					var node6 string
					// [0-9]
					if r, w := _next(parser, pos); r < '0' || r > '9' {
						goto fail7
					} else {
						node6 = parser.text[pos : pos+w]
						pos += w
					}
					node3 += node6
					continue
				fail7:
					pos = pos5
					break
				}
				label0, node3 = label0+node3, ""
				// ("." [0-9]+)?
				{
					pos9 := pos
					// ("." [0-9]+)
					// "." [0-9]+
					{
						var node11 string
						// "."
						if len(parser.text[pos:]) < 1 || parser.text[pos:pos+1] != "." {
							goto fail10
						}
						node11 = parser.text[pos : pos+1]
						pos++
						node3, node11 = node3+node11, ""
						// [0-9]+
						{
							var node14 string
							// [0-9]
							if r, w := _next(parser, pos); r < '0' || r > '9' {
								goto fail10
							} else {
								node14 = parser.text[pos : pos+w]
								pos += w
							}
							node11 += node14
						}
						for {
							pos13 := pos
							var node14 string
							// [0-9]
							if r, w := _next(parser, pos); r < '0' || r > '9' {
								goto fail15
							} else {
								node14 = parser.text[pos : pos+w]
								pos += w
							}
							node11 += node14
							continue
						fail15:
							pos = pos13
							break
						}
						node3, node11 = node3+node11, ""
					}
					goto ok16
				fail10:
					node3 = ""
					pos = pos9
				ok16:
				}
				label0, node3 = label0+node3, ""
			}
			labels[0] = parser.text[pos2:pos]
		}
		node = func(
			start, end int, n string) big.Float {
			var f big.Float
			f.Parse(n, 10)
			return (big.Float)(f)
		}(
			start0, pos, label0)
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}

func __Accepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	var labels [1]string
	use(labels)
	if dp, de, ok := _memo(parser, __, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// (s:. &{…})*
	for {
		pos1 := pos
		// (s:. &{…})
		// s:. &{…}
		// s:.
		{
			pos5 := pos
			// .
			if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
				perr = _max(perr, pos)
				goto fail3
			} else {
				pos += w
			}
			labels[0] = parser.text[pos5:pos]
		}
		// pred code
		if ok := func(s string) bool { return isSpace(s) }(labels[0]); !ok {
			perr = _max(perr, pos)
			goto fail3
		}
		continue
	fail3:
		pos = pos1
		break
	}
	perr = start
	return _memoize(parser, __, start, pos, perr)
}

func __Node(parser *_Parser, start int) (int, *peg.Node) {
	var labels [1]string
	use(labels)
	dp := parser.deltaPos[start][__]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: __}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "_"}
	// (s:. &{…})*
	for {
		nkids0 := len(node.Kids)
		pos1 := pos
		// (s:. &{…})
		{
			nkids4 := len(node.Kids)
			pos05 := pos
			// s:. &{…}
			// s:.
			{
				pos7 := pos
				// .
				if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
					goto fail3
				} else {
					node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
					pos += w
				}
				labels[0] = parser.text[pos7:pos]
			}
			// pred code
			if ok := func(s string) bool { return isSpace(s) }(labels[0]); !ok {
				goto fail3
			}
			sub := _sub(parser, pos05, pos, node.Kids[nkids4:])
			node.Kids = append(node.Kids[:nkids4], sub)
		}
		continue
	fail3:
		node.Kids = node.Kids[:nkids0]
		pos = pos1
		break
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
}

func __Fail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [1]string
	use(labels)
	pos, failure := _failMemo(parser, __, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "_",
		Pos:  int(start),
	}
	key := _key{start: start, rule: __}
	// (s:. &{…})*
	for {
		pos1 := pos
		// (s:. &{…})
		// s:. &{…}
		// s:.
		{
			pos5 := pos
			// .
			if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
				if pos >= errPos {
					failure.Kids = append(failure.Kids, &peg.Fail{
						Pos:  int(pos),
						Want: ".",
					})
				}
				goto fail3
			} else {
				pos += w
			}
			labels[0] = parser.text[pos5:pos]
		}
		// pred code
		if ok := func(s string) bool { return isSpace(s) }(labels[0]); !ok {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: "&{" + " isSpace(s) " + "}",
				})
			}
			goto fail3
		}
		continue
	fail3:
		pos = pos1
		break
	}
	failure.Kids = nil
	parser.fail[key] = failure
	return pos, failure
}

func __Action(parser *_Parser, start int) (int, *string) {
	var labels [1]string
	use(labels)
	var label0 string
	dp := parser.deltaPos[start][__]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: __}
	n := parser.act[key]
	if n != nil {
		n := n.(string)
		return start + int(dp-1), &n
	}
	var node string
	pos := start
	// (s:. &{…})*
	for {
		pos1 := pos
		var node2 string
		// (s:. &{…})
		// s:. &{…}
		{
			var node4 string
			// s:.
			{
				pos5 := pos
				// .
				if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
					goto fail3
				} else {
					label0 = parser.text[pos : pos+w]
					pos += w
				}
				node4 = label0
				labels[0] = parser.text[pos5:pos]
			}
			node2, node4 = node2+node4, ""
			// pred code
			if ok := func(s string) bool { return isSpace(s) }(labels[0]); !ok {
				goto fail3
			}
			node4 = ""
			node2, node4 = node2+node4, ""
		}
		node += node2
		continue
	fail3:
		pos = pos1
		break
	}
	parser.act[key] = node
	return pos, &node
}

func _EOFAccepts(parser *_Parser, start int) (deltaPos, deltaErr int) {
	if dp, de, ok := _memo(parser, _EOF, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// !.
	{
		pos1 := pos
		perr3 := perr
		// .
		if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
			perr = _max(perr, pos)
			goto ok0
		} else {
			pos += w
		}
		pos = pos1
		perr = _max(perr3, pos)
		goto fail
	ok0:
		pos = pos1
		perr = perr3
	}
	perr = start
	return _memoize(parser, _EOF, start, pos, perr)
fail:
	return _memoize(parser, _EOF, start, -1, perr)
}

func _EOFNode(parser *_Parser, start int) (int, *peg.Node) {
	dp := parser.deltaPos[start][_EOF]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _EOF}
	node := parser.node[key]
	if node != nil {
		return start + int(dp-1), node
	}
	pos := start
	node = &peg.Node{Name: "EOF"}
	// !.
	{
		pos1 := pos
		nkids2 := len(node.Kids)
		// .
		if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
			goto ok0
		} else {
			node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
			pos += w
		}
		pos = pos1
		node.Kids = node.Kids[:nkids2]
		goto fail
	ok0:
		pos = pos1
		node.Kids = node.Kids[:nkids2]
	}
	node.Text = parser.text[start:pos]
	parser.node[key] = node
	return pos, node
fail:
	return -1, nil
}

func _EOFFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	pos, failure := _failMemo(parser, _EOF, start, errPos)
	if failure != nil {
		return pos, failure
	}
	failure = &peg.Fail{
		Name: "EOF",
		Pos:  int(start),
	}
	key := _key{start: start, rule: _EOF}
	// !.
	{
		pos1 := pos
		nkids2 := len(failure.Kids)
		// .
		if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
			if pos >= errPos {
				failure.Kids = append(failure.Kids, &peg.Fail{
					Pos:  int(pos),
					Want: ".",
				})
			}
			goto ok0
		} else {
			pos += w
		}
		pos = pos1
		failure.Kids = failure.Kids[:nkids2]
		if pos >= errPos {
			failure.Kids = append(failure.Kids, &peg.Fail{
				Pos:  int(pos),
				Want: "!.",
			})
		}
		goto fail
	ok0:
		pos = pos1
		failure.Kids = failure.Kids[:nkids2]
	}
	failure.Kids = nil
	parser.fail[key] = failure
	return pos, failure
fail:
	failure.Kids = nil
	failure.Want = "end of file"
	parser.fail[key] = failure
	return -1, failure
}

func _EOFAction(parser *_Parser, start int) (int, *string) {
	dp := parser.deltaPos[start][_EOF]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _EOF}
	n := parser.act[key]
	if n != nil {
		n := n.(string)
		return start + int(dp-1), &n
	}
	var node string
	pos := start
	// !.
	{
		pos1 := pos
		// .
		if r, w := _next(parser, pos); w == 0 || r == '\uFFFD' {
			goto ok0
		} else {
			pos += w
		}
		pos = pos1
		goto fail
	ok0:
		pos = pos1
		node = ""
	}
	parser.act[key] = node
	return pos, &node
fail:
	return -1, nil
}
