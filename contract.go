package bean

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
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

var conCacheLock sync.Mutex
var contractCache = make(map[string]*Contract)

func ContractFromName(name string) (*Contract, error) {
	var expiry time.Time
	var callPut CallOrPut
	var underlying Pair
	var strike float64
	var err error

	conCacheLock.Lock()
	defer conCacheLock.Unlock()

	con, exists := contractCache[name]
	if exists {
		return con, nil
	}

	st := strings.Split(name, "-")
	if len(st) < 2 {
		return nil, errors.New("Bad contract formation")
	}

	switch st[0] {
	case "BTC":
		underlying = Pair{BTC, USD}
	case "ETH":
		underlying = Pair{ETH, USD}
	default:
		err = errors.New("do not recognise coin")
		return nil, err
	}

	switch len(st) {
	case 2:
		if st[1] == "PERPETUAL" {
			con = PerpContract(underlying)
		} else {
			dt, err := time.Parse("2Jan06", strings.ToTitle(st[1]))
			if err != nil {
				return nil, err
			}
			expiry = time.Date(dt.Year(), dt.Month(), dt.Day(), 8, 0, 0, 0, time.UTC) // 8am london expiry
			con = &Contract{
				underlying: underlying,
				expiry:     expiry,
				delivery:   expiry,
				callPut:    NA}
		}

	case 4:
		dt, err := time.Parse("2Jan06", strings.ToTitle(st[1]))
		if err != nil {
			return nil, err
		}
		expiry = time.Date(dt.Year(), dt.Month(), dt.Day(), 8, 0, 0, 0, time.UTC) // 8am london expiry

		strike, err = strconv.ParseFloat(st[2], 64)
		if err != nil {
			return nil, err
		}

		switch st[3] {
		case "C":
			callPut = Call
		case "P":
			callPut = Put
		default:
			return nil, errors.New("Need C OR P")

		}
		con = &Contract{
			isOption:   true,
			underlying: underlying,
			expiry:     expiry,
			delivery:   expiry,
			callPut:    callPut,
			strike:     strike}

	default:
		return nil, errors.New("not a good contract formation")
	}

	contractCache[name] = con
	return con, nil
}

// ContractFromPartialName accepts contracts in the form
// dec mar-10000 btc-jun-10000-c fri 2fr perp
func ContractFromPartialName(partialName string) (*Contract, error) {
	const example = "\nDon't understand contract\nExample JUN or 3500 or MAR-4000-C or BTC-3000-P"
	sts := strings.Split(partialName, "-")
	defaultExpiry, _ := time.Parse("02Jan06", "28Jun19")
	c := Contract{
		isOption:   false,
		underlying: Pair{BTC, USD},
		expiry:     defaultExpiry,
		delivery:   defaultExpiry,
		callPut:    Call,
		strike:     5000}

	for _, s := range sts {
		switch strings.ToUpper(s) {
		case "PERP":
			c.perp = true
			n := time.Now()
			tod := time.Date(n.Year(), n.Month(), n.Day(), 8, 0, 0, 0, time.UTC)
			c.expiry = tod
			continue
		case "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC":
			// Find the last friday of the relevant month
			tod := time.Now()
			mth, _ := time.Parse("Jan", strings.ToUpper(s))
			followingMonth := mth.Month()%12 + 1 // January is 1
			year := tod.Year()
			if tod.Month() >= followingMonth {
				year++ // Months before today - select next year
			}
			dt := time.Date(year, followingMonth, 1, 8.0, 0, 0, 0, time.UTC) // first day of the following month
			daysToAdd := -1 - (dt.Weekday()+1)%7                             // go back to the strictly previous friday
			c.expiry = dt.Add(time.Duration(daysToAdd) * time.Hour * 24)
			c.delivery = c.expiry
			continue
		case "FRI": // The next friday date. Today if a friday
			n := time.Now()
			tod := time.Date(n.Year(), n.Month(), n.Day(), 8, 0, 0, 0, time.UTC)
			daysToAdd := (5 - int64(tod.Weekday())) % 7
			c.expiry = tod.Add(time.Duration(daysToAdd) * time.Hour * 24)
			c.delivery = c.expiry
			continue
		case "2FR": // The following friday
			n := time.Now()
			tod := time.Date(n.Year(), n.Month(), n.Day(), 8, 0, 0, 0, time.UTC)
			daysToAdd := (5-int64(tod.Weekday()))%7 + 7
			c.expiry = tod.Add(time.Duration(daysToAdd) * time.Hour * 24)
			c.delivery = c.expiry
			continue
		case "BTC":
			c.underlying = Pair{BTC, USD}
			continue
		case "ETH":
			c.underlying = Pair{ETH, USD}
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
		return &c, errors.New("Don't recognise:" + s + example)
	}
	return &c, nil
}

func PerpContract(p Pair) *Contract {
	n := time.Now()
	tod := time.Date(n.Year(), n.Month(), n.Day(), 8, 0, 0, 0, time.UTC)
	return &Contract{
		perp:       true,
		expiry:     tod,
		underlying: p}
}

func PositionsFromNames(names []string, quantities []float64, prices []float64) (posns Positions, err error) {
	var c *Contract
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

func NewPosition(c *Contract, qty, price float64) Position {
	return Position{Con: c, Qty: qty, Price: price}
}

func OptContract(p Pair, d time.Time, strike float64, cp CallOrPut) *Contract {
	return &Contract{
		isOption:   true,
		underlying: p,
		expiry:     d,
		delivery:   d,
		strike:     strike,
		callPut:    cp}
}

func FutContract(p Pair, d time.Time) *Contract {
	return &Contract{
		isOption:   false,
		underlying: p,
		expiry:     d,
		delivery:   d,
		callPut:    NA}
}

func (c *Contract) UnderFuture() *Contract {
	if c.IsOption() {
		return &Contract{
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
			c.name = fmt.Sprintf("%s-%s-%.0f-%s", c.underlying.Coin, strings.ToUpper(c.expiry.Format("2Jan06")), c.strike, cptext)
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

func (c Contract) IsFuture() bool {
	return !c.isOption
}

func (c1 *Contract) Equal(c2 *Contract) bool {
	if c1.isOption {
		return c2.isOption &&
			c1.callPut == c2.callPut &&
			c1.expiry == c2.expiry &&
			c1.delivery == c2.delivery &&
			c1.strike == c2.strike &&
			c1.underlying == c2.underlying
	} else {
		return !c2.isOption &&
			c1.underlying == c2.underlying &&
			(c1.perp == c2.perp || c1.expiry == c2.expiry)
	}
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

// Return the 'simple' delta computed analytically
func (c Contract) SimpleDelta(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	expiryDays := dayDiff(asof, c.expiry)
	if c.callPut == Call {
		return cumNormDist((math.Log(futPrice / c.strike)) / (vol * math.Sqrt(float64(expiryDays)/365.0)))
		//		return cumNormDist((math.Log(futPrice/c.strike) + (vol*vol/2.0)*(float64(expiryDays)/365.0)) / (vol * math.Sqrt(float64(expiryDays)/365.0)))
	} else { // put
		return cumNormDist((math.Log(futPrice/c.strike))/(vol*math.Sqrt(float64(expiryDays)/365.0))) - 1.0
		//		return cumNormDist((math.Log(futPrice/c.strike)+(vol*vol/2.0)*(float64(expiryDays)/365.0))/(vol*math.Sqrt(float64(expiryDays)/365.0))) - 1.0
	}
}
