package bean

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
)

// OrderBook: an orderbook from exchange
type OrderBook struct {
	Bids []Order
	Asks []Order
}

// timed order book
type OrderBookT struct {
	Time time.Time
	OB   OrderBook
}

type OrderBookTS []OrderBookT

type Order struct {
	Price  float64
	Amount float64
}

func (ob *OrderBook) BestBid() Order {
	if ob != nil && len(ob.Bids) > 0 {
		return ob.Bids[0]
	} else {
		return Order{Price: 0.0, Amount: 99.0} // There's always a zero bid in any amount
	}
}

func (ob *OrderBook) BestAsk() Order {
	if ob != nil && len(ob.Asks) > 0 {
		return ob.Asks[0]
	} else {
		return Order{Price: math.NaN(), Amount: 0.0}
	}
}

func (ob *OrderBook) Mid() float64 {
	if ob != nil && len(ob.Bids) > 0 && len(ob.Asks) > 0 {
		return (ob.Bids[0].Price + ob.Asks[0].Price) / 2.0
	} else {
		return math.NaN()
	}
}

func (ob *OrderBook) BidAskMid() (bid, ask, mid float64, err error) {
	if ob == nil {
		bid, ask, mid = math.NaN(), math.NaN(), math.NaN()
		err = errors.New("Don't recognise contract")
		return
	}
	if len(ob.Bids) > 0 && len(ob.Asks) > 0 {
		bid = ob.Bids[0].Price
		ask = ob.Asks[0].Price
		mid = (bid + ask) / 2.0
	} else if len(ob.Asks) > 0 {
		bid = math.NaN()
		ask = ob.Asks[0].Price
		mid = ask
	} else if len(ob.Bids) > 0 {
		bid = ob.Bids[0].Price
		ask = math.NaN()
		mid = bid
	} else {
		bid, ask, mid = math.NaN(), math.NaN(), math.NaN()
	}
	return
}

func (ob OrderBook) Spread() float64 {
	if ob.Valid() {
		return ob.Asks[0].Price - ob.Bids[0].Price
	} else {
		return math.NaN()
	}
}

func (ob OrderBook) Valid() bool {
	return len(ob.Bids) > 0 && len(ob.Asks) > 0
}

func (ob OrderBook) Copy() OrderBook {
	ob2 := OrderBook{
		Bids: make([]Order, len(ob.Bids)),
		Asks: make([]Order, len(ob.Asks)),
	}
	for i := range ob.Bids {
		ob2.Bids[i] = ob.Bids[i]
	}
	for i := range ob.Asks {
		ob2.Asks[i] = ob.Asks[i]
	}
	return ob2
}

func (obt OrderBookT) Copy() OrderBookT {
	return OrderBookT{
		Time: obt.Time,
		OB:   obt.OB.Copy(),
	}
}

// Compare two orderbooks. Equal if the best bid and best offer hasn't changed
func (ob1 *OrderBook) Equal(ob2 *OrderBook) bool {
	if len(ob1.Bids) > 0 && len(ob2.Bids) > 0 && ob1.Bids[0].Price == ob2.Bids[0].Price &&
		len(ob1.Asks) > 0 && len(ob2.Asks) > 0 && ob1.Asks[0].Price == ob2.Asks[0].Price {
		return true
	} else {
		return false
	}
}

// filter out orders with amount less than the Coin minimum trading amount
// assuming ob is sorted
func Denoise(pair Pair, ob OrderBook) OrderBook {
	var bids []Order
	var asks []Order
	minimumAmount := pair.MinimumTradingAmount()
	for i, b := range ob.Bids {
		if b.Amount < minimumAmount {
			if i+1 < len(ob.Bids) {
				ob.Bids[i+1].Amount += b.Amount
			}
		} else {
			bids = append(bids, b)
		}
	}
	for i, a := range ob.Asks {
		if a.Amount < minimumAmount {
			if i+1 < len(ob.Asks) {
				ob.Asks[i+1].Amount += a.Amount
			}
		} else {
			asks = append(asks, a)
		}
	}
	return OrderBook{Bids: bids, Asks: asks}
}

func (ob OrderBook) Sort() OrderBook {
	// asks in ascending order
	sort.Slice(ob.Asks, func(i, j int) bool { return ob.Asks[i].Price < ob.Asks[j].Price })
	// bids in descending order
	sort.Slice(ob.Bids, func(i, j int) bool { return ob.Bids[i].Price > ob.Bids[j].Price })
	return ob
}

// OrderBook display
func (ob OrderBook) ShowBrief() {
	if ob.Valid() {
		fmt.Println("depth:", len(ob.Asks), "bestBid:", ob.Bids[0].Price, "bestAsk:", ob.Asks[0].Price)
	} else {
		fmt.Println("empty orderbook")
	}
}

