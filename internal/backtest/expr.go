package backtest

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------
// 信号表达式引擎：自研轻量表达式解析与求值（无第三方依赖）。
//
// 支持语法：
//   - 数值：123、1.5
//   - 指标引用：ma_fast、rsi14（取当前 bar 位置的值）
//   - 字段引用：open/high/low/close/volume/turnover（当前 bar）
//   - 算术：+ - * / （如 close - ma_slow）
//   - 比较：== != > >= < <= （NaN 参与比较恒为 false）
//   - 逻辑：AND OR NOT（大小写不敏感）
//   - 内置函数：
//       cross_up(a, b)     a 上穿 b（a[i]>b[i] 且 a[i-1]<=b[i-1]）
//       cross_down(a, b)   a 下穿 b
//       ref(id, n)         n 根 bar 前的值（n>=0）
//       highest(id, n)     最近 n 根（含当前）最大值
//       lowest(id, n)      最近 n 根（含当前）最小值
//       abs(x)             绝对值
//   - 括号
//
// 顶层表达式必须求值为布尔（供买卖信号使用）。
// ---------------------------------------------------------------------

// exprNode AST 节点。
type exprNode interface{}

type numNode struct{ value float64 }
type identNode struct{ name string }
type binaryNode struct {
	op    string
	left  exprNode
	right exprNode
}
type unaryNode struct {
	op   string
	node exprNode
}
type callNode struct {
	name string
	args []exprNode
}

// EvalContext 求值上下文（回测引擎逐 bar 提供）。
type EvalContext struct {
	At         int                  // 当前 bar 索引
	Fields     map[string][]float64 // 字段序列（open/high/low/close/volume/turnover）
	Indicators map[string][]float64 // 指标序列（id → 序列）
}

func (c *EvalContext) series(name string) ([]float64, bool) {
	if s, ok := c.Indicators[name]; ok {
		return s, true
	}
	if s, ok := c.Fields[name]; ok {
		return s, true
	}
	return nil, false
}

func (c *EvalContext) at(name string) (float64, bool) {
	s, ok := c.series(name)
	if !ok {
		return 0, false
	}
	if c.At < 0 || c.At >= len(s) {
		return 0, false
	}
	return s[c.At], true
}

// ---------------------------------------------------------------------
// 词法分析
// ---------------------------------------------------------------------

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokNumber
	tokIdent
	tokOp // == != > >= < <= + - * /
	tokLParen
	tokRParen
	tokComma
	tokAnd
	tokOr
	tokNot
)

type token struct {
	kind  tokenKind
	text  string
	value float64
}

type lexer struct {
	src  string
	pos  int
	toks []token
}

func lexExpr(src string) ([]token, error) {
	l := &lexer{src: src}
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			l.pos++
		case ch == '(':
			l.toks = append(l.toks, token{kind: tokLParen, text: "("})
			l.pos++
		case ch == ')':
			l.toks = append(l.toks, token{kind: tokRParen, text: ")"})
			l.pos++
		case ch == ',':
			l.toks = append(l.toks, token{kind: tokComma, text: ","})
			l.pos++
		case ch == '+' || ch == '-' || ch == '*' || ch == '/':
			l.toks = append(l.toks, token{kind: tokOp, text: string(ch)})
			l.pos++
		case ch == '=' || ch == '!' || ch == '>' || ch == '<':
			start := l.pos
			l.pos++
			if l.pos < len(l.src) && l.src[l.pos] == '=' {
				l.pos++
			}
			l.toks = append(l.toks, token{kind: tokOp, text: l.src[start:l.pos]})
		case unicode.IsDigit(rune(ch)) || (ch == '.' && l.pos+1 < len(l.src) && unicode.IsDigit(rune(l.src[l.pos+1]))):
			start := l.pos
			for l.pos < len(l.src) && (unicode.IsDigit(rune(l.src[l.pos])) || l.src[l.pos] == '.') {
				l.pos++
			}
			num, err := strconv.ParseFloat(l.src[start:l.pos], 64)
			if err != nil {
				return nil, fmt.Errorf("非法数字 %q", l.src[start:l.pos])
			}
			l.toks = append(l.toks, token{kind: tokNumber, text: l.src[start:l.pos], value: num})
		case unicode.IsLetter(rune(ch)) || ch == '_':
			start := l.pos
			for l.pos < len(l.src) && (unicode.IsLetter(rune(l.src[l.pos])) || unicode.IsDigit(rune(l.src[l.pos])) || l.src[l.pos] == '_') {
				l.pos++
			}
			word := l.src[start:l.pos]
			switch strings.ToUpper(word) {
			case "AND":
				l.toks = append(l.toks, token{kind: tokAnd, text: word})
			case "OR":
				l.toks = append(l.toks, token{kind: tokOr, text: word})
			case "NOT":
				l.toks = append(l.toks, token{kind: tokNot, text: word})
			default:
				l.toks = append(l.toks, token{kind: tokIdent, text: word})
			}
		default:
			return nil, fmt.Errorf("非法字符 %q（位置 %d）", string(ch), l.pos)
		}
	}
	l.toks = append(l.toks, token{kind: tokEOF})
	return l.toks, nil
}

