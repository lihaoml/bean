package bean

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type CallOrPut string

const (
	Call CallOrPut = "C"
	Put  CallOrPut = "P"
	NA   CallOrPut = "N"
)

type Contract struct {
	name       string
	isOption   bool
	underlying Pair
	expiry     time.Time
	delivery   time.Time
	strike     float64
	callPut    CallOrPut
	perp       bool
}

//type Contracts map[Contract]float64

type Position struct {
	Con   Contract
	Qty   float64
	Price float64
}

type Positions []Position

func ContractFromName(name string) (Contract, error) {
	var expiry time.Time
	var callPut CallOrPut
	var underlying Pair
	var strike float64
	var err error
	var perp bool

	st := strings.Split(name, "-")
	if len(st) != 4 && len(st) != 2 {
		err = errors.New("not a good contract formation")
		return Contract{}, err
	}

	switch st[0] {
	case "BTC":
		underlying = Pair{BTC, USDT}
	default:
		err = errors.New("do not recognise coin")
		return Contract{}, err
	}

	if st[1] == "PERPETUAL" {
		perp = true
		expiry = time.Now()
	} else {
		perp = false
		dt, err := time.Parse("2Jan06", strings.ToTitle(st[1]))
		if err != nil {
			return Contract{}, err
		}
		expiry = time.Date(dt.Year(), dt.Month(), dt.Day(), 9, 0, 0, 0, time.UTC) // 9am london expiry
	}

	if len(st) == 2 {
		return Contract{
			isOption:   false,
			underlying: underlying,
			expiry:     expiry,
			delivery:   expiry,
			callPut:    NA,
			strike:     0.0,
			perp:       perp}, nil
	}

	strike, err = strconv.ParseFloat(st[2], 64)
	if err != nil {
		return Contract{}, err
	}

	switch st[3] {
	case "C":
		callPut = Call
	case "P":
		callPut = Put
	default:
		return Contract{}, errors.New("Need C OR P")

	}
	return Contract{
		isOption:   true,
		underlying: underlying,
		expiry:     expiry,
		delivery:   expiry,
		callPut:    callPut,
		strike:     strike}, nil

}

func ContractFromPartialName(partialName string) (Contract, error) {
	const example = "\nDon't understand contract\nExample JUN or 3500 or MAR-4000-C or BTC-3000-P"
	sts := strings.Split(partialName, "-")
	defaultExpiry, _ := time.Parse("02Jan06", "28Jun19")
	c := Contract{
		isOption:   false,
		underlying: Pair{BTC, USDT},
		expiry:     defaultExpiry,
		delivery:   defaultExpiry,
		callPut:    Call,
		strike:     5000}

	for _, s := range sts {
		switch strings.ToUpper(s) {
		case "PERP":
			c.perp = true
			c.expiry = time.Now()
			c.isOption = false
			continue
		case "MAR":
			c.expiry, _ = time.Parse("2Jan06", "29Mar19")
			c.delivery = c.expiry
			continue
		case "APR":
			c.expiry, _ = time.Parse("2Jan06", "26Apr19")
			c.delivery = c.expiry
			continue
		case "JUN":
			c.expiry, _ = time.Parse("2Jan06", "28Jun19")
			c.delivery = c.expiry
			continue
		case "SEP":
			c.expiry, _ = time.Parse("2Jan06", "27Sep19")
			c.delivery = c.expiry
			continue
		case "BTC":
			c.underlying = Pair{BTC, USDT}
			continue
		case "ETH":
			c.underlying = Pair{ETH, USDT}
			continue
		case "C":
			c.callPut = Call
			c.isOption = true
			continue
		case "P":
			c.callPut = Put
			c.isOption = true
			continue
		case "":
			continue
		}
		if d, err := time.Parse("2Jan06", strings.ToTitle(s)); err == nil {
			c.expiry = d
			c.delivery = d
			continue
		}
		if n, err := strconv.Atoi(s); err == nil {
			c.strike = float64(n)
			c.isOption = true
			continue
		}
		return c, errors.New("Don't recognise:" + s + example)
	}
	return c, nil
}

func PerpContract(p Pair) Contract {
	return Contract{
		isOption:   false,
		perp:       true,
		expiry:     time.Now(),
		underlying: p}
}

func PositionsFromNames(names []string, quantities []float64, prices []float64) (posns Positions, err error) {
	var c Contract
	posns = make(Positions, 0)
	for i := range names {
		c, err = ContractFromName(names[i])
		if err != nil {
			return
		}
		var p Position
		if prices == nil || quantities == nil {
			p = NewPosition(c, 0.0, 0.0)
		} else {
			p = NewPosition(c, quantities[i], prices[i])
		}
		posns = append(posns, p)
	}
	return
}

func NewPosition(c Contract, qty, price float64) Position {
	return Position{Con: c, Qty: qty, Price: price}
}

func OptContractFromDets(c Coin, d time.Time, strike float64, cp CallOrPut) Contract {
	return Contract{
		isOption:   true,
		underlying: Pair{c, USDT},
		expiry:     d,
		delivery:   d,
		strike:     strike,
		callPut:    cp}
}

