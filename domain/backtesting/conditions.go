// domain/conditions.go
package domain

type Timeframe string

var AllowedTF = map[Timeframe]struct{}{
	"1m": {}, "3m": {}, "5m": {}, "15m": {}, "30m": {}, "1H": {}, "2H": {}, "4H": {}, "1D": {}, "1W": {}, "1M": {},
}

var TimeframeToMinutes = map[Timeframe]int{
	"1m": 1, "3m": 3, "5m": 5, "15m": 15, "30m": 30, "1H": 60, "2H": 120, "4H": 240, "1D": 1440, "1W": 10080, "1M": 43200,
}

type TokenType string

const (
	TokenIndicator TokenType = "indicator"
	TokenOperator  TokenType = "operator" // math or comparison
	TokenNumber    TokenType = "number"
	TokenFunction  TokenType = "function"
	TokenLogical   TokenType = "logical" // AND/OR/NOT between clauses
)

type Operator string

type Token struct {
	ID        string    `json:"id"`
	Type      TokenType `json:"type"`
	Timeframe Timeframe `json:"timeframe,omitempty"`
	Indicator string    `json:"indicator,omitempty"`
	Params    any       `json:"params,omitempty"` // map[string]any expected
	Offset    int       `json:"offset,omitempty"`
	Operator  string    `json:"operator,omitempty"` // +, >, AND, crosses_above, etc.
	Value     float64   `json:"value,omitempty"`    // for number
	Function  string    `json:"function,omitempty"`
}

type Condition struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Tokens []Token `json:"tokens"`
}
