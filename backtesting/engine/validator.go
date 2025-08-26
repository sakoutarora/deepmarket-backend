// engine/validator.go
package engine

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	domain "github.com/gulll/deepmarket/backtesting/domain"
)

var mathPrecedence = map[string]int{
	"^": 4,
	"*": 3, "/": 3, "%": 3,
	"+": 2, "-": 2,
}

func isMath(op string) bool { _, ok := mathPrecedence[op]; return ok }

func isCompare(op string) bool {
	switch op {
	case ">", ">=", "<", "<=", "==", "!=", "crosses_above", "crosses_below":
		return true
	}
	return false
}
func isLogical(op string) bool {
	return op == "AND" || op == "OR" || op == "NOT"
}

type Parser struct {
	Reg *Registry
}

func (p *Parser) ValidateCondition(c domain.Condition) error {
	_, err := p.ParsePredicate(c.Tokens)
	return err
}

func (p *Parser) ParsePredicate(ts []domain.Token) (domain.PredNode, error) {
	// Split by logical operators of lowest precedence (left-to-right),
	// handle NOT as unary
	type chunk struct {
		negated bool
		toks    []domain.Token
	}
	var chunks []chunk
	var cur []domain.Token
	var pendingNeg bool

	flush := func() {
		if len(cur) == 0 {
			return
		}
		chunks = append(chunks, chunk{negated: pendingNeg, toks: cur})
		cur = nil
		pendingNeg = false
	}

	for i := 0; i < len(ts); i++ {
		t := ts[i]
		if t.Type == domain.TokenLogical {
			op := t.Operator
			if op == "NOT" {
				// mark next chunk negated
				if len(cur) != 0 {
					// "X NOT Y" without separator is invalid
					return nil, errors.New("NOT must appear before a comparison or group")
				}
				pendingNeg = !pendingNeg
				continue
			}
			if op == "AND" || op == "OR" {
				flush()
				// we store the operator token itself as its own chunk delimiter
				chunks = append(chunks, chunk{toks: []domain.Token{t}})
				continue
			}
			return nil, fmt.Errorf("unknown logical operator %q", op)
		}
		cur = append(cur, t)
	}
	flush()
	if len(chunks) == 0 {
		return nil, errors.New("empty condition")
	}

	// Now chunks are like: [cmpChunk] [AND] [cmpChunk] [OR] [cmpChunk] ...
	// Build a left-assoc logical tree
	var pred domain.PredNode
	var expectOp bool
	var lastOp string

	for _, ch := range chunks {
		// Operator?
		if len(ch.toks) == 1 && ch.toks[0].Type == domain.TokenLogical {
			if !expectOp {
				return nil, errors.New("unexpected logical operator")
			}
			lastOp = ch.toks[0].Operator
			expectOp = false
			continue
		}

		// Comparison chunk → must parse into a CompareNode
		cmp, err := p.parseComparison(ch.toks)
		if err != nil {
			return nil, err
		}
		var node domain.PredNode = cmp
		if ch.negated {
			node = domain.LogicalNode{Op: "NOT", Lhs: node}
		}

		if pred == nil {
			pred = node
		} else {
			if lastOp == "" {
				return nil, errors.New("missing logical operator between comparisons")
			}
			pred = domain.LogicalNode{Op: lastOp, Lhs: pred, Rhs: node}
			lastOp = ""
		}
		expectOp = true
	}
	return pred, nil
}

func (p *Parser) parseComparison(ts []domain.Token) (domain.CompareNode, error) {
	// Find the main comparison operator (there should be exactly one)
	idx := -1
	for i, t := range ts {
		if t.Type == domain.TokenOperator && isCompare(t.Operator) {
			if idx != -1 {
				return domain.CompareNode{}, errors.New("multiple comparison operators in one clause")
			}
			idx = i
		}
	}
	if idx == -1 {
		return domain.CompareNode{}, errors.New("missing comparison operator")
	}
	leftTs := ts[:idx]
	rightTs := ts[idx+1:]
	if len(leftTs) == 0 || len(rightTs) == 0 {
		return domain.CompareNode{}, errors.New("incomplete comparison")
	}

	leftExpr, err := p.parseExpr(leftTs)
	if err != nil {
		return domain.CompareNode{}, fmt.Errorf("left side: %w", err)
	}
	rightExpr, err := p.parseExpr(rightTs)
	if err != nil {
		return domain.CompareNode{}, fmt.Errorf("right side: %w", err)
	}

	return domain.CompareNode{
		Left:  leftExpr,
		Op:    ts[idx].Operator,
		Right: rightExpr,
	}, nil
}

