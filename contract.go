package bean

import (
	"errors"
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
	NA   callOrPut = "N"
)

const USDiscountRate = 0.02

type Contract struct {
	isOption   bool
	underlying Pair
	expiry     time.Time
	delivery   time.Time
	strike     float64
	callPut    callOrPut
	perp       bool
}

type Contracts map[Contract]float64

func ContractFromName(name string) (Contract, error) {
	var expiry time.Time
	var callPut callOrPut
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
	sts := strings.Split(partialName, "-")
	defaultExpiry, _ := time.Parse("02Jan06", "29Mar19")
	c := Contract{
		isOption:   false,
		underlying: Pair{BTC, USDT},
		expiry:     defaultExpiry,
		delivery:   defaultExpiry,
		callPut:    Call,
		strike:     3500}

	for _, s := range sts {
		switch strings.ToUpper(s) {
		case "PERP":
			c.perp = true
			c.isOption = false
			continue
		case "MAR":
			c.expiry, _ = time.Parse("02Jan06", "29Mar19")
			c.delivery = c.expiry
			continue
		case "JUN":
			c.expiry, _ = time.Parse("02Jan06", "28Jun19")
			c.delivery = c.expiry
			continue
		case "SEP":
			c.expiry, _ = time.Parse("02Jan06", "27Sep19")
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
		if d, err := time.Parse("02Jan06", strings.ToTitle(s)); err == nil {
			c.expiry = d
			c.delivery = d
			continue
		}
		if n, err := strconv.Atoi(s); err == nil {
			c.strike = float64(n)
			c.isOption = true
			continue
		}
		return c, errors.New("Don't recognise as contract component:" + s)
	}
	return c, nil
}

func PerpContract(c Coin) (Contract, error) {
	return ContractFromName(string(c) + "-PERPETUAL")
}

func ContractsFromNames(names []string, quantities []float64) (pos Contracts, err error) {
	var c Contract
	pos = make(Contracts)
	for i := range names {
		c, err = ContractFromName(names[i])
		if err != nil {
			return
		}
		if quantities != nil {
			pos[c] = quantities[i]
		} else {
			pos[c] = 0.0
		}
	}
	return
}

func OptContractFromDets(c Coin, d time.Time, strike float64, cp callOrPut) Contract {
	return Contract{
		isOption:   true,
		underlying: Pair{c, USDT},
		expiry:     d,
		delivery:   d,
		strike:     strike,
		callPut:    cp}
}

func FutContractFromDets(c Coin, d time.Time, price float64) Contract {
	return Contract{
		isOption:   false,
		underlying: Pair{c, USDT},
		expiry:     d,
		delivery:   d,
		strike:     price, // strike doubles up as price for futures
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

func (c Contract) Name() string {
	if c.isOption {
		var cptext string
		if c.callPut == Call {
			cptext = "C"
		} else {
			cptext = "P"
		}
		return fmt.Sprintf("%s-%s-%4.0f-%s", c.underlying.Coin, strings.ToUpper(c.expiry.Format("2Jan06")), c.strike, cptext)
	} else {
		if c.perp {
			return fmt.Sprintf("%s-PERPETUAL", c.underlying.Coin)
		} else {
			return fmt.Sprintf("%s-%s", c.underlying.Coin, strings.ToUpper(c.expiry.Format("2Jan06")))
		}
	}
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

func (c Contract) CallPut() callOrPut {
	return c.callPut
}

func (c Contract) Underlying() Pair {
	return c.underlying
}

func (c Contract) IsOption() bool {
	return c.isOption
}

// this is temporary until we add price field
func (c *Contract) SetPrice(price float64) {
	c.strike = price
}

// Calculate the implied vol of a contract given its price in LHS coin
func (c Contract) ImpVol(asof time.Time, spotPrice, futPrice, domRate, optionPrice float64) float64 {
	expiry := c.Expiry()
	strike := c.Strike()
	cp := c.CallPut()
	expiryDays := dayDiff(asof, expiry)
	deliveryDays := expiryDays // temp

	return optionImpliedVol(expiryDays, deliveryDays, strike, futPrice, domRate, optionPrice*spotPrice, cp)
}

// Calculate the price of a contract given market parameters. Price is in RHS coin
func (c Contract) PV(asof time.Time, spotPrice, futPrice, domRate, vol float64) float64 {
	expiry := c.Expiry()
	expiryDays := dayDiff(asof, expiry)
	deliveryDays := expiryDays // temp
	if c.IsOption() {
		strike := c.Strike()
		cp := c.CallPut()
		return forwardOptionPrice(expiryDays, strike, futPrice, vol, cp) * dF(deliveryDays, domRate)
	} else {
		return 10.0 * (1.0/c.Strike() - 1.0/futPrice) * futPrice // strike doubles up as future price. Deribit futures in multiples of 10$. need to check the discounting
	}
}

// in rhs coin
func (c Contract) Vega(asof time.Time, spotPrice, futPrice, vol, domRate float64) float64 {
	return c.PV(asof, spotPrice, futPrice, domRate, vol+0.005) - c.PV(asof, spotPrice, futPrice, domRate, vol-0.005)
}

//in lhs coin
func (c Contract) Delta(asof time.Time, spotPrice, futPrice, vol, domRate float64) float64 {
	deltaFiat := (c.PV(asof, spotPrice*1.005, futPrice*1.005, domRate, vol) - c.PV(asof, spotPrice*0.995, futPrice*0.995, domRate, vol)) * 100.0
	return deltaFiat / spotPrice
}

//in lhs coin
func (c Contract) Gamma(asof time.Time, spotPrice, futPrice, vol, domRate float64) float64 {
	gammaFiat := c.Delta(asof, spotPrice*1.005, futPrice*1.005, domRate, vol) - c.Delta(asof, spotPrice*0.995, futPrice*0.995, domRate, vol)

	return gammaFiat
}

//in rhs coin
func (c Contract) Theta(asof time.Time, spotPrice, futPrice, vol, domRate float64) float64 {
	return c.PV(asof.Add(24*time.Hour), spotPrice, futPrice, domRate, vol) - c.PV(asof, spotPrice, futPrice, domRate, vol)
}

// maths stuff now

// day difference rounded.
func dayDiff(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC) // remove time information and force to utc
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)
	return int(math.Round(t2.Sub(t1).Truncate(time.Hour).Hours() / 24.0))
}

// premium expected in domestic - rhs coin
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

// in domestic - rhs coin
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
