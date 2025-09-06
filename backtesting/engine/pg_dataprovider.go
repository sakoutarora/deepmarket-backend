package engine

import (
	"errors"
	"fmt"
	"log"
	"time"

	domain "github.com/gulll/deepmarket/backtesting/domain"
	"gorm.io/gorm"
)

type PGProvider struct {
	db *gorm.DB
}

func NewPGProvider(db *gorm.DB) *PGProvider { return &PGProvider{db: db} }

func (p *PGProvider) LoadOHLCV(symbol string, tf domain.Timeframe) ([]domain.Candle, error) {
	startTime := time.Now()
	interval, found := domain.TimeframeToMinutes[tf]

	if !found {
		return nil, fmt.Errorf("timeframe %q not supported", tf)
	}

	// NOTE: consider making start/end times function arguments instead of hardcoding strings.
	rows, err := p.db.Raw(`
		WITH src AS (
		  SELECT *
		  FROM public.ohlc_data_nse_eq
		  WHERE (time::time >= time '09:15:00' AND time::time <= time '15:30:00'
		         AND ticker = ? AND "time" > ? AND "time" < ?)
		),
		annot AS (
		  SELECT
		    ticker,
		    time,
		    open, high, low, close, volume, oi,
		    (date_trunc('day', time) + interval '9 hours 15 minutes') AS session_open,
		    EXTRACT(EPOCH FROM (time - (date_trunc('day', time) + interval '9 hours 15 minutes'))) AS secs_since_open
		  FROM src
		),
		buckets AS (
		  SELECT
		    ticker,
		    time,
		    open, high, low, close, volume, oi,
		    session_open,
		    FLOOR(secs_since_open / (60.0 * ?))::int AS bucket_no
		  FROM annot
		)
		SELECT
		  -- bucket start timestamp
		  (session_open + (bucket_no * make_interval(mins => ?))) AS bucket_start,
		  (array_agg(open ORDER BY time ASC))[1] AS open,
		  MAX(high)                             AS high,
		  MIN(low)                              AS low,
		  (array_agg(close ORDER BY time DESC))[1] AS close,
		  SUM(volume)                           AS volume
		FROM buckets
		GROUP BY session_open, bucket_no
		ORDER BY bucket_start;
	`, symbol, "2025-01-01", "2025-08-01", interval, interval).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []domain.Candle

	for rows.Next() {
		var item domain.Candle
		// the query returns columns in this order:
		// bucket_start, open, high, low, close, volume
		if err := rows.Scan(
			&item.Time,
			&item.Open,
			&item.High,
			&item.Low,
			&item.Close,
			&item.Volume,
		); err != nil {
			return nil, err
		}

		candles = append(candles, item)
	}

	endTime := time.Now()
	log.Printf("⏲️ LoadOHLCV for %s took %s\n", symbol, endTime.Sub(startTime))
	return candles, nil
}

func (p *PGProvider) AlignTo(baseTF domain.Timeframe, ser Series, fromTF domain.Timeframe) (Series, error) {
	// Simplest approach: if fromTF is higher than baseTF, forward-fill each base bar within the same higher-timeframe window.
	// If fromTF is lower than baseTF, resample by last value within the base bar boundary.
	// You need timestamps for precise alignment—store them alongside price arrays in LoadOHLCV to do this correctly.
	return nil, errors.New("AlignTo not implemented")
}
