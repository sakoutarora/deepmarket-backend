// engine/planner.go
package engine

import (
	"fmt"
	"hash/fnv"

	domain "github.com/gulll/deepmarket/backtesting/domain"
)

type NodeType int

const (
	NodeSeries NodeType = iota + 1
	NodeBool
	NodeAlign // internal align op
	NodeShift // internal shift op (offset)
)

type PlanNode struct {
	ID   string
	Kind NodeType
	Op   string
	Meta map[string]any // tf, params, offset, etc
	Deps []*PlanNode
}

type Plan struct {
	Roots []*PlanNode
	Order []*PlanNode // topological order
}

// hashKey creates a stable-ish key
func hashKey(parts ...any) string {
	h := fnv.New64a()
	for _, p := range parts {
		fmt.Fprintf(h, "|%T:%v", p, p)
	}
	return fmt.Sprintf("%x", h.Sum64())
}

// Planner turns AST into a DAG with alignment & CSE
type Planner struct {
	baseTF domain.Timeframe
	cache  map[string]*PlanNode
}

func NewPlanner(baseTF domain.Timeframe) *Planner {
	return &Planner{baseTF: baseTF, cache: map[string]*PlanNode{}}
}

func (p *Planner) planExpr(x domain.ExprNode) (*PlanNode, error) {
	switch v := x.(type) {
	case domain.NumberNode:
		key := hashKey("num", v.Value, p.baseTF)
		if n, ok := p.cache[key]; ok {
			return n, nil
		}
		n := &PlanNode{ID: key, Kind: NodeSeries, Op: "const", Meta: map[string]any{"value": v.Value}}
		p.cache[key] = n
		return n, nil

	case domain.IndicatorNode:
		key := hashKey("ind", v.Name, v.Timeframe, v.Params, v.Offset)
		if n, ok := p.cache[key]; ok {
			return n, nil
		}
		pn := &PlanNode{
			ID:   key,
			Kind: NodeSeries,
			Op:   "indicator",
			Meta: map[string]any{
				"name":   v.Name,
				"tf":     v.Timeframe,
				"params": v.Params,
				"offset": v.Offset,
			},
		}
		for _, a := range v.Args {
			dep, err := p.planExpr(a)
			if err != nil {
				return nil, err
			}
			pn.Deps = append(pn.Deps, dep)
		}
		return pn, nil

	case domain.FunctionNode:
		key := hashKey("fn", v.Name, v.Params, v.Args)
		if n, ok := p.cache[key]; ok {
			return n, nil
		}

		meta := map[string]any{"name": v.Name, "params": v.Params}
		deps := []*PlanNode{}
		for _, argVal := range v.Args {
			dep, err := p.planExpr(argVal)
			if err != nil {
				return nil, err
			}
			deps = append(deps, dep)
		}
		n := &PlanNode{ID: key, Kind: NodeSeries, Op: "function", Meta: meta, Deps: deps}
		p.cache[key] = n
		return n, nil

	case domain.BinaryMathNode:
		l, err := p.planExpr(v.Left)
		if err != nil {
			return nil, err
		}
		r, err := p.planExpr(v.Right)
		if err != nil {
			return nil, err
		}
		key := hashKey("math", v.Op, l.ID, r.ID)
		if n, ok := p.cache[key]; ok {
			return n, nil
		}
		n := &PlanNode{ID: key, Kind: NodeSeries, Op: v.Op, Deps: []*PlanNode{l, r}}
		p.cache[key] = n
		return n, nil
	}
	return nil, fmt.Errorf("unknown expr node")
}

func (p *Planner) planPred(x domain.PredNode) (*PlanNode, error) {
	switch v := x.(type) {
	case domain.CompareNode:
		l, err := p.planExpr(v.Left)
		if err != nil {
			return nil, err
		}
		r, err := p.planExpr(v.Right)
		if err != nil {
			return nil, err
		}
		key := hashKey("cmp", v.Op, l.ID, r.ID)
		if n, ok := p.cache[key]; ok {
			return n, nil
		}
		n := &PlanNode{ID: key, Kind: NodeBool, Op: "cmp:" + v.Op, Deps: []*PlanNode{l, r}}
		p.cache[key] = n
		return n, nil

	case domain.LogicalNode:
		if v.Op == "NOT" {
			l, err := p.planPred(v.Lhs)
			if err != nil {
				return nil, err
			}
			key := hashKey("not", l.ID)
			if n, ok := p.cache[key]; ok {
				return n, nil
			}
			n := &PlanNode{ID: key, Kind: NodeBool, Op: "NOT", Deps: []*PlanNode{l}}
			p.cache[key] = n
			return n, nil
		}
		l, err := p.planPred(v.Lhs)
		if err != nil {
			return nil, err
		}
		r, err := p.planPred(v.Rhs)
		if err != nil {
			return nil, err
		}
		key := hashKey("logic", v.Op, l.ID, r.ID)
		if n, ok := p.cache[key]; ok {
			return n, nil
		}
		n := &PlanNode{ID: key, Kind: NodeBool, Op: v.Op, Deps: []*PlanNode{l, r}}
		p.cache[key] = n
		return n, nil
	}
	return nil, fmt.Errorf("unknown predicate node")
}

func (p *Planner) Build(root domain.PredNode) (*Plan, error) {
	r, err := p.planPred(root)
	if err != nil {
		return nil, err
	}

	// simple DFS for topo order
	seen := map[string]bool{}
	order := []*PlanNode{}
	var dfs func(n *PlanNode)
	dfs = func(n *PlanNode) {
		if seen[n.ID] {
			return
		}
		seen[n.ID] = true
		for _, d := range n.Deps {
			dfs(d)
		}
		order = append(order, n)
	}
	dfs(r)

	return &Plan{Roots: []*PlanNode{r}, Order: order}, nil
}
