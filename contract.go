package bean

import (
	"errors"
	"fmt"
	"math"
	"sort"
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

func (c *Contract) Hash() (hash int) {
	if c.perp {
		hash = 0
	} else if c.isOption {
		hash = int(c.strike) + int(c.expiry.YearDay())
		if c.callPut == Put {
			hash++
		} else {
			hash = int(c.expiry.YearDay())
		}
	}

	return
}

var conCacheLock sync.Mutex
var contractCache = make(map[string]*Contract)

type Position struct {
	Con   *Contract
	Qty   float64
	Price float64
}

type Positions []Position

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
			//			dt, err := time.Parse("2Jan06", strings.ToTitle(st[1]))
			expiry, err = strToExpiry(st[1])
			if err != nil {
				return nil, err
			}
			//			expiry = time.Date(dt.Year(), dt.Month(), dt.Day(), 8, 0, 0, 0, time.UTC) // 8am london expiry
			con = &Contract{
				underlying: underlying,
				expiry:     expiry,
				delivery:   expiry,
				callPut:    NA}
		}

	case 4:
		//		dt, err := time.Parse("2Jan06", strings.ToTitle(st[1]))
		expiry, err = strToExpiry(st[1])
		if err != nil {
			return nil, err
		}
		//		expiry = time.Date(dt.Year(), dt.Month(), dt.Day(), 8, 0, 0, 0, time.UTC) // 8am london expiry

		strikei, err := strconv.Atoi(st[2])
		if err != nil {
			return nil, err
		}
		strike = float64(strikei)

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

// strToTime converts dates in the strict format DMMMYY or DDMMMYY
// hopefully faster than the more generic time.Parse
func strToExpiry(s string) (t time.Time, err error) {
	// date must be 2FEB20 or 22MAR21
	var daystr, monthstr, yearstr string
	switch len(s) {
	case 6:
		daystr = s[0:1]
		monthstr = s[1:4]
		yearstr = s[4:6]
	case 7:
		daystr = s[0:2]
		monthstr = s[2:5]
		yearstr = s[5:7]
	default:
		err = errors.New("Date not recognised:" + s)
		return
	}

	day, err := strconv.Atoi(daystr)
	if err != nil {
		err = errors.New("Date not recognised:" + s)
		return
	}
	year, err := strconv.Atoi(yearstr)
	if err != nil {
		err = errors.New("Date not recognised:" + s)
		return
	}
	var month time.Month
	switch monthstr {
	case "JAN":
		month = time.January
	case "FEB":
		month = time.February
	case "MAR":
		month = time.March
	case "APR":
		month = time.April
	case "MAY":
		month = time.May
	case "JUN":
		month = time.June
	case "JUL":
		month = time.July
	case "AUG":
		month = time.August
	case "SEP":
		month = time.September
	case "OCT":
		month = time.October
	case "NOV":
		month = time.November
	case "DEC":
		month = time.December
	default:
		err = errors.New("Contract date not recognised" + s)
	}

	t = time.Date(year+2000, month, day, 8, 0, 0, 0, time.UTC)
	return
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

func (p Positions) Sort() Positions {
	sort.Slice(p,
		func(i, j int) bool {
			// Sort by date
			if p[i].Con.Delivery().Before(p[j].Con.Delivery()) {
				return true
			}
			if p[i].Con.Delivery().After(p[j].Con.Delivery()) {
				return false
			}
			// Futures first
			if !p[i].Con.IsOption() {
				return true
			}
			if !p[j].Con.IsOption() {
				return false
			}
			// Then by strike
			return p[i].Con.Strike() < p[j].Con.Strike()
		})
	return p
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
func (c *Contract) CallPutMirror() *Contract {
	p := *c
	if c.callPut == Call {
		p.callPut = Put
	} else {
		p.callPut = Call
	}
	p.name = ""
	return &p
}

// Calculate the implied vol of a contract given its price in LHS coin value spot
func (c Contract) ImpVol(asof time.Time, spotPrice, futPrice, optionPrice float64) float64 {
	if !c.IsOption() {
		return math.NaN()
	}
	strike := c.Strike()
	cp := c.CallPut()
	expiryDays := c.ExpiryDays(asof)
	deliveryDays := expiryDays // temp

	return optionImpliedVol(expiryDays, deliveryDays, strike, spotPrice, futPrice, optionPrice*spotPrice, cp)
}

func (c Contract) OptPrice(asof time.Time, spotPrice, futPrice, vol float64) float64 {
	if c.IsOption() {
		expiryDays := c.ExpiryDays(asof)
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
	expiryDays := c.ExpiryDays(asof)
	if c.callPut == Call {
		return cumNormDist((math.Log(futPrice / c.strike)) / (vol * math.Sqrt(float64(expiryDays)/365.0)))
		//		return cumNormDist((math.Log(futPrice/c.strike) + (vol*vol/2.0)*(float64(expiryDays)/365.0)) / (vol * math.Sqrt(float64(expiryDays)/365.0)))
	} else { // put
		return cumNormDist((math.Log(futPrice/c.strike))/(vol*math.Sqrt(float64(expiryDays)/365.0))) - 1.0
		//		return cumNormDist((math.Log(futPrice/c.strike)+(vol*vol/2.0)*(float64(expiryDays)/365.0))/(vol*math.Sqrt(float64(expiryDays)/365.0))) - 1.0
	}
}

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

// DayDiff returns numbers of days from t1 to t2 after rounding
func DayDiff(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC) // remove time information and force to utc
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)
	return int(math.Round(t2.Sub(t1).Truncate(time.Hour).Hours() / 24.0))
}

func (c Contract) ExpiryDays(now time.Time) int {
	return DayDiff(now, c.Expiry())
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
