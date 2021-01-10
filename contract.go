package bean

import (
	"errors"
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

const ContractDateFormat = "2Jan06"
const optionPriceRounding = 0.0005
const futurePriceRounding = 0.5

type Contract struct {
	name       string
	isOption   bool
	underlying Pair
	expiry     time.Time
	delivery   time.Time
	strike     float64
	callPut    CallOrPut
	perp       bool
	index      bool
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

func ContractFromName(name string) (*Contract, error) {
	var expiry time.Time
	var callPut CallOrPut
	var underlying Pair
	var strike float64
	var err error

	conCacheLock.Lock()
	defer conCacheLock.Unlock()
	// Faster without cache ?
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
	case "BCH":
		underlying = Pair{BCH, USD}
	default:
		err = errors.New("do not recognise coin")
		return nil, err
	}

	switch len(st) {
	case 2:
		if st[1] == "PERPETUAL" {
			con = PerpContract(underlying)
		} else {
			//			dt, err := time.Parse(ContractDateFormat, strings.ToTitle(st[1]))
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
	case 3:
		if st[1] == "DERIBIT" && st[2] == "INDEX" {
			con = IndexContract(underlying)
		} else {
			return nil, errors.New("Don't recognise contract")
		}

	case 4:
		//		dt, err := time.Parse(ContractDateFormat, strings.ToTitle(st[1]))
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
			c.expiry = n.Add(24 * time.Hour)
			continue
		case "INDEX":
			c.index = true
			n := time.Now()
			c.expiry = n
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
		case "BCH":
			c.underlying = Pair{BCH, USD}
			continue
		case "XRP":
			c.underlying = Pair{XRP, USD}
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
		if d, err := time.Parse(ContractDateFormat, strings.ToTitle(s)); err == nil {
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
	return &Contract{
		perp:       true,
		expiry:     n.Add(24 * time.Hour),
		delivery:   n.Add(24 * time.Hour),
		underlying: p}
}

func IndexContract(p Pair) *Contract {
	n := time.Now()
	return &Contract{
		index:      true,
		expiry:     n,
		delivery:   n,
		underlying: p,
	}
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
			c.name = string(c.underlying.Coin) + "-" + c.ExpiryStr() + "-" + strconv.FormatFloat(c.strike, 'f', 0, 64) + "-" + cptext
		} else {
			if c.perp {
				c.name = string(c.underlying.Coin) + "-PERPETUAL"
			} else if c.index {
				c.name = string(c.underlying.Coin) + "-DERIBIT-INDEX"
			} else {
				c.name = string(c.underlying.Coin) + "-" + c.ExpiryStr()
			}
		}
	}
	return c.name
}

func (c Contract) Expiry() (dt time.Time) {
	return c.expiry
}

func (c Contract) ExpiryStr() string {
	return strings.ToUpper(c.Expiry().Format(ContractDateFormat))
}

func (c Contract) Perp() bool {
	return c.perp
}

func (c Contract) Index() bool {
	return c.index
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
	return !c.isOption && !c.index
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
			((c1.perp && c2.perp) ||
				(!c1.perp && !c2.perp && c1.expiry == c2.expiry))
	}
}

func ContractSort(cons []*Contract) {
	sort.Slice(cons, func(i, j int) bool { return cons[i].Before(cons[j]) })
}

func (c1 *Contract) Before(c2 *Contract) bool {
	// Sort by date
	if c1.Delivery().Before(c2.Delivery()) {
		return true
	}
	if c1.Delivery().After(c2.Delivery()) {
		return false
	}
	// Futures first
	if c1.IsOption() && !c2.IsOption() {
		return true
	}
	if c2.IsOption() && !c1.IsOption() {
		return false
	}
	// Then by strike
	if c1.Strike() < c2.Strike() {
		return true
	}
	if c1.Strike() > c2.Strike() {
		return false
	}
	if c1.CallPut() == Call {
		return true
	} else {
		return false
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

func (c *Contract) RoundPrice(price float64) float64 {
	if c.IsOption() {
		return math.Round(price/optionPriceRounding) * optionPriceRounding
	} else {
		return math.Round(price/futurePriceRounding) * futurePriceRounding
	}
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

// maths stuff now

// DayDiff returns numbers of days from t1 to t2 after rounding
func DayDiff(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC) // remove time information and force to utc
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)
	return int(math.Round(t2.Sub(t1).Truncate(time.Hour).Hours() / 24.0))
}

func (c Contract) ExpiryDays(now time.Time) float64 {
	return c.Expiry().Sub(now).Hours() / 24.0 //DayDiff(now, c.Expiry())
}

// premium expected in domestic - rhs coin value spot
func optionImpliedVol(expiryDays, deliveryDays, strike, spot, forward, prm float64, callPut CallOrPut) (bs float64) {

	if expiryDays <= 0 {
		return math.NaN()
	}
	if expiryDays <= 0.1 {
		expiryDays = 0.1
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

func dF(days float64, rate float64) float64 {
	return math.Exp(-days / 365 * rate)
}

// in domestic - rhs coin forward value
func forwardOptionPrice(expiryDays, strike, forward, vol float64, callPut CallOrPut) (prm float64) {
	if expiryDays <= 0 {
		vol = 0
	}

	d1 := (math.Log(forward/strike) + (vol*vol/2.0)*(expiryDays/365)) / (vol * math.Sqrt(expiryDays/365))
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

func optionVega(expiryDays, deliveryDays, strike, spot, forward, vol float64) float64 {
	//	d1 := (math.Log(forward/strike) + (vol*vol/2.0)*(float64(expiryDays)/365)) / (vol * math.Sqrt(float64(expiryDays)/365))
	//	return forward * cumNormDist(d1) * math.Sqrt(float64(expiryDays)/365.0) * dF(deliveryDays, domRate)
	return spot / forward * (forwardOptionPrice(expiryDays, strike, forward, vol+0.005, Call) - forwardOptionPrice(expiryDays, strike, forward, vol-0.005, Call))
}
