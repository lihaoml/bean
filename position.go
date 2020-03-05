package bean

import (
	"math"
	"time"
)

type Position struct {
	Con   *Contract
	Qty   float64
	Price float64
}

type Positions []Position

// Calculate the price of a contract given market parameters. Price is in RHS coin value spot
// Discounting assumes zero interest rate on LHS coin (normally BTC) which is deribit standard. Note USD rates float and are generally negative.
func (p Position) PV(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	if p.Con.IsOption() {
		/*		return p.Con.OptPrice(asof, spotPrice, futPrice, vol) * p.Qty*/
		return p.Con.OptPrice(asof, spotPrice, futPrice, vol)*p.Qty - p.Price*spotPrice*p.Qty
	} else {
		return (1.0/p.Price - 1.0/futPrice) * spotPrice * p.Qty * 10.0
	}
}

// in rhs coin spot value
func (p Position) Vega(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	return p.PV(asof, spotPrice, futPrice, vol+0.005) - p.PV(asof, spotPrice, futPrice, vol-0.005)
}

//in lhs coin spot value
func (p Position) Delta(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	deltaFiat := (p.PV(asof, spotPrice*1.005, futPrice*1.005, vol) - p.PV(asof, spotPrice*0.995, futPrice*0.995, vol)) * 100.0
	return deltaFiat / spotPrice
}

func (p Position) BucketDelta(asof time.Time, spotPrice, futPrice, vol float64) map[string]float64 {
	totdelta := (p.PV(asof, spotPrice*1.005, futPrice*1.005, vol) - p.PV(asof, spotPrice*0.995, futPrice*0.995, vol)) * 100.0
	spotDelta := (p.PV(asof, spotPrice*1.005, futPrice, vol) - p.PV(asof, spotPrice*0.995, futPrice, vol)) * 100.0

	underFuture := p.Con.UnderFuture()
	delta := make(map[string]float64)
	delta["CASH"] = spotDelta / spotPrice
	delta[underFuture.Name()] = (totdelta - spotDelta) / spotPrice

	return delta
}

//in lhs coin spot value
func (p Position) Gamma(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	gammaFiat := p.Delta(asof, spotPrice*1.005, futPrice*1.005, vol) - p.Delta(asof, spotPrice*0.995, futPrice*0.995, vol)

	return gammaFiat
}

//in rhs coin spot value
func (p Position) Theta(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	return p.PV(asof.Add(24*time.Hour), spotPrice, futPrice, vol) - p.PV(asof, spotPrice, futPrice, vol)
}

// maths stuff now

// day difference rounded.
func dayDiff(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC) // remove time information and force to utc
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)
	return int(math.Round(t2.Sub(t1).Truncate(time.Hour).Hours() / 24.0))
}

func (c Contract) ExpiryDays(now time.Time) int {
	return dayDiff(now, c.Expiry())
}

// premium expected in domestic - rhs coin value spot
func optionImpliedVol(expiryDays, deliveryDays int, strike, spot, forward, prm float64, callPut CallOrPut) (bs float64) {

	if expiryDays == 0 {
		return math.NaN()
	}

	// if premium is less than intrinsic then return zero
	floorPrm := spot / forward * forwardOptionPrice(expiryDays, strike, forward, 0.0, callPut)
	if prm <= floorPrm {
		return 0.0
	}

	// newton raphson on vega and bs
	//	guessVol := math.Sqrt(2.0*math.Pi/(float64(expiryDays)/365)) * prm / forward
	guessVol := 1.0
	for i := 0; i < 1000; i++ {
		guessPrm := spot / forward * forwardOptionPrice(expiryDays, strike, forward, guessVol, callPut)
		vega := optionVega(expiryDays, deliveryDays, strike, spot, forward, guessVol)
		vega = math.Max(vega, 0.00001*spot) // floor the vega at 1bp to avoid guesses flying off
		guessVol = guessVol - (guessPrm-prm)/(vega*100.0)
		guessVol = math.Max(guessVol, 0.0) // floor guess vol at zero
		guessVol = math.Min(guessVol, 5.0) // cap guess vol at 500%
		if math.Abs(guessPrm-prm)/forward < 0.00001 {
			return guessVol
		}
	}
	return math.NaN()
}

func dF(days int, rate float64) float64 {
	return math.Exp(-float64(days) / 365 * rate)
}

// in domestic - rhs coin forward value
func forwardOptionPrice(expiryDays int, strike, forward, vol float64, callPut CallOrPut) (prm float64) {
	if expiryDays == 0 {
		vol = 0
	}

	d1 := (math.Log(forward/strike) + (vol*vol/2.0)*(float64(expiryDays)/365)) / (vol * math.Sqrt(float64(expiryDays)/365))
	d2 := d1 - vol*math.Sqrt(float64(expiryDays)/365.0)

	if callPut == Call {
		prm = forward*cumNormDist(d1) - strike*cumNormDist(d2)
	} else {
		prm = -forward*cumNormDist(-d1) + strike*cumNormDist(-d2)
	}
	return
}

// Seems to work!
func cumNormDist(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}

func optionVega(expiryDays, deliveryDays int, strike, spot, forward, vol float64) float64 {
	//	d1 := (math.Log(forward/strike) + (vol*vol/2.0)*(float64(expiryDays)/365)) / (vol * math.Sqrt(float64(expiryDays)/365))
	//	return forward * cumNormDist(d1) * math.Sqrt(float64(expiryDays)/365.0) * dF(deliveryDays, domRate)
	return spot / forward * (forwardOptionPrice(expiryDays, strike, forward, vol+0.005, Call) - forwardOptionPrice(expiryDays, strike, forward, vol-0.005, Call))
}
