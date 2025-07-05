package models

import "time"

type Ticker struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	TickerSymbol string    `gorm:"column:ticker_symbol" json:"symbol"`
	FullName     string    `gorm:"column:full_name" json:"full_name"`
	Sector       string    `gorm:"column:sector" json:"sector"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// Tell GORM explicit table name
func (Ticker) TableName() string {
	return "tickers"
}
