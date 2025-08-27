package engine

import (
	"math"
	"time"
)

// Candle represents an OHLCV bar.
type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// --- Helpers ---

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// sliceLastN returns a copy of the last n candles (n<=len).
func sliceLastN(bars []Candle, n int) []Candle {
	if n <= 0 {
		return []Candle{}
	}
	if n > len(bars) {
		n = len(bars)
	}
	out := make([]Candle, n)
	copy(out, bars[len(bars)-n:])
	return out
}

// SMA Simple Moving Average of closes for period p.
func SMA(values []float64, p int) []float64 {
	if p <= 0 || len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	var sum float64
	for i := range values {
		sum += values[i]
		if i >= p {
			sum -= values[i-p]
		}
		if i+1 >= p {
			out[i] = sum / float64(p)
		} else {
			out[i] = math.NaN()
		}
	}
	return out
}

// EMA Exponential Moving Average of values for period p.
func EMA(values []float64, p int) []float64 {
	if p <= 0 || len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	k := 2.0 / (float64(p) + 1.0)
	var prev float64
	for i, v := range values {
		if i == 0 {
			prev = v
			out[i] = v
			continue
		}
		prev = v*k + prev*(1-k)
		out[i] = prev
	}
	return out
}

// WMA Weighted Moving Average (linear weights 1..p).
func WMA(values []float64, p int) []float64 {
	if p <= 0 || len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	den := float64(p*(p+1)) / 2.0
	for i := range values {
		if i+1 < p {
			out[i] = math.NaN()
			continue
		}
		var num float64
		for w := 1; w <= p; w++ {
			num += float64(w) * values[i-(p-w)]
		}
		out[i] = num / den
	}
	return out
}

// DEMA Double Exponential Moving Average
func DEMA(values []float64, p int) []float64 {
	e := EMA(values, p)
	e2 := EMA(e, p)
	out := make([]float64, len(values))
	for i := range out {
		out[i] = 2*e[i] - e2[i]
	}
	return out
}

// TEMA Triple Exponential Moving Average
func TEMA(values []float64, p int) []float64 {
	e1 := EMA(values, p)
	e2 := EMA(e1, p)
	e3 := EMA(e2, p)
	out := make([]float64, len(values))
	for i := range out {
		out[i] = 3*(e1[i]-e2[i]) + e3[i]
	}
	return out
}

// TMA Triangular Moving Average (SMA of SMA with adjusted period)
func TMA(values []float64, p int) []float64 {
	if p <= 0 {
		return nil
	}
	m := (p + 1) / 2
	return SMA(SMA(values, m), m)
}

// HMA Hull Moving Average: WMA(2*WMA(n/2) - WMA(n), sqrt(n))
func HMA(values []float64, p int) []float64 {
	if p <= 1 {
		return values
	}
	half := WMA(values, p/2)
	full := WMA(values, p)
	diff := make([]float64, len(values))
	for i := range diff {
		diff[i] = 2*half[i] - full[i]
	}
	return WMA(diff, int(math.Round(math.Sqrt(float64(p)))))
}

// StdDev rolling standard deviation for period p.
func StdDev(values []float64, p int) []float64 {
	if p <= 0 || len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	mean := SMA(values, p)
	for i := range values {
		if i+1 < p {
			out[i] = math.NaN()
			continue
		}
		var s float64
		for j := i - p + 1; j <= i; j++ {
			d := values[j] - mean[i]
			s += d * d
		}
		out[i] = math.Sqrt(s / float64(p))
	}
	return out
}

// TrueRange for each candle.
func TrueRange(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		hl := bars[i].High - bars[i].Low
		if i == 0 {
			out[i] = hl
			continue
		}
		pc := bars[i-1].Close
		out[i] = max(hl, max(abs(bars[i].High-pc), abs(bars[i].Low-pc)))
	}
	return out
}

// ATR Average True Range.
func ATR(bars []Candle, p int) []float64 {
	tr := TrueRange(bars)
	return EMA(tr, p)
}

