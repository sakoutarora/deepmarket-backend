// domain/ast.go
package domain

// ---- Expression & Predicate AST ----
type ExprNode interface{ exprNode() }
type PredNode interface{ predNode() }

type NumberNode struct {
	Value float64
}

func (NumberNode) exprNode() {}

type IndicatorNode struct {
	Name      string
	Timeframe Timeframe
	Params    map[string]float64
	Offset    int
	Args      []ExprNode
}

func (IndicatorNode) exprNode() {}

type FunctionNode struct {
	Name string
	// Args can contain scalars (float64, int, string), nested ExprNode, or even lists
	Params map[string]any
	Args   []ExprNode
}

func (FunctionNode) exprNode() {}

type BinaryMathNode struct {
	Left  ExprNode
	Op    string // "+", "-", "*", "/", "%", "^"
	Right ExprNode
}

func (BinaryMathNode) exprNode() {}

type CompareNode struct {
	Left  ExprNode
	Op    string // ">", ">=", "<", "<=", "==", "!=", "crosses_above", "crosses_below"
	Right ExprNode
}

func (CompareNode) predNode() {}

type LogicalNode struct {
	Op  string // "AND", "OR", "NOT"
	Lhs PredNode
	Rhs PredNode // optional for NOT
}

func (LogicalNode) predNode() {}