func FutContractFromDets(c Coin, d time.Time) Contract {
	return Contract{
		isOption:   false,
		underlying: Pair{c, USDT},
		expiry:     d,
		delivery:   d,
		callPut:    NA}
}

func (c Contract) UnderFuture() Contract {
	if c.IsOption() {
		return Contract{
			isOption:   false,
			underlying: c.underlying,
			expiry:     c.expiry,
			delivery:   c.delivery,
			strike:     0.0,
			callPut:    NA}
	} else {
		return c
	}
}

func (c *Contract) Name() string {
	if c.name == "" {
		if c.isOption {
			var cptext string
			if c.callPut == Call {
				cptext = "C"
			} else {
				cptext = "P"
			}
			c.name = fmt.Sprintf("%s-%s-%4.0f-%s", c.underlying.Coin, strings.ToUpper(c.expiry.Format("2Jan06")), c.strike, cptext)
		} else {
			if c.perp {
				c.name = fmt.Sprintf("%s-PERPETUAL", c.underlying.Coin)
			} else {
				c.name = fmt.Sprintf("%s-%s", c.underlying.Coin, strings.ToUpper(c.expiry.Format("2Jan06")))
			}
		}
	}
	return c.name
}

func (c Contract) Expiry() (dt time.Time) {
	return c.expiry
}

func (c Contract) Perp() bool {
	return c.perp
}

func (c Contract) Delivery() time.Time {
	return c.delivery
}

func (c Contract) Strike() (st float64) {
	return c.strike
}

func (c Contract) CallPut() CallOrPut {
	return c.callPut
}

func (c Contract) Underlying() Pair {
	return c.underlying
}

func (c Contract) IsOption() bool {
	return c.isOption
}

// if a call, return the identical put and vice versa
func (c Contract) CallPutMirror() (p Contract) {
	p = c
	if c.callPut == Call {
		p.callPut = Put
	} else {
		p.callPut = Call
	}
	p.name = ""
	return
}

// Calculate the implied vol of a contract given its price in LHS coin value spot
func (c Contract) ImpVol(asof time.Time, spotPrice, futPrice, optionPrice float64) float64 {
	if !c.IsOption() {
		return math.NaN()
	}
	expiry := c.Expiry()
	strike := c.Strike()
	cp := c.CallPut()
	expiryDays := dayDiff(asof, expiry)
	deliveryDays := expiryDays // temp

	return optionImpliedVol(expiryDays, deliveryDays, strike, spotPrice, futPrice, optionPrice*spotPrice, cp)
}

func (c Contract) OptPrice(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	if c.IsOption() {
		expiry := c.Expiry()
		expiryDays := dayDiff(asof, expiry)
		strike := c.Strike()
		cp := c.CallPut()
		//		return (forwardOptionPrice(expiryDays, strike, futPrice, vol, cp)*spotPrice/futPrice - p.Price*spotPrice) * p.Qty
		// deribit includes option price in the cash balance
		return (forwardOptionPrice(expiryDays, strike, futPrice, vol, cp) * spotPrice / futPrice)
	} else {
		return math.NaN()
	}
}

// Calculate the price of a contract given market parameters. Price is in RHS coin value spot
// Discounting assumes zero interest rate on LHS coin (normally BTC) which is deribit standard. Note USDT rates float and are generally negative.
func (p Position) PV(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	if p.Con.IsOption() {
		return p.Con.OptPrice(asof, spotPrice, futPrice, vol) * p.Qty
	} else {
		return 10.0 * (1.0/p.Price - 1.0/futPrice) * spotPrice * p.Qty // Deribit futures in multiples of 10$. need to check the discounting
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
	guessVol := 0.80
	for i := 0; i < 1000; i++ {
		guessPrm := spot / forward * forwardOptionPrice(expiryDays, strike, forward, guessVol, callPut)
		vega := optionVega(expiryDays, deliveryDays, strike, spot, forward, guessVol)
		vega = math.Max(vega, 0.0001*spot) // floor the vega at 1bp to avoid guesses flying off
		guessVol = guessVol - (guessPrm-prm)/(vega*100.0)
		guessVol = math.Max(guessVol, 0.0) // floor guess vol at zero
		guessVol = math.Min(guessVol, 4.0) // cap guess vol at 400%
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
	return 0.5 * math.Erfc(-x/math.Sqrt(2))
}

func optionVega(expiryDays, deliveryDays int, strike, spot, forward, vol float64) float64 {
	//	d1 := (math.Log(forward/strike) + (vol*vol/2.0)*(float64(expiryDays)/365)) / (vol * math.Sqrt(float64(expiryDays)/365))
	//	return forward * cumNormDist(d1) * math.Sqrt(float64(expiryDays)/365.0) * dF(deliveryDays, domRate)
	return spot / forward * (forwardOptionPrice(expiryDays, strike, forward, vol+0.005, Call) - forwardOptionPrice(expiryDays, strike, forward, vol-0.005, Call))
}