// NATR Normalized ATR = 100 * ATR / Close
func NATR(bars []Candle, p int) []float64 {
	atr := ATR(bars, p)
	out := make([]float64, len(bars))
	for i := range bars {
		out[i] = 100 * atr[i] / bars[i].Close
	}
	return out
}

// BollingerBands returns upper, middle, lower bands for period p, k std.
func BollingerBands(values []float64, p int, k float64) (ubb, mbb, lbb []float64) {
	mbb = SMA(values, p)
	sd := StdDev(values, p)
	ubb = make([]float64, len(values))
	lbb = make([]float64, len(values))
	for i := range values {
		ubb[i] = mbb[i] + k*sd[i]
		lbb[i] = mbb[i] - k*sd[i]
	}
	return
}

// KeltnerChannels using EMA and ATR: middle=EMA, upper=EMA+k*ATR, lower=EMA-k*ATR
func KeltnerChannels(bars []Candle, emaPeriod int, atrPeriod int, k float64) (upper, mid, lower []float64) {
	closes := ExtractCloses(bars)
	mid = EMA(closes, emaPeriod)
	atr := ATR(bars, atrPeriod)
	upper = make([]float64, len(bars))
	lower = make([]float64, len(bars))
	for i := range bars {
		upper[i] = mid[i] + k*atr[i]
		lower[i] = mid[i] - k*atr[i]
	}
	return
}

// VWAP cumulative from start of slice.
func VWAP(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	var pvCum, vCum float64
	for i := range bars {
		typ := (bars[i].High + bars[i].Low + bars[i].Close) / 3
		pvCum += typ * bars[i].Volume
		vCum += bars[i].Volume
		if vCum == 0 {
			out[i] = math.NaN()
		} else {
			out[i] = pvCum / vCum
		}
	}
	return out
}

// VWAPMA - moving average of VWAP.
func VWAPMA(bars []Candle, p int) []float64 { return SMA(VWAP(bars), p) }

// Extract close values.
func ExtractCloses(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		out[i] = bars[i].Close
	}
	return out
}
func ExtractHighs(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		out[i] = bars[i].High
	}
	return out
}
func ExtractLows(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		out[i] = bars[i].Low
	}
	return out
}

// ROC Rate of Change over p periods: (Close/Close[p])-1
func ROC(values []float64, p int) []float64 {
	out := make([]float64, len(values))
	for i := range values {
		if i < p {
			out[i] = math.NaN()
			continue
		}
		out[i] = (values[i]/values[i-p] - 1.0) * 100
	}
	return out
}

// Momentum (Close - Close[p])
func Momentum(values []float64, p int) []float64 {
	out := make([]float64, len(values))
	for i := range values {
		if i < p {
			out[i] = math.NaN()
			continue
		}
		out[i] = values[i] - values[i-p]
	}
	return out
}

// RSI Relative Strength Index (Wilder's).
func RSI(values []float64, p int) []float64 {
	if p <= 0 || len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	var prevGain, prevLoss float64
	for i := 1; i < len(values); i++ {
		chg := values[i] - values[i-1]
		gain := math.Max(chg, 0)
		loss := math.Max(-chg, 0)
		if i <= p {
			prevGain += gain
			prevLoss += loss
			if i == p {
				avgGain := prevGain / float64(p)
				avgLoss := prevLoss / float64(p)
				rs := avgGain / max(avgLoss, 1e-12)
				out[i] = 100 - 100/(1+rs)
			}
			continue
		}
		avgGain := (prevGain*(float64(p-1)) + gain) / float64(p)
		avgLoss := (prevLoss*(float64(p-1)) + loss) / float64(p)
		prevGain, prevLoss = avgGain, avgLoss
		rs := avgGain / max(avgLoss, 1e-12)
		out[i] = 100 - 100/(1+rs)
	}
	return out
}

