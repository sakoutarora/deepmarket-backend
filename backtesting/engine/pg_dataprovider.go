package engine

import (
	"errors"
	"fmt"
	"time"

	domain "github.com/gulll/deepmarket/backtesting/domain"
	"gorm.io/gorm"
)

type PGProvider struct {
	db *gorm.DB
}

func NewPGProvider(db *gorm.DB) *PGProvider { return &PGProvider{db: db} }

func (p *PGProvider) LoadOHLCV(symbol string, tf domain.Timeframe) (map[string]Series, error) {
	startTime := time.Now()
	interval, found := domain.TimeframeToMinutes[tf]

	if !found {
		return nil, fmt.Errorf("timeframe %q not supported", tf)
	}

	rows, err := p.db.Raw(`
		WITH base AS (
			SELECT
				*,
				ROW_NUMBER() OVER (ORDER BY "time") AS global_row_num
			FROM ohlc_data_nse_eq
			WHERE ticker = ? AND "time" >= ? AND "time" <= ?
		),
		grouped AS (
			SELECT *,
				(global_row_num - 1) / ? AS group_id,
				ROW_NUMBER() OVER (PARTITION BY (global_row_num - 1) / ? ORDER BY "time") AS row_in_group,
				COUNT(*) OVER (PARTITION BY (global_row_num - 1) / ?) AS rows_in_group
			FROM base
		)
		SELECT
			MIN("time") AS interval_start,
			MAX(CASE WHEN row_in_group = 1 THEN open END) AS open,
			MAX(high) AS high,
			MIN(low) AS low,
			MAX(CASE WHEN row_in_group = rows_in_group THEN close END) AS close,
			SUM(volume) AS volume
		FROM grouped
		GROUP BY group_id
		ORDER BY group_id;
	`, symbol, "2025-01-01", "2025-08-01", interval, interval, interval).Rows()

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var open, high, low, close, volume Series

	for rows.Next() {
		var item struct {
			IntervalStart time.Time
			Open          float64
			High          float64
			Low           float64
			Close         float64
			Volume        float64
		}

		if err := rows.Scan(&item.IntervalStart, &item.Open, &item.High, &item.Low, &item.Close, &item.Volume); err != nil {
			return nil, err
		}

		open = append(open, item.Open)
		high = append(high, item.High)
		low = append(low, item.Low)
		close = append(close, item.Close)
		volume = append(volume, item.Volume)
	}
	endTime := time.Now()
	fmt.Printf("LoadOHLCV for %s took %s\n", symbol, endTime.Sub(startTime))

	return map[string]Series{
		"open":   open,
		"high":   high,
		"low":    low,
		"close":  close,
		"volume": volume,
	}, nil
}

func (p *PGProvider) AlignTo(baseTF domain.Timeframe, ser Series, fromTF domain.Timeframe) (Series, error) {
	// Simplest approach: if fromTF is higher than baseTF, forward-fill each base bar within the same higher-timeframe window.
	// If fromTF is lower than baseTF, resample by last value within the base bar boundary.
	// You need timestamps for precise alignment—store them alongside price arrays in LoadOHLCV to do this correctly.
	return nil, errors.New("AlignTo not implemented")
}