func (p *Parser) parseExpr(ts []domain.Token) (domain.ExprNode, error) {
	// stacks
	var opst []string         // operator stack
	var out []domain.ExprNode // output (expression) stack

	emitOp := func(op string) error {
		if len(out) < 2 {
			return errors.New("malformed expression")
		}
		r := out[len(out)-1]
		out = out[:len(out)-1]
		l := out[len(out)-1]
		out = out[:len(out)-1]
		out = append(out, domain.BinaryMathNode{Left: l, Op: op, Right: r})
		return nil
	}

	i := 0
	for i < len(ts) {
		t := ts[i]
		switch t.Type {
		case domain.TokenNumber:
			out = append(out, domain.NumberNode{Value: t.Value})

		case domain.TokenIndicator:
			if _, ok := domain.AllowedTF[t.Timeframe]; !ok {
				return nil, fmt.Errorf("invalid timeframe %q", t.Timeframe)
			}
			params, err := coerceNumMap(t.Params)
			if err != nil {
				return nil, fmt.Errorf("indicator %s params: %w", t.Indicator, err)
			}
			spec, ok := p.Reg.Indicators[t.Indicator]
			if !ok {
				return nil, fmt.Errorf("unknown indicator %q", t.Indicator)
			}
			if err := checkArgs(spec.Params, params); err != nil {
				return nil, fmt.Errorf("indicator %s: %w", t.Indicator, err)
			}
			out = append(out, domain.IndicatorNode{
				Name: t.Indicator, Timeframe: t.Timeframe, Params: params, Offset: t.Offset,
			})

		case domain.TokenFunction:
			spec, ok := p.Reg.Functions[t.Function]
			if !ok {
				return nil, fmt.Errorf("unknown function %q", t.Function)
			}
			raw, ok := t.Params.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("function %s params must be object", t.Function)
			}
			if err := checkFuncArgs(spec.Params, raw); err != nil {
				return nil, fmt.Errorf("function %s: %w", t.Function, err)
			}
			out = append(out, domain.FunctionNode{Name: t.Function, Args: raw})

		case domain.TokenOperator:
			op := t.Operator
			if !isMath(op) {
				return nil, fmt.Errorf("unexpected operator %q in expression", op)
			}
			// shunting-yard precedence
			for len(opst) > 0 && mathPrecedence[opst[len(opst)-1]] >= mathPrecedence[op] {
				if err := emitOp(opst[len(opst)-1]); err != nil {
					return nil, err
				}
				opst = opst[:len(opst)-1]
			}
			opst = append(opst, op)

		default:
			return nil, fmt.Errorf("unexpected token type %s in expression", t.Type)
		}
		i++
	}

	for len(opst) > 0 {
		if err := emitOp(opst[len(opst)-1]); err != nil {
			return nil, err
		}
		opst = opst[:len(opst)-1]
	}

	if len(out) != 1 {
		return nil, errors.New("malformed expression (extra values/operators)")
	}
	return out[0], nil
}

func coerceNumMap(v any) (map[string]float64, error) {
	if v == nil {
		return map[string]float64{}, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("params must be object")
	}
	out := map[string]float64{}
	for k, val := range m {
		switch vv := val.(type) {
		case float64:
			out[k] = vv
		case int, int32, int64, uint, uint64, float32:
			out[k] = toF64(vv)
		default:
			return nil, fmt.Errorf("param %s must be number", k)
		}
	}
	return out, nil
}
func toF64(v any) float64 {
	switch t := v.(type) {
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case uint:
		return float64(t)
	case uint64:
		return float64(t)
	}
	return v.(float64)
}

func checkArgs(spec []ArgSpec, provided map[string]float64) error {
	req := map[string]bool{}
	for _, a := range spec {
		if a.Req {
			req[a.Name] = true
		}
	}
	for k := range provided {
		if !slices.ContainsFunc(spec, func(a ArgSpec) bool { return a.Name == k }) {
			return fmt.Errorf("unknown param %q", k)
		}
		delete(req, k)
	}
	if len(req) > 0 {
		return fmt.Errorf("missing params: %v", maps.Keys(req))
	}
	return nil
}
func checkFuncArgs(spec []ArgSpec, provided map[string]any) error {
	req := map[string]bool{}
	for _, a := range spec {
		if a.Req {
			req[a.Name] = true
		}
	}
	for k, v := range provided {
		if !slices.ContainsFunc(spec, func(a ArgSpec) bool { return a.Name == k }) {
			return fmt.Errorf("unknown param %q", k)
		}
		// very light type check
		// You can extend to allow nested expression structure here
		switch v.(type) {
		case float64, int, int32, int64, uint, uint64, float32:
		case map[string]any, []any:
			// acceptable if expression is nested
		default:
			// allow 0/1 sentinel etc.
		}
		delete(req, k)
	}
	if len(req) > 0 {
		return fmt.Errorf("missing params: %v", maps.Keys(req))
	}
	return nil
}