// Stochastic %K and %D (SMA of %K)
func Stochastic(bars []Candle, kPeriod, dPeriod int) (kVals, dVals []float64) {
	kVals = make([]float64, len(bars))
	for i := range bars {
		if i+1 < kPeriod {
			kVals[i] = math.NaN()
			continue
		}
		lowest := bars[i].Low
		highest := bars[i].High
		for j := i - kPeriod + 1; j <= i; j++ {
			if bars[j].Low < lowest {
				lowest = bars[j].Low
			}
			if bars[j].High > highest {
				highest = bars[j].High
			}
		}
		den := highest - lowest
		if den == 0 {
			kVals[i] = 0
		} else {
			kVals[i] = (bars[i].Close - lowest) / den * 100
		}
	}
	dVals = SMA(kVals, dPeriod)
	return
}

// Williams %R
func WilliamsR(bars []Candle, p int) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		if i+1 < p {
			out[i] = math.NaN()
			continue
		}
		h := bars[i].High
		l := bars[i].Low
		for j := i - p + 1; j <= i; j++ {
			if bars[j].High > h {
				h = bars[j].High
			}
			if bars[j].Low < l {
				l = bars[j].Low
			}
		}
		den := h - l
		if den == 0 {
			out[i] = 0
		} else {
			out[i] = -100 * (h - bars[i].Close) / den
		}
	}
	return out
}

// MACD returns macd line, signal, histogram. Fast=12, Slow=26, Signal=9 typical.
func MACD(values []float64, fast, slow, signal int) (macd, sig, hist []float64) {
	fastEMA := EMA(values, fast)
	slowEMA := EMA(values, slow)
	macd = make([]float64, len(values))
	for i := range values {
		macd[i] = fastEMA[i] - slowEMA[i]
	}
	sig = EMA(macd, signal)
	hist = make([]float64, len(values))
	for i := range values {
		hist[i] = macd[i] - sig[i]
	}
	return
}

// OBV On-Balance Volume
func OBV(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	var cum float64
	for i := range bars {
		if i == 0 {
			out[i] = 0
			continue
		}
		if bars[i].Close > bars[i-1].Close {
			cum += bars[i].Volume
		}
		if bars[i].Close < bars[i-1].Close {
			cum -= bars[i].Volume
		}
		out[i] = cum
	}
	return out
}

// MFI Money Flow Index
func MFI(bars []Candle, p int) []float64 {
	if len(bars) == 0 {
		return nil
	}
	typ := make([]float64, len(bars))
	raw := make([]float64, len(bars))
	for i := range bars {
		typ[i] = (bars[i].High + bars[i].Low + bars[i].Close) / 3
		raw[i] = typ[i] * bars[i].Volume
	}
	out := make([]float64, len(bars))
	for i := range bars {
		if i < p {
			out[i] = math.NaN()
			continue
		}
		var pos, neg float64
		for j := i - p + 1; j <= i; j++ {
			delta := typ[j] - typ[j-1]
			if delta >= 0 {
				pos += raw[j]
			} else {
				neg += raw[j]
			}
		}
		ratio := pos / max(neg, 1e-12)
		out[i] = 100 - 100/(1+ratio)
	}
	return out
}

// CCI Commodity Channel Index
func CCI(bars []Candle, p int) []float64 {
	typ := make([]float64, len(bars))
	for i := range bars {
		typ[i] = (bars[i].High + bars[i].Low + bars[i].Close) / 3
	}
	ma := SMA(typ, p)
	sd := StdDev(typ, p)
	out := make([]float64, len(bars))
	for i := range bars {
		out[i] = (typ[i] - ma[i]) / (0.015 * sd[i])
	}
	return out
}

