// engine/runtime.go
package engine

import (
	"fmt"
	"log"
	"math"
	"slices"
	"strings"

	domain "github.com/gulll/deepmarket/backtesting/domain"
)

type Runtime struct {
	ctx *EvalCtx
}

func NewRuntime(ctx *EvalCtx) *Runtime { return &Runtime{ctx: ctx} }

// ExecPlan evaluates the plan and returns the root BoolSeries.
func (rt *Runtime) ExecPlan(pl *Plan) (BoolSeries, error) {
	for _, n := range pl.Order {
		switch n.Kind {
		case NodeSeries, NodeAlign, NodeShift:
			if _, ok := rt.ctx.cache[n.ID]; ok {
				continue
			}
			ser, err := rt.execSeriesNode(n)
			if err != nil {
				return nil, err
			}
			rt.ctx.cache[n.ID] = ser

		case NodeBool:
			if _, ok := rt.ctx.bcache[n.ID]; ok {
				continue
			}
			bs, err := rt.execBoolNode(n)
			if err != nil {
				return nil, err
			}
			rt.ctx.bcache[n.ID] = bs
		}
	}
	// Root is Bool
	root := pl.Roots[0]
	if bs, ok := rt.ctx.bcache[root.ID]; ok {
		return bs, nil
	}
	return nil, fmt.Errorf("root bool not found")
}

func (rt *Runtime) loadSeries(n *PlanNode, idx int) (Series, error) {
	d := n.Deps[idx]
	if s, ok := rt.ctx.cache[d.ID]; ok {
		return s, nil
	}
	// Should have been created earlier via topo order
	ser, err := rt.execSeriesNode(d)
	if err != nil {
		return nil, err
	}
	rt.ctx.cache[d.ID] = ser
	return ser, nil
}

func (rt *Runtime) loadBool(n *PlanNode, idx int) (BoolSeries, error) {
	d := n.Deps[idx]
	if s, ok := rt.ctx.bcache[d.ID]; ok {
		return s, nil
	}
	bs, err := rt.execBoolNode(d)
	if err != nil {
		return nil, err
	}
	rt.ctx.bcache[d.ID] = bs
	return bs, nil
}

// func (rt *Runtime) execSeriesNode(n *PlanNode) (Series, error) {
// 	switch n.Op {
// 	case "const":
// 		value := n.Meta["value"].(float64)
// 		// TODO: check is this close key in rt.ctx.cache is correct or not
// 		L := len(rt.ctx.cache["close"])
// 		out := make(Series, L)
// 		for i := range out {
// 			out[i] = value
// 		}
// 		return out, nil

// 	case "indicator":
// 		name := n.Meta["name"].(string)
// 		tf := n.Meta["tf"].(domain.Timeframe)
// 		params := n.Meta["params"].(map[string]float64)
// 		offset := int(n.Meta["offset"].(int))
// 		spec, ok := rt.ctx.Reg.Indicators[name]
// 		if !ok {
// 			return nil, fmt.Errorf("unknown indicator %s", name)
// 		}
// 		ser, err := spec.Eval(rt.ctx, tf, params, offset)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if tf != rt.ctx.BaseTF {
// 			ser, err = rt.ctx.Data.AlignTo(rt.ctx.BaseTF, ser, tf)
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		// internal cache key already the plan node id
// 		return ser, nil

// 	case "function":
// 		args := n.Meta["args"].(map[string]any)
// 		// Resolve any ExprNode arguments by looking for their planned deps
// 		// The deps order aligns with the ExprNode args found in planner, but to remain robust,
// 		// allow both: pull series from deps and also pass ExprNodes if user specs expect them.
// 		resolved := map[string]any{}
// 		for k, v := range args {
// 			if _, isExpr := v.(domain.ExprNode); isExpr {
// 				// find a dep whose key contains k (best-effort; alternative: store mapping in Meta)
// 				// fallback: use first unresolved dep
// 				var ser Series
// 				switch {
// 				case len(n.Deps) == 1:
// 					ser = rt.ctx.cache[n.Deps[0].ID]
// 				default:
// 					// try name match in ID
// 					found := false
// 					for _, d := range n.Deps {
// 						if strings.Contains(d.ID, k) {
// 							ser = rt.ctx.cache[d.ID]
// 							found = true
// 							break
// 						}
// 					}
// 					if !found && len(n.Deps) > 0 {
// 						ser = rt.ctx.cache[n.Deps[0].ID]
// 					}
// 				}
// 				resolved[k] = ser
// 			} else {
// 				resolved[k] = v
// 			}
// 		}
// 		spec, ok := rt.ctx.Reg.Functions[n.Meta["name"].(string)]
// 		if !ok {
// 			return nil, fmt.Errorf("unknown function %s", n.Meta["name"])
// 		}
// 		return spec.Eval(rt.ctx, resolved)

