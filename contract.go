package bean

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type callOrPut string

const (
	Call callOrPut = "C"
	Put  callOrPut = "P"
)

const USDiscountRate = 0.02

type Contract struct {
	name string
	qty  float64
}

func ContractFromName(name string, quantity float64) Contract {
	return Contract{name, quantity}
}

func ContractsFromNames(names []string, quantities []float64) (c []Contract) {
	c = make([]Contract, 0, 0)
	for i := range names {
		if quantities != nil {
			c = append(c, Contract{names[i], quantities[i]})
		} else {
			c = append(c, Contract{names[i], 0.0})
		}
	}
	return
}

func ContractFromDets(c Coin, d time.Time, strike float64, cp callOrPut, quantity float64) Contract {
	var cptext string
	if cp == Call {
		cptext = "C"
	} else {
		cptext = "P"
	}

	return Contract{fmt.Sprintf("%s-%s-%4.0f-%s", c.Format(), d.Format("2JAN06"), strike, cptext), quantity}
}

func (c Contract) Name() string {
	return c.name
}

func (c Contract) Quantity() float64 {
	return c.qty
}

func (c Contract) Expiry() (dt time.Time) {
	dt, _ = time.Parse("02Jan06", strings.ToTitle(strings.Split(c.name, "-")[1]))
	dt = time.Date(dt.Year(), dt.Month(), dt.Day(), 9, 0, 0, 0, time.UTC) // 9am london expiry
	return
}

func (c Contract) Strike() (st float64) {
	st, _ = strconv.ParseFloat(strings.Split(c.name, "-")[2], 64)
	return
}

func (c Contract) CallPut() callOrPut {
	switch strings.Split(c.name, "-")[3] {
	case "C":
		return Call
	case "P":
		return Put
	}
	panic("Need C OR P")
}

func (c Contract) Underlying() Pair {
	switch strings.Split(c.name, "-")[0] {
	case "BTC":
		return Pair{BTC, USDT}
	}
	panic("Only accept BTC underlying")
}

// Calculate the implied vol of a contract given its premium
func (c Contract) ImpVol(asof time.Time, spotPrice, futPrice, domRate, optionPrice float64) float64 {
	expiry := c.Expiry()
	strike := c.Strike()
	cp := c.CallPut()
	expiryDays := dayDiff(asof, expiry)
	deliveryDays := expiryDays // temp

	return optionImpliedVol(expiryDays, deliveryDays, strike, futPrice, domRate, optionPrice*spotPrice, cp)
}

// in fiat
func (c Contract) PV(asof time.Time, spotPrice, futPrice, domRate, vol float64) float64 {
	expiry := c.Expiry()
	strike := c.Strike()
	cp := c.CallPut()
	expiryDays := dayDiff(asof, expiry)
	deliveryDays := expiryDays // temp
	return c.Quantity() * forwardOptionPrice(expiryDays, strike, futPrice, vol, cp) * dF(deliveryDays, domRate)
}

// in fiat
func (c Contract) Vega(asof time.Time, spotPrice, futPrice, vol, domRate float64) float64 {
	return c.PV(asof, spotPrice, futPrice, domRate, vol+0.005) - c.PV(asof, spotPrice, futPrice, domRate, vol-0.005)
}

//in coin
func (c Contract) Delta(asof time.Time, spotPrice, futPrice, vol, domRate float64) float64 {
	deltaFiat := (c.PV(asof, spotPrice*1.005, futPrice*1.005, domRate, vol) - c.PV(asof, spotPrice*0.995, futPrice*0.995, domRate, vol)) * 100.0
	return deltaFiat / spotPrice
}

// maths stuff now

// day difference rounded.
func dayDiff(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC) // remove time information and force to utc
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)
	return int(math.Round(t2.Sub(t1).Truncate(time.Hour).Hours() / 24.0))
}

func optionImpliedVol(expiryDays, deliveryDays int, strike, forward, domRate, prm float64, callPut callOrPut) (bs float64) {
	// newton raphson on vega and bs
	//	guessVol := math.Sqrt(2.0*math.Pi/(float64(expiryDays)/365)) * prm / forward
	guessVol := 0.80
	for i := 0; i < 1000; i++ {
		guessPrm := dF(deliveryDays, domRate) * forwardOptionPrice(expiryDays, strike, forward, guessVol, callPut)
		vega := optionVega(expiryDays, deliveryDays, strike, forward, guessVol, domRate)
		guessVol = guessVol - (guessPrm-prm)/(vega*100.0)
		if guessPrm/prm < 1.00001 && guessPrm/prm > 0.99999 {
			return guessVol
		}
	}
	return math.NaN()
}

func dF(days int, rate float64) float64 {
	return math.Exp(-float64(days) / 365 * rate)
}

// in fiat
func forwardOptionPrice(expiryDays int, strike, forward, vol float64, callPut callOrPut) (prm float64) {
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
	return 0.5 * math.Erfc(-x/math.Sqrt(2))
}

func optionVega(expiryDays, deliveryDays int, strike, forward, vol, domRate float64) float64 {
	//	d1 := (math.Log(forward/strike) + (vol*vol/2.0)*(float64(expiryDays)/365)) / (vol * math.Sqrt(float64(expiryDays)/365))
	//	return forward * cumNormDist(d1) * math.Sqrt(float64(expiryDays)/365.0) * dF(deliveryDays, domRate)
	return dF(deliveryDays, domRate) * (forwardOptionPrice(expiryDays, strike, forward, vol+0.005, Call) - forwardOptionPrice(expiryDays, strike, forward, vol-0.005, Call))
}

func optionDelta(expiryDays, deliveryDays int, callPut callOrPut, strike, forward, vol, domRate float64) float64 {
	//	d1 := (math.Log(forward/strike) + (vol*vol/2.0)*(float64(expiryDays)/365)) / (vol * math.Sqrt(float64(expiryDays)/365))
	//	return forward * cumNormDist(d1) * math.Sqrt(float64(expiryDays)/365.0) * dF(deliveryDays, domRate)
	return dF(deliveryDays, domRate) * (forwardOptionPrice(expiryDays, strike, forward, vol, callPut) - forwardOptionPrice(expiryDays, strike, forward, vol, callPut))
}
