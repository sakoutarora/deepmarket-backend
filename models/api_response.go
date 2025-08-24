package models

// "log"

type TickerResponse struct {
	ID            int    `json:"id"`
	TradingSymbol string `json:"trading_symbol"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