// 	case "align":
// 		fromTF := n.Meta["fromTF"].(domain.Timeframe)
// 		_ = fromTF // info only; dep[0] series is already computed at fromTF by indicator/function
// 		src := rt.ctx.cache[n.Deps[0].ID]
// 		return rt.ctx.Data.AlignTo(rt.ctx.BaseTF, src, fromTF)

// 	case "+", "-", "*", "/", "%", "^":
// 		l, err := rt.loadSeries(n, 0)
// 		if err != nil {
// 			return nil, err
// 		}
// 		r, err := rt.loadSeries(n, 1)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if len(l) != len(r) {
// 			return nil, fmt.Errorf("series length mismatch")
// 		}
// 		out := make(Series, len(l))
// 		switch n.Op {
// 		case "+":
// 			for i := range out {
// 				out[i] = l[i] + r[i]
// 			}
// 		case "-":
// 			for i := range out {
// 				out[i] = l[i] - r[i]
// 			}
// 		case "*":
// 			for i := range out {
// 				out[i] = l[i] * r[i]
// 			}
// 		case "/":
// 			for i := range out {
// 				if r[i] == 0 {
// 					out[i] = math.NaN()
// 				} else {
// 					out[i] = l[i] / r[i]
// 				}
// 			}
// 		case "%":
// 			for i := range out {
// 				out[i] = math.Mod(l[i], r[i])
// 			}
// 		case "^":
// 			for i := range out {
// 				out[i] = math.Pow(l[i], r[i])
// 			}
// 		}
// 		return out, nil
// 	}
// 	return nil, fmt.Errorf("unknown series op %s", n.Op)
// }

func (rt *Runtime) execSeriesNode(n *PlanNode) (Series, error) {
	switch n.Op {
	case "const":
		value := n.Meta["value"].(float64)
		L := len(rt.ctx.cache["close"])
		out := make(Series, L)
		for i := range out {
			out[i] = value
		}
		return out, nil

	case "indicator":
		name := n.Meta["name"].(string)
		tf := n.Meta["tf"].(domain.Timeframe)
		params := n.Meta["params"].(map[string]float64)
		offset := n.Meta["offset"].(int)

		spec, ok := rt.ctx.Reg.Indicators[name]
		if !ok {
			return nil, fmt.Errorf("unknown indicator %s", name)
		}

		// ✅ collect arg series from deps
		var argSeries []Series
		for i, _ := range n.Deps {
			ser, err := rt.loadSeries(n, i)
			if err != nil {
				return nil, err
			}
			argSeries = append(argSeries, ser)
		}

		ser, err := spec.Eval(rt.ctx, tf, params, offset, argSeries...)
		if err != nil {
			return nil, err
		}

		if tf != rt.ctx.BaseTF {
			ser, err = rt.ctx.Data.AlignTo(rt.ctx.BaseTF, ser, tf)
			if err != nil {
				return nil, err
			}
		}
		return ser, nil

	case "function":
		params := n.Meta["params"].(map[string]any)

		spec, ok := rt.ctx.Reg.Functions[n.Meta["name"].(string)]
		if !ok {
			return nil, fmt.Errorf("unknown function %s", n.Meta["name"])
		}

		// ✅ collect arg series from deps
		var argSeries []Series
		log.Println("☑️ adding function", n.Meta["name"].(string))
		for i, _ := range n.Deps {
			log.Println("👉", n.Deps[i].ID, n.Deps[i].Kind)
			ser, err := rt.loadSeries(n, i)
			if err != nil {
				return nil, err
			}
			argSeries = append(argSeries, ser)
		}

		return spec.Eval(rt.ctx, params, argSeries...)

	case "align":
		fromTF := n.Meta["fromTF"].(domain.Timeframe)
		src := rt.ctx.cache[n.Deps[0].ID]
		return rt.ctx.Data.AlignTo(rt.ctx.BaseTF, src, fromTF)

	case "+", "-", "*", "/", "%", "^":
		l, err := rt.loadSeries(n, 0)
		if err != nil {
			return nil, err
		}
		r, err := rt.loadSeries(n, 1)
		if err != nil {
			return nil, err
		}
		if len(l) != len(r) {
			return nil, fmt.Errorf("series length mismatch")
		}
		out := make(Series, len(l))
		switch n.Op {
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
		}
		return out, nil
	}
	return nil, fmt.Errorf("unknown series op %s", n.Op)
}