// ADX with +DI and -DI (Wilder's smoothing)
func ADX(bars []Candle, p int) (adx, plusDI, minusDI []float64) {
	tr := TrueRange(bars)
	plusDM := make([]float64, len(bars))
	minusDM := make([]float64, len(bars))
	for i := 1; i < len(bars); i++ {
		upMove := bars[i].High - bars[i-1].High
		downMove := bars[i-1].Low - bars[i].Low
		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		} else {
			plusDM[i] = 0
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		} else {
			minusDM[i] = 0
		}
	}
	sTR := EMA(tr, p) // Wilder's smoothing approximation
	sPlusDM := EMA(plusDM, p)
	sMinusDM := EMA(minusDM, p)
	plusDI = make([]float64, len(bars))
	minusDI = make([]float64, len(bars))
	dx := make([]float64, len(bars))
	for i := range bars {
		if sTR[i] == 0 {
			continue
		}
		plusDI[i] = 100 * (sPlusDM[i] / sTR[i])
		minusDI[i] = 100 * (sMinusDM[i] / sTR[i])
		diff := abs(plusDI[i] - minusDI[i])
		sum := plusDI[i] + minusDI[i]
		if sum == 0 {
			dx[i] = 0
		} else {
			dx[i] = 100 * (diff / sum)
		}
	}
	adx = EMA(dx, p)
	return
}

// Donchian Channels: highest high / lowest low over p
func Donchian(bars []Candle, p int) (upper, lower, middle []float64) {
	upper = make([]float64, len(bars))
	lower = make([]float64, len(bars))
	middle = make([]float64, len(bars))
	for i := range bars {
		if i+1 < p {
			upper[i], lower[i], middle[i] = math.NaN(), math.NaN(), math.NaN()
			continue
		}
		h := bars[i].High
		l := bars[i].Low
		for j := i - p + 1; j <= i; j++ {
			if bars[j].High > h {
				h = bars[j].High
			}
			if bars[j].Low < l {
				l = bars[j].Low
			}
		}
		upper[i] = h
		lower[i] = l
		middle[i] = (h + l) / 2
	}
	return
}

// Supertrend: returns supertrend line and direction (1 uptrend, -1 downtrend)
func Supertrend(bars []Candle, atrPeriod int, multiplier float64) (trend []float64, dir []int) {
	atr := ATR(bars, atrPeriod)
	basicUpper := make([]float64, len(bars))
	basicLower := make([]float64, len(bars))
	finalUpper := make([]float64, len(bars))
	finalLower := make([]float64, len(bars))
	trend = make([]float64, len(bars))
	dir = make([]int, len(bars))
	for i := range bars {
		mid := (bars[i].High + bars[i].Low) / 2
		basicUpper[i] = mid + multiplier*atr[i]
		basicLower[i] = mid - multiplier*atr[i]
		if i == 0 {
			finalUpper[i] = basicUpper[i]
			finalLower[i] = basicLower[i]
			trend[i] = finalLower[i]
			dir[i] = 1
			continue
		}
		if basicUpper[i] < finalUpper[i-1] || bars[i-1].Close > finalUpper[i-1] {
			finalUpper[i] = basicUpper[i]
		} else {
			finalUpper[i] = finalUpper[i-1]
		}
		if basicLower[i] > finalLower[i-1] || bars[i-1].Close < finalLower[i-1] {
			finalLower[i] = basicLower[i]
		} else {
			finalLower[i] = finalLower[i-1]
		}
		if trend[i-1] == finalUpper[i-1] {
			if bars[i].Close > finalUpper[i] {
				trend[i] = finalLower[i]
				dir[i] = 1
			} else {
				trend[i] = finalUpper[i]
				dir[i] = -1
			}
		} else {
			if bars[i].Close < finalLower[i] {
				trend[i] = finalUpper[i]
				dir[i] = -1
			} else {
				trend[i] = finalLower[i]
				dir[i] = 1
			}
		}
	}
	return
}