// ---------------------------------------------------------------------
// 语法分析（递归下降）
// ---------------------------------------------------------------------

type parser struct {
	toks []token
	pos  int
}

func parseExpr(src string) (exprNode, error) {
	toks, err := lexExpr(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	node, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tokEOF {
		return nil, fmt.Errorf("表达式尾部存在多余内容 %q", p.peek().text)
	}
	return node, nil
}

func (p *parser) peek() token { return p.toks[p.pos] }
func (p *parser) next() token { t := p.toks[p.pos]; p.pos++; return t }

func (p *parser) parseOr() (exprNode, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokOr {
		op := p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op.text, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (exprNode, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokAnd {
		op := p.next()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op.text, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseNot() (exprNode, error) {
	if p.peek().kind == tokNot {
		p.next()
		node, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &unaryNode{op: "NOT", node: node}, nil
	}
	return p.parseComparison()
}

func (p *parser) parseComparison() (exprNode, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	if p.peek().kind == tokOp {
		op := p.peek().text
		if op == "==" || op == "!=" || op == ">" || op == ">=" || op == "<" || op == "<=" {
			p.next()
			right, err := p.parseAdditive()
			if err != nil {
				return nil, err
			}
			return &binaryNode{op: op, left: left, right: right}, nil
		}
	}
	return left, nil
}

func (p *parser) parseAdditive() (exprNode, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokOp && (p.peek().text == "+" || p.peek().text == "-") {
		op := p.next()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op.text, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseMultiplicative() (exprNode, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokOp && (p.peek().text == "*" || p.peek().text == "/") {
		op := p.next()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op.text, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parsePrimary() (exprNode, error) {
	t := p.next()
	switch t.kind {
	case tokNumber:
		return &numNode{value: t.value}, nil
	case tokIdent:
		if p.peek().kind == tokLParen {
			p.next() // 吃掉 (
			call := &callNode{name: t.text}
			if p.peek().kind != tokRParen {
				for {
					arg, err := p.parseOr()
					if err != nil {
						return nil, err
					}
					call.args = append(call.args, arg)
					if p.peek().kind == tokComma {
						p.next()
						continue
					}
					break
				}
			}
			if p.peek().kind != tokRParen {
				return nil, fmt.Errorf("函数 %s 缺少右括号", t.text)
			}
			p.next()
			return call, nil
		}
		return &identNode{name: t.text}, nil
	case tokLParen:
		node, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.peek().kind != tokRParen {
			return nil, fmt.Errorf("缺少右括号")
		}
		p.next()
		return node, nil
	}
	return nil, fmt.Errorf("意外的 token %q", t.text)
}

// ---------------------------------------------------------------------
// 求值
// ---------------------------------------------------------------------

// evalNum 数值求值。
func evalNum(node exprNode, ctx *EvalContext) (float64, error) {
	switch n := node.(type) {
	case *numNode:
		return n.value, nil
	case *identNode:
		v, ok := ctx.at(n.name)
		if !ok {
			return 0, fmt.Errorf("未知标识符 %q（未在指标或字段中定义）", n.name)
		}
		return v, nil
	case *binaryNode:
		switch n.op {
		case "+":
			l, err := evalNum(n.left, ctx)
			if err != nil {
				return 0, err
			}
			r, err := evalNum(n.right, ctx)
			if err != nil {
				return 0, err
			}
			return l + r, nil
		case "-":
			l, err := evalNum(n.left, ctx)
			if err != nil {
				return 0, err
			}
			r, err := evalNum(n.right, ctx)
			if err != nil {
				return 0, err
			}
			return l - r, nil
		case "*":
			l, err := evalNum(n.left, ctx)
			if err != nil {
				return 0, err
			}
			r, err := evalNum(n.right, ctx)
			if err != nil {
				return 0, err
			}
			return l * r, nil
		case "/":
			l, err := evalNum(n.left, ctx)
			if err != nil {
				return 0, err
			}
			r, err := evalNum(n.right, ctx)
			if err != nil {
				return 0, err
			}
			if r == 0 {
				return math.NaN(), nil
			}
			return l / r, nil
		}
		return 0, fmt.Errorf("运算符 %s 不适用于数值", n.op)
	case *unaryNode:
		if n.op == "-" {
			v, err := evalNum(n.node, ctx)
			if err != nil {
				return 0, err
			}
			return -v, nil
		}
		return 0, fmt.Errorf("运算符 %s 不适用于数值", n.op)
	case *callNode:
		return evalCallNum(n, ctx)
	}
	return 0, fmt.Errorf("表达式无法求值为数值")
}

// evalCallNum 数值型内置函数。
func evalCallNum(call *callNode, ctx *EvalContext) (float64, error) {
	name := strings.ToLower(call.name)
	switch name {
	case "abs":
		if len(call.args) != 1 {
			return 0, fmt.Errorf("abs 需要 1 个参数")
		}
		v, err := evalNum(call.args[0], ctx)
		if err != nil {
			return 0, err
		}
		return math.Abs(v), nil
	case "ref":
		if len(call.args) != 2 {
			return 0, fmt.Errorf("ref(id, n) 需要 2 个参数")
		}
		id, err := evalIdentName(call.args[0])
		if err != nil {
			return 0, err
		}
		n, err := evalIntArg(call.args[1], ctx)
		if err != nil {
			return 0, err
		}
		s, ok := ctx.series(id)
		if !ok {
			return 0, fmt.Errorf("ref: 未知标识符 %q", id)
		}
		idx := ctx.At - n
		if idx < 0 || idx >= len(s) {
			return math.NaN(), nil
		}
		return s[idx], nil
	case "highest", "lowest":
		if len(call.args) != 2 {
			return 0, fmt.Errorf("%s(id, n) 需要 2 个参数", name)
		}
		id, err := evalIdentName(call.args[0])
		if err != nil {
			return 0, err
		}
		n, err := evalIntArg(call.args[1], ctx)
		if err != nil {
			return 0, err
		}
		s, ok := ctx.series(id)
		if !ok {
			return 0, fmt.Errorf("%s: 未知标识符 %q", name, id)
		}
		if n < 1 {
			return math.NaN(), nil
		}
		start := ctx.At - n + 1
		if start < 0 {
			start = 0
		}
		if start > ctx.At {
			return math.NaN(), nil
		}
		result := s[start]
		if name == "highest" {
			for i := start + 1; i <= ctx.At; i++ {
				if s[i] > result {
					result = s[i]
				}
			}
		} else {
			for i := start + 1; i <= ctx.At; i++ {
				if s[i] < result {
					result = s[i]
				}
			}
		}
		return result, nil
	}
	return 0, fmt.Errorf("未知函数 %q", call.name)
}

// evalIdentName 参数为标识符时取名称。
func evalIdentName(node exprNode) (string, error) {
	if n, ok := node.(*identNode); ok {
		return n.name, nil
	}
	return "", fmt.Errorf("函数参数需要标识符")
}

func evalIntArg(node exprNode, ctx *EvalContext) (int, error) {
	v, err := evalNum(node, ctx)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// evalBool 布尔求值。
func evalBool(node exprNode, ctx *EvalContext) (bool, error) {
	switch n := node.(type) {
	case *binaryNode:
		switch n.op {
		case "AND":
			l, err := evalBool(n.left, ctx)
			if err != nil {
				return false, err
			}
			if !l {
				return false, nil
			}
			return evalBool(n.right, ctx)
		case "OR":
			l, err := evalBool(n.left, ctx)
			if err != nil {
				return false, err
			}
			if l {
				return true, nil
			}
			return evalBool(n.right, ctx)
		case "==", "!=", ">", ">=", "<", "<=":
			l, err := evalNum(n.left, ctx)
			if err != nil {
				return false, err
			}
			r, err := evalNum(n.right, ctx)
			if err != nil {
				return false, err
			}
			if math.IsNaN(l) || math.IsNaN(r) {
				return false, nil
			}
			switch n.op {
			case "==":
				return l == r, nil
			case "!=":
				return l != r, nil
			case ">":
				return l > r, nil
			case ">=":
				return l >= r, nil
			case "<":
				return l < r, nil
			case "<=":
				return l <= r, nil
			}
		case "+", "-", "*", "/":
			// 数值表达式被当作布尔时（非零即真）
			v, err := evalNum(n, ctx)
			if err != nil {
				return false, err
			}
			return !math.IsNaN(v) && v != 0, nil
		}
	case *unaryNode:
		if n.op == "NOT" {
			v, err := evalBool(n.node, ctx)
			if err != nil {
				return false, err
			}
			return !v, nil
		}
	case *numNode:
		return n.value != 0, nil
	case *identNode:
		v, ok := ctx.at(n.name)
		if !ok {
			return false, fmt.Errorf("未知标识符 %q", n.name)
		}
		return !math.IsNaN(v) && v != 0, nil
	case *callNode:
		return evalCallBool(n, ctx)
	}
	return false, fmt.Errorf("表达式无法求值为布尔")
}

// evalCallBool 布尔型内置函数（cross_up / cross_down）。
func evalCallBool(call *callNode, ctx *EvalContext) (bool, error) {
	name := strings.ToLower(call.name)
	switch name {
	case "cross_up", "cross_down":
		if len(call.args) != 2 {
			return false, fmt.Errorf("%s(a, b) 需要 2 个参数", name)
		}
		idA, err := evalIdentName(call.args[0])
		if err != nil {
			return false, err
		}
		idB, err := evalIdentName(call.args[1])
		if err != nil {
			return false, err
		}
		sa, okA := ctx.series(idA)
		sb, okB := ctx.series(idB)
		if !okA || !okB {
			return false, fmt.Errorf("%s: 未知标识符", name)
		}
		i := ctx.At
		if i <= 0 || i >= len(sa) || i >= len(sb) {
			return false, nil
		}
		curA, prevA, curB, prevB := sa[i], sa[i-1], sb[i], sb[i-1]
		if math.IsNaN(curA) || math.IsNaN(prevA) || math.IsNaN(curB) || math.IsNaN(prevB) {
			return false, nil
		}
		if name == "cross_up" {
			return prevA <= prevB && curA > curB, nil
		}
		return prevA >= prevB && curA < curB, nil
	}
	return false, fmt.Errorf("未知函数 %q", call.name)
}

// knownFunctions 内置函数及其参数个数（-1 表示可变）。
var knownFunctions = map[string][2]int{
	"cross_up":   {2, 2},
	"cross_down": {2, 2},
	"ref":        {2, 2},
	"highest":    {2, 2},
	"lowest":     {2, 2},
	"abs":        {1, 1},
}

// validateAST 校验 AST：函数名与参数个数合法性（编译期拦截，避免运行期才报错）。
func validateAST(node exprNode) error {
	switch n := node.(type) {
	case *binaryNode:
		if err := validateAST(n.left); err != nil {
			return err
		}
		return validateAST(n.right)
	case *unaryNode:
		return validateAST(n.node)
	case *callNode:
		bounds, ok := knownFunctions[strings.ToLower(n.name)]
		if !ok {
			return fmt.Errorf("未知函数 %q（支持: cross_up/cross_down/ref/highest/lowest/abs）", n.name)
		}
		if len(n.args) < bounds[0] || len(n.args) > bounds[1] {
			return fmt.Errorf("函数 %s 需要 %d 个参数，实际 %d", n.name, bounds[0], len(n.args))
		}
		for _, arg := range n.args {
			if err := validateAST(arg); err != nil {
				return err
			}
		}
	}
	return nil
}

// CompileExpr 编译表达式（校验语法与函数），返回可直接求值的闭包。
func CompileExpr(src string) (func(ctx *EvalContext) (bool, error), error) {
	node, err := parseExpr(src)
	if err != nil {
		return nil, fmt.Errorf("信号表达式语法错误: %w", err)
	}
	if err := validateAST(node); err != nil {
		return nil, fmt.Errorf("信号表达式错误: %w", err)
	}
	return func(ctx *EvalContext) (bool, error) {
		return evalBool(node, ctx)
	}, nil
}