// OrderBookT display
func (ob OrderBookT) ShowBrief() {
	if ob.OB.Valid() {
		fmt.Println(ob.Time.Local().Format("Jan _2 15:04:05"), "depth:", len(ob.OB.Asks), "bestBid:", ob.OB.Bids[0].Price, "bestAsk:", ob.OB.Asks[0].Price)
	} else {
		fmt.Println(ob.Time.Local(), len(ob.OB.Asks))
	}
}

// OrderBookTS display
func (obts OrderBookTS) ShowBrief() {
	for _, ob := range obts {
		ob.ShowBrief()
	}
}

func (obts OrderBookTS) Sort() OrderBookTS {
	sort.Slice(obts, func(i, j int) bool { return obts[i].Time.Before(obts[j].Time) })
	return obts
}

// return the orderbook of time t (the closest in sample), assuming the obts is sorted
func (obts OrderBookTS) GetOrderBook(t time.Time) OrderBook {
	ob := obts[0].OB
	for _, obt := range obts {
		if t.After(obt.Time) {
			ob = obt.OB
		} else {
			break
		}
	}
	return ob
}

// Returns the worst bid and worst ask that need to be hit in the orderbook in order to execute a requested size
// Also returns the total size available at that price (may be more than requested size)
// If orderstack does not have sufficient liquidity, then it returns the size available
func (ob OrderBook) PriceIn(size float64) (bid, ask, bidSize, askSize float64) {
	bid, bidSize = ob.BidIn(size)
	ask, askSize = ob.AskIn(size)
	return
}

func (ob OrderBook) BidIn(size float64) (price, available float64) {
	price, available = priceInAmount(size, ob.Bids)
	return
}

func (ob OrderBook) AskIn(size float64) (price, available float64) {
	price, available = priceInAmount(size, ob.Asks)
	return
}

func priceInAmount(requiredAmount float64, stack []Order) (price, available float64) {
	available = 0.0

	if len(stack) == 0 {
		price = math.NaN()
		return
	}

	for _, ord := range stack {
		available += ord.Amount
		if available > requiredAmount {
			price = ord.Price
			return
		}
	}
	price = stack[len(stack)-1].Price
	return
}

// sell / buy ratio, alpha in (0, 1]
func (ob OrderBook) SBRatio(alpha float64) float64 {
	var sell float64
	var buy float64
	if ob.Valid() {
		// FIXME: generalize spread, work for IOTX at the moment
		sprd := (ob.Asks[0].Price - ob.Bids[0].Price) * 1e8

		for i, v := range ob.Asks {
			if i == 10 {
				break
			} else {
				sell += math.Pow(alpha, float64(i)) * v.Price * v.Amount
			}
		}
		for i, v := range ob.Bids {
			if i == 10 {
				break
			} else {
				buy += math.Pow(alpha, sprd-1+float64(i)) * v.Price * v.Amount
			}
		}
	}
	return sell / buy
}

// Match ... Takes a placed order and matches against the existing orderbook.
// If it can be filled then the filled amount and rate are returned
// Orders (aggressor) are filled at the orderbook (market maker) rate
func (ob OrderBook) Match(placedOrder Order) Order {
	fillCounterAmount := 0.0
	fillAmount := 0.0
	if placedOrder.Amount > 0.0 {
		for _, o := range ob.Asks {
			if o.Price <= placedOrder.Price {
				fillCounterAmount += math.Min(placedOrder.Amount-fillAmount, o.Amount) * o.Price
				fillAmount += math.Min(placedOrder.Amount-fillAmount, o.Amount)
			}
		}
		if fillAmount > 0.0 {
			return Order{Price: fillCounterAmount / fillAmount, Amount: fillAmount}
		} else {
			return Order{Price: 0.0, Amount: 0.0}
		}
	} else {
		for _, o := range ob.Bids {
			if o.Price >= placedOrder.Price {
				fillCounterAmount += math.Min(-placedOrder.Amount-fillAmount, o.Amount) * o.Price
				fillAmount += math.Min(-placedOrder.Amount-fillAmount, o.Amount)
			}
		}
		if fillAmount > 0.0 {
			return Order{Price: fillCounterAmount / fillAmount, Amount: -fillAmount}
		} else {
			return Order{Price: 0.0, Amount: 0.0}
		}
	}
}

type OrderState string

const (
	ALIVE     OrderState = "ALIVE"
	FILLED    OrderState = "FILLED"
	CANCELLED OrderState = "CANCELLED"
)

type Side string

const (
	BUY  Side = "BUY"
	SELL Side = "SELL"
)

func AmountToSide(amt float64) Side {
	if amt < 0.0 {
		return SELL
	} else {
		return BUY
	}
}

// Status of the placed order,
type OrderStatus struct {
	OrderID         string
	PlacedTime      time.Time
	Side            Side
	FilledAmount    float64
	LeftAmount      float64
	PlacedPrice     float64 // initial price
	Price           float64 // filled price, if not applicable then placed price
	State           OrderState
	Commission      float64
	CommissionAsset Coin
}