// Ichimoku Cloud components: Tenkan (conversion), Kijun (base), Senkou A/B (leading), Chikou (lagging)
func Ichimoku(bars []Candle, convPeriod, basePeriod, spanBPeriod, displacement int) (tenkan, kijun, senkouA, senkouB, chikou []float64) {
	n := len(bars)
	tenkan = make([]float64, n)
	kijun = make([]float64, n)
	senkouA = make([]float64, n)
	senkouB = make([]float64, n)
	chikou = make([]float64, n)
	for i := range bars {
		// Tenkan
		if i+1 >= convPeriod {
			h := bars[i].High
			l := bars[i].Low
			for j := i - convPeriod + 1; j <= i; j++ {
				if bars[j].High > h {
					h = bars[j].High
				}
				if bars[j].Low < l {
					l = bars[j].Low
				}
			}
			tenkan[i] = (h + l) / 2
		} else {
			tenkan[i] = math.NaN()
		}
		// Kijun
		if i+1 >= basePeriod {
			h := bars[i].High
			l := bars[i].Low
			for j := i - basePeriod + 1; j <= i; j++ {
				if bars[j].High > h {
					h = bars[j].High
				}
				if bars[j].Low < l {
					l = bars[j].Low
				}
			}
			kijun[i] = (h + l) / 2
		} else {
			kijun[i] = math.NaN()
		}
		// Senkou A/B (placed forward by displacement; caller can handle indexing)
		if !math.IsNaN(tenkan[i]) && !math.IsNaN(kijun[i]) {
			senkouA[i] = (tenkan[i] + kijun[i]) / 2
		} else {
			senkouA[i] = math.NaN()
		}
		if i+1 >= spanBPeriod {
			h := bars[i].High
			l := bars[i].Low
			for j := i - spanBPeriod + 1; j <= i; j++ {
				if bars[j].High > h {
					h = bars[j].High
				}
				if bars[j].Low < l {
					l = bars[j].Low
				}
			}
			senkouB[i] = (h + l) / 2
		} else {
			senkouB[i] = math.NaN()
		}
		// Chikou lagging span
		chikou[i] = bars[i].Close // user can shift back by displacement when plotting
	}
	return
}

// PSAR Parabolic SAR (Wilder). af=acceleration factor, inc=increment, max=maximum
func PSAR(bars []Candle, af, inc, afMax float64) []float64 {
	n := len(bars)
	out := make([]float64, n)
	if n == 0 {
		return out
	}
	// initialize
	long := true
	afCurr := af
	ep := bars[0].High // extreme point
	sar := bars[0].Low
	out[0] = sar
	for i := 1; i < n; i++ {
		prev := i - 1
		sar = sar + afCurr*(ep-sar)
		if long {
			if sar > min(bars[prev].Low, bars[i].Low) {
				sar = min(bars[prev].Low, bars[i].Low)
			}
			if bars[i].High > ep {
				ep = bars[i].High
				afCurr = min(afCurr+inc, afMax)
			}
			if bars[i].Low < sar { // switch to short
				long = false
				sar = ep
				ep = bars[i].Low
				afCurr = af
			}
		} else { // short trend
			if sar < max(bars[prev].High, bars[i].High) {
				sar = max(bars[prev].High, bars[i].High)
			}
			if bars[i].Low < ep {
				ep = bars[i].Low
				afCurr = min(afCurr+inc, afMax)
			}
			if bars[i].High > sar { // switch to long
				long = true
				sar = ep
				ep = bars[i].High
				afCurr = af
			}
		}
		out[i] = sar
	}
	return out
}

// Pivot Points (Classic): returns PP, R1-3, S1-3
func PivotClassic(prevHigh, prevLow, prevClose float64) (pp, r1, r2, r3, s1, s2, s3 float64) {
	pp = (prevHigh + prevLow + prevClose) / 3
	r1 = 2*pp - prevLow
	s1 = 2*pp - prevHigh
	r2 = pp + (prevHigh - prevLow)
	s2 = pp - (prevHigh - prevLow)
	r3 = prevHigh + 2*(pp-prevLow)
	s3 = prevLow - 2*(prevHigh-pp)
	return
}

