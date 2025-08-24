// domain/ast.go
package domain

type ExprNode interface {
	isExpr()
}

type (
	// leaf
	NumberNode    struct{ Value float64 }
	IndicatorNode struct {
		Name      string
		Timeframe Timeframe
		Params    map[string]float64
		Offset    int
	}
	FunctionNode struct {
		Name string
		Args map[string]any // may include nested ExprNode for "expression" args
	}

	// math binary: + - * / % ^
	BinaryMathNode struct {
		Left  ExprNode
		Op    string
		Right ExprNode
	}
)

func (NumberNode) isExpr()     {}
func (IndicatorNode) isExpr()  {}
func (FunctionNode) isExpr()   {}
func (BinaryMathNode) isExpr() {}

type PredNode interface {
	isPred()
}

type CompareNode struct {
	Left  ExprNode
	Op    string
	Right ExprNode
}

func (CompareNode) isPred() {}

type LogicalNode struct {
	Op  string // AND/OR/NOT
	Lhs PredNode
	Rhs PredNode // Rhs is nil for NOT
}

func (LogicalNode) isPred() {}
