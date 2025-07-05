package models

import (
	"time"
)

type OptionNiftyOHLC struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Symbol      string    `gorm:"column:symbol" json:"symbol"`
	ExpiryDate  time.Time `gorm:"column:expiry_date" json:"expiry_date"`
	StrikePrice float64   `gorm:"column:strike_price" json:"strike_price"`
	OptionType  string    `gorm:"column:option_type" json:"option_type"`
	Interval    string    `gorm:"column:interval" json:"interval"`
	CandleTime  time.Time `gorm:"column:candle_time" json:"candle_time"`
	Open        float64   `gorm:"column:open" json:"open"`
	High        float64   `gorm:"column:high" json:"high"`
	Low         float64   `gorm:"column:low" json:"low"`
	Close       float64   `gorm:"column:close" json:"close"`
	Volume      *int64    `gorm:"column:volume" json:"volume,omitempty"`
	OI          *int64    `gorm:"column:oi" json:"oi,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (OptionNiftyOHLC) TableName() string {
	return "option_nifty_ohlc"
}
