// Test labels with the same name but in different choice branches.
// 	peggy -o label_names.go label_names.peggy
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/eaburns/peggy/peg"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		p, err := _NewParser(line)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if pos, perr := _ExprAccepts(p, 0); pos < 0 {
			_, fail := _ExprFail(p, 0, perr)
			fmt.Println(peg.SimpleError(line, fail))
			continue
		}
		_, result := _ExprAction(p, 0)
		fmt.Println(*result)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const (
	_Expr int = 0

	_N int = 1
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

type tooBigError struct{}

func (tooBigError) Error() string { return "input is too big" }

func _NewParser(text string) (*_Parser, error) {
	n := len(text) + 1
	if n < 0 {
		return nil, tooBigError{}
	}
	p := &_Parser{
		text:     text,
		deltaPos: make([][_N]int32, n),
		deltaErr: make([][_N]int32, n),
		node:     make(map[_key]*peg.Node),
		fail:     make(map[_key]*peg.Fail),
		act:      make(map[_key]interface{}),
	}
	return p, nil
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
	var labels [2]string
	use(labels)
	if dp, de, ok := _memo(parser, _Expr, start); ok {
		return dp, de
	}
	pos, perr := start, -1
	// letter:[a] {…}/letter:[b] {…}
	{
		pos3 := pos
		// action
		// letter:[a]
		{
			pos5 := pos
			// [a]
			if r, w := _next(parser, pos); r != 'a' {
				perr = _max(perr, pos)
				goto fail4
			} else {
				pos += w
			}
			labels[0] = parser.text[pos5:pos]
		}
		goto ok0
	fail4:
		pos = pos3
		// action
		// letter:[b]
		{
			pos7 := pos
			// [b]
			if r, w := _next(parser, pos); r != 'b' {
				perr = _max(perr, pos)
				goto fail6
			} else {
				pos += w
			}
			labels[1] = parser.text[pos7:pos]
		}
		goto ok0
	fail6:
		pos = pos3
		goto fail
	ok0:
	}
	return _memoize(parser, _Expr, start, pos, perr)
fail:
	return _memoize(parser, _Expr, start, -1, perr)
}

func _ExprNode(parser *_Parser, start int) (int, *peg.Node) {
	var labels [2]string
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
	// letter:[a] {…}/letter:[b] {…}
	{
		pos3 := pos
		nkids1 := len(node.Kids)
		// action
		// letter:[a]
		{
			pos5 := pos
			// [a]
			if r, w := _next(parser, pos); r != 'a' {
				goto fail4
			} else {
				node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
				pos += w
			}
			labels[0] = parser.text[pos5:pos]
		}
		goto ok0
	fail4:
		node.Kids = node.Kids[:nkids1]
		pos = pos3
		// action
		// letter:[b]
		{
			pos7 := pos
			// [b]
			if r, w := _next(parser, pos); r != 'b' {
				goto fail6
			} else {
				node.Kids = append(node.Kids, _leaf(parser, pos, pos+w))
				pos += w
			}
			labels[1] = parser.text[pos7:pos]
		}
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

func _ExprFail(parser *_Parser, start, errPos int) (int, *peg.Fail) {
	var labels [2]string
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
	// letter:[a] {…}/letter:[b] {…}
	{
		pos3 := pos
		// action
		// letter:[a]
		{
			pos5 := pos
			// [a]
			if r, w := _next(parser, pos); r != 'a' {
				if pos >= errPos {
					failure.Kids = append(failure.Kids, &peg.Fail{
						Pos:  int(pos),
						Want: "[a]",
					})
				}
				goto fail4
			} else {
				pos += w
			}
			labels[0] = parser.text[pos5:pos]
		}
		goto ok0
	fail4:
		pos = pos3
		// action
		// letter:[b]
		{
			pos7 := pos
			// [b]
			if r, w := _next(parser, pos); r != 'b' {
				if pos >= errPos {
					failure.Kids = append(failure.Kids, &peg.Fail{
						Pos:  int(pos),
						Want: "[b]",
					})
				}
				goto fail6
			} else {
				pos += w
			}
			labels[1] = parser.text[pos7:pos]
		}
		goto ok0
	fail6:
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

func _ExprAction(parser *_Parser, start int) (int, *string) {
	var labels [2]string
	use(labels)
	var label0 string
	var label1 string
	dp := parser.deltaPos[start][_Expr]
	if dp < 0 {
		return -1, nil
	}
	key := _key{start: start, rule: _Expr}
	n := parser.act[key]
	if n != nil {
		n := n.(string)
		return start + int(dp-1), &n
	}
	var node string
	pos := start
	// letter:[a] {…}/letter:[b] {…}
	{
		pos3 := pos
		var node2 string
		// action
		{
			start5 := pos
			// letter:[a]
			{
				pos6 := pos
				// [a]
				if r, w := _next(parser, pos); r != 'a' {
					goto fail4
				} else {
					label0 = parser.text[pos : pos+w]
					pos += w
				}
				labels[0] = parser.text[pos6:pos]
			}
			node = func(
				start, end int, letter string) string {
				fmt.Printf("a=[%s]\n", letter)
				return string(letter)
			}(
				start5, pos, label0)
		}
		goto ok0
	fail4:
		node = node2
		pos = pos3
		// action
		{
			start8 := pos
			// letter:[b]
			{
				pos9 := pos
				// [b]
				if r, w := _next(parser, pos); r != 'b' {
					goto fail7
				} else {
					label1 = parser.text[pos : pos+w]
					pos += w
				}
				labels[1] = parser.text[pos9:pos]
			}
			node = func(
				start, end int, letter string) string {
				fmt.Printf("b=[%s]\n", letter)
				return string(letter)
			}(
				start8, pos, label1)
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
