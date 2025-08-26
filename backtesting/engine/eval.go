// engine/eval.go
package engine

import (
	"errors"
	"fmt"
	"math"

	domain "github.com/gulll/deepmarket/backtesting/domain"
)

type Series []float64
type BoolSeries []bool

// provides historical OHLCV for symbol+timeframe (aligned same length per tf)
type DataProvider interface {
	LoadOHLCV(symbol string, tf domain.Timeframe) (map[string]Series, error)
	AlignTo(baseTF domain.Timeframe, series Series, fromTF domain.Timeframe) (Series, error)
}

type EvalCtx struct {
	Symbol string
	BaseTF domain.Timeframe
	Data   DataProvider
	Reg    *Registry

	// memoize computed indicator/function series by a key
	cache map[string]Series
}

func NewEvalCtx(sym string, baseTF domain.Timeframe, dp DataProvider, reg *Registry) *EvalCtx {
	return &EvalCtx{Symbol: sym, BaseTF: baseTF, Data: dp, Reg: reg, cache: map[string]Series{}}
}

func (ctx *EvalCtx) EvalExpr(n domain.ExprNode) (Series, error) {
	switch v := n.(type) {
	case domain.NumberNode:
		// broadcast a constant to base length
		base, err := ctx.Data.LoadOHLCV(ctx.Symbol, ctx.BaseTF)
		if err != nil {
			return nil, err
		}
		L := len(base["close"])
		out := make(Series, L)
		for i := range out {
			out[i] = v.Value
		}
		return out, nil

	case domain.IndicatorNode:
		spec, ok := ctx.Reg.Indicators[v.Name]
		if !ok {
			return nil, fmt.Errorf("unknown indicator %s", v.Name)
		}
		key := fmt.Sprintf("ind:%s:%s:%v:%d", v.Name, v.Timeframe, v.Params, v.Offset)
		if s, ok := ctx.cache[key]; ok {
			return s, nil
		}
		ser, err := spec.Eval(ctx, v.Timeframe, v.Params, v.Offset)
		if err != nil {
			return nil, err
		}
		if v.Timeframe != ctx.BaseTF {
			ser, err = ctx.Data.AlignTo(ctx.BaseTF, ser, v.Timeframe)
			if err != nil {
				return nil, err
			}
		}
		ctx.cache[key] = ser
		return ser, nil

	case domain.FunctionNode:
		spec, ok := ctx.Reg.Functions[v.Name]
		if !ok {
			return nil, fmt.Errorf("unknown function %s", v.Name)
		}
		key := fmt.Sprintf("fn:%s:%v", v.Name, v.Args)
		if s, ok := ctx.cache[key]; ok {
			return s, nil
		}
		ser, err := spec.Eval((*EvalCtx)(ctx), v.Args) // if you separate pkgs, adjust type
		if err != nil {
			return nil, err
		}
		ctx.cache[key] = ser
		return ser, nil

	case domain.BinaryMathNode:
		l, err := ctx.EvalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		r, err := ctx.EvalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		if len(l) != len(r) {
			return nil, errors.New("series length mismatch")
		}
		out := make(Series, len(l))
		switch v.Op {
		case "+":
			for i := range out {
				out[i] = l[i] + r[i]
			}
		case "-":
			for i := range out {
				out[i] = l[i] - r[i]
			}
		case "*":
			for i := range out {
				out[i] = l[i] * r[i]
			}
		case "/":
			for i := range out {
				if r[i] == 0 {
					out[i] = math.NaN()
				} else {
					out[i] = l[i] / r[i]
				}
			}
		case "%":
			for i := range out {
				out[i] = math.Mod(l[i], r[i])
			}
		case "^":
			for i := range out {
				out[i] = math.Pow(l[i], r[i])
			}
		default:
			return nil, fmt.Errorf("math op %s not supported", v.Op)
		}
		return out, nil
	}
	return nil, errors.New("unknown expr node")
}

func (ctx *EvalCtx) EvalPred(n domain.PredNode) (BoolSeries, error) {
	switch v := n.(type) {
	case domain.CompareNode:
		l, err := ctx.EvalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		r, err := ctx.EvalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		if len(l) != len(r) {
			return nil, errors.New("series length mismatch")
		}
		out := make(BoolSeries, len(l))
		switch v.Op {
		case ">":
			for i := range out {
				out[i] = l[i] > r[i]
			}
		case ">=":
			for i := range out {
				out[i] = l[i] >= r[i]
			}
		case "<":
			for i := range out {
				out[i] = l[i] < r[i]
			}
		case "<=":
			for i := range out {
				out[i] = l[i] <= r[i]
			}
		case "==":
			for i := range out {
				out[i] = l[i] == r[i]
			}
		case "!=":
			for i := range out {
				out[i] = l[i] != r[i]
			}
		case "crosses_above":
			for i := 1; i < len(out); i++ {
				out[i] = (l[i-1] <= r[i-1]) && (l[i] > r[i])
			}
		case "crosses_below":
			for i := 1; i < len(out); i++ {
				out[i] = (l[i-1] >= r[i-1]) && (l[i] < r[i])
			}
		default:
			return nil, fmt.Errorf("unknown compare op %s", v.Op)
		}
		return out, nil

	case domain.LogicalNode:
		if v.Op == "NOT" {
			x, err := ctx.EvalPred(v.Lhs)
			if err != nil {
				return nil, err
			}
			for i := range x {
				x[i] = !x[i]
			}
			return x, nil
		}
		l, err := ctx.EvalPred(v.Lhs)
		if err != nil {
			return nil, err
		}
		r, err := ctx.EvalPred(v.Rhs)
		if err != nil {
			return nil, err
		}
		if len(l) != len(r) {
			return nil, errors.New("boolean length mismatch")
		}
		switch v.Op {
		case "AND":
			for i := range l {
				l[i] = l[i] && r[i]
			}
		case "OR":
			for i := range l {
				l[i] = l[i] || r[i]
			}
		default:
			return nil, fmt.Errorf("logical op %s", v.Op)
		}
		return l, nil
	}
	return nil, errors.New("unknown predicate node")
}
