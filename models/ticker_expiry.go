package models

import (
	"time"
)

type TickerExpiry struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UExToken       *int      `json:"u_ex_token"` // optional
	TickerID       int       `json:"ticker_id"`
	ExpiryDate     time.Time `json:"expiry_date"`
	InstrumentType string    `json:"instrument_type"`
	OptionType     string    `json:"option_type"`
	StrikePrice    float64   `json:"strike_price"`
	CreatedAt      time.Time `json:"created_at"`
	Exchange       string    `json:"exchange"`
	LotSize        float64   `json:"lot_size"`
}