// Aroon Oscillator
func Aroon(bars []Candle, p int) (up, down, osc []float64) {
	n := len(bars)
	up = make([]float64, n)
	down = make([]float64, n)
	osc = make([]float64, n)
	for i := 0; i < n; i++ {
		if i+1 < p {
			up[i], down[i], osc[i] = math.NaN(), math.NaN(), math.NaN()
			continue
		}
		hIdx := i
		lIdx := i
		for j := i - p + 1; j <= i; j++ {
			if bars[j].High >= bars[hIdx].High {
				hIdx = j
			}
			if bars[j].Low <= bars[lIdx].Low {
				lIdx = j
			}
		}
		u := 100 * float64(p-1-(i-hIdx)) / float64(p-1)
		d := 100 * float64(p-1-(i-lIdx)) / float64(p-1)
		up[i], down[i] = u, d
		osc[i] = u - d
	}
	return
}

// Vortex Indicator
func Vortex(bars []Candle, p int) (viPlus, viMinus []float64) {
	n := len(bars)
	viPlus = make([]float64, n)
	viMinus = make([]float64, n)
	tr := TrueRange(bars)
	trSum := EMA(tr, p)
	plus := make([]float64, n)
	minus := make([]float64, n)
	for i := 1; i < n; i++ {
		plus[i] = abs(bars[i].High - bars[i-1].Low)
		minus[i] = abs(bars[i].Low - bars[i-1].High)
	}
	plusSum := EMA(plus, p)
	minusSum := EMA(minus, p)
	for i := 0; i < n; i++ {
		if trSum[i] == 0 {
			viPlus[i], viMinus[i] = 0, 0
		} else {
			viPlus[i] = plusSum[i] / trSum[i]
			viMinus[i] = minusSum[i] / trSum[i]
		}
	}
	return
}

// Choppiness Index
func ChoppinessIndex(bars []Candle, p int) []float64 {
	tr := TrueRange(bars)
	out := make([]float64, len(bars))
	for i := range bars {
		if i+1 < p {
			out[i] = math.NaN()
			continue
		}
		var trSum float64
		highest := bars[i].High
		lowest := bars[i].Low
		for j := i - p + 1; j <= i; j++ {
			trSum += tr[j]
			if bars[j].High > highest {
				highest = bars[j].High
			}
			if bars[j].Low < lowest {
				lowest = bars[j].Low
			}
		}
		den := highest - lowest
		if den == 0 {
			out[i] = 0
		} else {
			out[i] = 100 * math.Log10(trSum/den) / math.Log10(float64(p))
		}
	}
	return out
}

// TRIX Triple Exponential Average Percent Rate of Change
func TRIX(values []float64, p int) []float64 {
	e1 := EMA(values, p)
	e2 := EMA(e1, p)
	e3 := EMA(e2, p)
	roc := make([]float64, len(values))
	for i := 1; i < len(values); i++ {
		if e3[i-1] == 0 {
			roc[i] = 0
		} else {
			roc[i] = (e3[i]/e3[i-1] - 1) * 100
		}
	}
	return roc
}

// Awesome Oscillator: SMA(5) of median price - SMA(34) of median price
func AwesomeOscillator(bars []Candle) []float64 {
	n := len(bars)
	med := make([]float64, n)
	for i := range bars {
		med[i] = (bars[i].High + bars[i].Low) / 2
	}
	fast := SMA(med, 5)
	slow := SMA(med, 34)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = fast[i] - slow[i]
	}
	return out
}

// Chaikin Money Flow (CMF)
func ChaikinMoneyFlow(bars []Candle, p int) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		if i+1 < p {
			out[i] = math.NaN()
			continue
		}
		var mfvSum, volSum float64
		for j := i - p + 1; j <= i; j++ {
			den := (bars[j].High - bars[j].Low)
			mfMult := 0.0
			if den != 0 {
				mfMult = ((bars[j].Close - bars[j].Low) - (bars[j].High - bars[j].Close)) / den
			}
			mfvSum += mfMult * bars[j].Volume
			volSum += bars[j].Volume
		}
		if volSum == 0 {
			out[i] = 0
		} else {
			out[i] = mfvSum / volSum
		}
	}
	return out
}