func (rt *Runtime) execBoolNode(n *PlanNode) (BoolSeries, error) {
	switch n.Op {
	case "NOT":
		l, err := rt.loadBool(n, 0)
		if err != nil {
			return nil, err
		}
		out := slices.Clone(l)
		for i := range out {
			out[i] = !out[i]
		}
		return out, nil

	case "AND", "OR":
		l, err := rt.loadBool(n, 0)
		if err != nil {
			return nil, err
		}
		r, err := rt.loadBool(n, 1)
		if err != nil {
			return nil, err
		}
		if len(l) != len(r) {
			return nil, fmt.Errorf("boolean length mismatch")
		}
		switch n.Op {
		case "AND":
			for i := range l {
				l[i] = l[i] && r[i]
			}
		case "OR":
			for i := range l {
				l[i] = l[i] || r[i]
			}
		}
		return l, nil

	default:
		// cmp: prefix "cmp:"
		if strings.HasPrefix(n.Op, "cmp:") {
			op := strings.TrimPrefix(n.Op, "cmp:")
			l, err := rt.loadSeries(n, 0)
			if err != nil {
				return nil, err
			}
			r, err := rt.loadSeries(n, 1)
			if err != nil {
				return nil, err
			}
			if len(l) != len(r) {
				return nil, fmt.Errorf("series length mismatch")
			}

			switch op {
			case ">":
				return rt.ctx.sanitizeBoolFromPair(l, r, func(i int) bool { return l[i] > r[i] }), nil
			case ">=":
				return rt.ctx.sanitizeBoolFromPair(l, r, func(i int) bool { return l[i] >= r[i] }), nil
			case "<":
				return rt.ctx.sanitizeBoolFromPair(l, r, func(i int) bool { return l[i] < r[i] }), nil
			case "<=":
				return rt.ctx.sanitizeBoolFromPair(l, r, func(i int) bool { return l[i] <= r[i] }), nil
			case "==":
				return rt.ctx.sanitizeBoolFromPair(l, r, func(i int) bool { return l[i] == r[i] }), nil
			case "!=":
				return rt.ctx.sanitizeBoolFromPair(l, r, func(i int) bool { return l[i] != r[i] }), nil
			case "crosses_above":
				out := make(BoolSeries, len(l))
				for i := 1; i < len(out); i++ {
					if math.IsNaN(l[i-1]) || math.IsNaN(r[i-1]) || math.IsNaN(l[i]) || math.IsNaN(r[i]) {
						out[i] = !rt.ctx.Policy.NaNIsFalse
					} else {
						out[i] = (l[i-1] <= r[i-1]) && (l[i] > r[i])
					}
				}
				return out, nil
			case "crosses_below":
				out := make(BoolSeries, len(l))
				for i := 1; i < len(out); i++ {
					if math.IsNaN(l[i-1]) || math.IsNaN(r[i-1]) || math.IsNaN(l[i]) || math.IsNaN(r[i]) {
						out[i] = !rt.ctx.Policy.NaNIsFalse
					} else {
						out[i] = (l[i-1] >= r[i-1]) && (l[i] < r[i])
					}
				}
				return out, nil
			}
		}
	}
	return nil, fmt.Errorf("unknown bool op %s", n.Op)
}
