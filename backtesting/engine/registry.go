// engine/registry.go
package engine

import domain "github.com/gulll/deepmarket/backtesting/domain"

type ArgSpec struct {
	Name string
	Type string // "float", "int", "expr", "series", "string", etc.
	Req  bool
}

type IndicatorSpec struct {
	Category    string
	Description string
	Params      []ArgSpec // e.g. period:int, mult:float
	// Eval returns a series for the requested timeframe
	Eval func(ctx *EvalCtx, tf domain.Timeframe, params map[string]float64,
		offset int, args ...Series) ([]float64, error)
}

type FunctionSpec struct {
	Category    string
	Description string
	Params      []ArgSpec
	// Eval: args may include nested ExprNode; resolve inside via ctx.EvalExpr if needed
	Eval func(ctx *EvalCtx, params map[string]any, args ...Series) ([]float64, error)
}

// Optional: direct Bool generators (rare but useful, e.g., "IsUpTrend")
type PredicateSpec struct {
	Category    string
	Description string
	Params      []ArgSpec
	Eval        func(ctx *EvalCtx, params map[string]any, args ...Series) ([]float64, error)
}

type Registry struct {
	Indicators map[string]IndicatorSpec
	Functions  map[string]FunctionSpec
	Predicates map[string]PredicateSpec
}

func NewRegistry() *Registry {
	return &Registry{
		Indicators: map[string]IndicatorSpec{},
		Functions:  map[string]FunctionSpec{},
		Predicates: map[string]PredicateSpec{},
	}
}