// Twiggs Money Flow (approx)
func TwiggsMoneyFlow(bars []Candle, p int) []float64 {
	out := make([]float64, len(bars))
	emaVol := EMA(ExtractVolumes(bars), p)
	typ := make([]float64, len(bars))
	for i := range bars {
		typ[i] = (bars[i].High + bars[i].Low + bars[i].Close) / 3
	}
	emaTyp := EMA(typ, p)
	for i := range bars {
		den := bars[i].High - bars[i].Low
		mf := 0.0
		if den != 0 {
			mf = (bars[i].Close - bars[i].Low - (bars[i].High - bars[i].Close)) / den
		}
		if emaVol[i] == 0 {
			out[i] = 0
		} else {
			out[i] = (mf * bars[i].Volume) / emaVol[i] * emaTyp[i] / max(emaTyp[i], 1e-9)
		}
	}
	return out
}

// ExtractVolumes helper
func ExtractVolumes(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		out[i] = bars[i].Volume
	}
	return out
}

// Median Price and its MA
func MedianPrice(bars []Candle) []float64 {
	out := make([]float64, len(bars))
	for i := range bars {
		out[i] = (bars[i].High + bars[i].Low) / 2
	}
	return out
}
func MedianPriceMA(bars []Candle, p int) []float64 { return SMA(MedianPrice(bars), p) }

// Pivot variants (Camarilla, CPR) left as TODOs with signatures
func CamarillaPivots(prevHigh, prevLow, prevClose float64) (h1, h2, h3, h4, l1, l2, l3, l4 float64) {
	// TODO: implement exact Camarilla formula used in your platform if required.
	return
}

// Standard Deviation on VWAP
func StdDevOnVWAP(bars []Candle, p int) []float64 { return StdDev(VWAP(bars), p) }

// Opening Range helpers
func OpeningRange(bars []Candle, window int) (high, low, open, close []float64) {
	n := len(bars)
	high = make([]float64, n)
	low = make([]float64, n)
	open = make([]float64, n)
	close = make([]float64, n)
	for i := 0; i < n; i++ {
		if i+1 < window {
			high[i], low[i], open[i], close[i] = math.NaN(), math.NaN(), math.NaN(), math.NaN()
			continue
		}
		h := bars[i-window+1].High
		l := bars[i-window+1].Low
		for j := i - window + 1; j <= i; j++ {
			if bars[j].High > h {
				h = bars[j].High
			}
			if bars[j].Low < l {
				l = bars[j].Low
			}
		}
		high[i], low[i], open[i], close[i] = h, l, bars[i-window+1].Open, bars[i].Close
	}
	return
}

// Convenience: PrevN close/hlc/volume; offset<0 for previous bars.
func PrevN(bars []Candle, field string, offset int) float64 {
	idx := len(bars) + offset
	if idx < 0 || idx >= len(bars) {
		return math.NaN()
	}
	switch field {
	case "open":
		return bars[idx].Open
	case "high":
		return bars[idx].High
	case "low":
		return bars[idx].Low
	case "close":
		return bars[idx].Close
	case "volume":
		return bars[idx].Volume
	default:
		return math.NaN()
	}
}

// Placeholder signatures for additional indicators mentioned by Streak docs:
// Alligator, ATR Trailing Stop, Fractal Chaos Bands, Linear Regression Forecast,
// McGinley Dynamic, VWAP MA (provided), Positive/Negative Volume Index, TVI,
// Volume Oscillator, Standard Deviation (provided), Choppiness MA, Standard Deviation,
// RSI MA, CCI MA, ADX MA, Stochastic Momentum Index, Stochastic RSI, TII, KST,
// Elder Force Index, Intraday Momentum Index, Moving Average Deviation, Ultimate Oscillator,
// True Strength Index, Swing Index, Bollinger %B, Bandwidth, Central Pivot Range, CPR, etc.

// For each of the above, add implementations as needed using canonical formulas.
