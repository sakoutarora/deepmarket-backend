package domain

import (
	"time"
)

type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type BacktestReq struct {
	Symbol          string     `json:"symbol"`
	BaseTF          Timeframe  `json:"base_timeframe"`
	EntryConditions Condition  `json:"entry_conditions"`
	ExitConditions  *Condition `json:"exit_conditions,omitempty"`
	Direction       string     `json:"direction"` // "long" or "short"
	Quantity        int        `json:"quantity"`
	Capital         float32    `json:"capital"`

	StopLoss      float64       `json:"stop_loss,omitempty"`   // %
	TakeProfit    float64       `json:"take_profit,omitempty"` // %
	TrailingSL    float64       `json:"trailing_sl,omitempty"` // %
	Start         *string       `json:"start,omitempty"`
	End           *string       `json:"end,omitempty"`
	Intraday      *IntradayRule `json:"intraday,omitempty"`
	HoldingPeriod *int          `json:"holding_period,omitempty"`
}

type ExitCondition struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Tokens []Token `json:"tokens"`
}

type IntradayRule struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time,omitempty"` // e.g. "09:45"
	ExitTime  string `json:"exit_time"`            // "15:20"
	ReEnter   bool   `json:"re_enter"`
}

type BacktestResp struct {
	BaseTF  Timeframe              `json:"base_timeframe"`
	Results []BacktestSymbolResult `json:"results"`
	Summary BacktestSummary        `json:"summary"`
}

type BacktestSymbolResult struct {
	Symbol  string     `json:"symbol"`
	Trades  []TradeLog `json:"trades"`
	Signal  []bool     `json:"signal"`
	Entries []int      `json:"entries"`
	Exits   []int      `json:"exits"`
}

type TradeLog struct {
	Direction   string    `json:"direction"` // "long" or "short"
	EntryTime   time.Time `json:"entry_time"`
	EntryPrice  float64   `json:"entry_price"`
	ExitTime    time.Time `json:"exit_time"`
	ExitPrice   float64   `json:"exit_price"`
	ExitReason  string    `json:"exit_reason"`
	Qty         int       `json:"qty"`
	PnL         float64   `json:"pnl"`
	HoldingBars int       `json:"holding_bars"`
}

type BacktestSummary struct {
	// Profitability
	TotalTrades  int     `json:"total_trades"`
	NetProfit    float64 `json:"net_profit"`
	GrossProfit  float64 `json:"gross_profit"`
	GrossLoss    float64 `json:"gross_loss"`
	ProfitFactor float64 `json:"profit_factor"`
	Expectancy   float64 `json:"expectancy"`

	// Risk-Adjusted
	SharpeRatio  float64 `json:"sharpe_ratio"`
	SortinoRatio float64 `json:"sortino_ratio"`
	CalmarRatio  float64 `json:"calmar_ratio"`
	OmegaRatio   float64 `json:"omega_ratio"`

	// Drawdown
	MaxDrawdown    float64 `json:"max_drawdown"`
	AvgDrawdown    float64 `json:"avg_drawdown"`
	RecoveryFactor float64 `json:"recovery_factor"`
	UlcerIndex     float64 `json:"ulcer_index"`

	// Trade Quality
	WinRate         float64 `json:"win_rate"`
	AvgWin          float64 `json:"avg_win"`
	AvgLoss         float64 `json:"avg_loss"`
	RiskRewardRatio float64 `json:"risk_reward_ratio"`
	MaxConsecWins   int     `json:"max_consec_wins"`
	MaxConsecLosses int     `json:"max_consec_losses"`

	// Capital Growth
	CAGR             float64 `json:"cagr"`
	EquityVolatility float64 `json:"equity_volatility"`
	Skewness         float64 `json:"skewness"`
	Kurtosis         float64 `json:"kurtosis"`

	// Exposure
	AvgHoldBars   float64    `json:"avg_hold_bars"`
	ExposureRatio float64    `json:"exposure_ratio"`
	TurnoverRatio float64    `json:"turnover_ratio"`
	Trades        []TradeLog `json:"trades"`
}

type TradeState struct {
	TradeID       string
	HighWaterMark float64
	LowWaterMark  float64
	ActiveTSL     float64
	// future extensible fields:
	// ATRStop, BreakEvenTriggered, PartialExitDone, etc.
}

type SimulatorState struct {
	Trades  map[string]*TradeState // key = tradeID
	Equity  []float64
	Capital float64
}
