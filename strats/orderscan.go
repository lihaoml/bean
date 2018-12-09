package strats

import (
	. "bean"
	"fmt"
	"math"
	"os"
	"time"
)

// we keep a track of the static (not moving) orders in the price stack and filter out the dynamic orders which are assumed to be market makers
// we trade ahead of the static orders

// will use this structure to keep track of persistant orders
type orderAge struct {
	Price  float64
	Amount float64
	Age    int
}

type OrderScan struct {
	pair          Pair
	exName        string
	tick          time.Duration
	staticBids    []orderAge // keep track of bids/age in the stack
	staticAsks    []orderAge // keep track of asks/age in the stack
	threshold     float64    // minimum size of order to track
	minAge        int        // minimum age (in cycles) of a static order
	tooFar        float64    // distance from mid when the order becomes irrelevent
	distance      float64    // how far ahead of the static order to place our order
	tradingAmount float64    // our trading amount
	dumpFile      *os.File   // if not null then dump relevent info to csv file
}

func NewOrderScan(exName string, pair Pair, tick time.Duration, tradingamount float64, dump bool) *OrderScan {
	var f *os.File
	if dump {
		f, _ = os.Create("orderscan.csv")
		fmt.Fprintf(f, "Position,Bid, Ask, CloseBid, Amount, Age,CloseAsk,Amount,Age,Trade\n")
	} else {
		f = nil
	}
	staticBids := make([]orderAge, 0, 10)
	staticAsks := make([]orderAge, 0, 10)
	return &OrderScan{
		exName:        exName,
		pair:          pair,
		tick:          tick,
		threshold:     10.0,
		distance:      2.0,
		tooFar:        10.0,
		minAge:        2,
		tradingAmount: tradingamount,
		dumpFile:      f,
		staticBids:    staticBids,
		staticAsks:    staticAsks,
	}
}

func (s OrderScan) GetExchangeNames() []string {
	return []string{s.exName}
}

func (s OrderScan) GetPairs() []Pair {
	return []Pair{s.pair}
}

func (s OrderScan) GetTick() time.Duration {
	return s.tick
}

// use a set of new orders to update an array of known static orders.
// increase age of orders still present. remove those known to have been cancelled. keep those beyond the range of the newstack.
func mergeOrders(ascending bool, threshold float64, staticOrders []orderAge, newOrders []Order) []orderAge {
	mergeOrder := make([]orderAge, 0, 10)

	// prepare to scan through the static orders
	var soi int

	// static orders below the first new order have probably been executed
	if len(staticOrders) > 0 {
		if ascending {
			for soi = 0; soi < len(staticOrders) && staticOrders[soi].Price < newOrders[0].Price; soi++ {
			}
		} else {
			for soi = 0; soi < len(staticOrders) && staticOrders[soi].Price > newOrders[0].Price; soi++ {
			}
		}
	}

	for _, nord := range newOrders {
		// scan for next matching or exceeding static order
		if ascending {
			for ; soi < len(staticOrders) && staticOrders[soi].Price < nord.Price; soi++ {
			}
		} else {
			for ; soi < len(staticOrders) && staticOrders[soi].Price > nord.Price; soi++ {
			}
		}
		// if matching then add to list with increased age. otherwsie add with age 1
		if nord.Amount > threshold {
			if soi < len(staticOrders) && staticOrders[soi].Price == nord.Price {
				mergeOrder = append(mergeOrder, orderAge{nord.Price, nord.Amount, staticOrders[soi].Age + 1})
			} else {
				mergeOrder = append(mergeOrder, orderAge{nord.Price, nord.Amount, 1})
			}
		}
	}

	// orders beyond the last new order are retained
	for ; soi < len(staticOrders); soi++ {
		mergeOrder = append(mergeOrder, staticOrders[soi])
	}

	return mergeOrder
}

func (s *OrderScan) Grind(exs map[string]Exchange) []TradeAction {
	var actions []TradeAction
	ex := exs[s.exName]

	// get price stack
	ob := ex.GetOrderBook(s.pair)
	if !ob.Valid() {
		return actions
	}

	// cancel existing orders
	myorders := ex.GetMyOrders(s.pair)
	for _, ord := range myorders {
		if ord.State == ALIVE {
			actions = append(actions, CancelOrderAction(s.exName, s.pair, ord.OrderID))
		}
	}

	// get position
	port := ex.GetPortfolio()
	position := port.Balance(s.pair.Coin)

	// merge the order stack with the existing static order stack / age
	s.staticAsks = mergeOrders(true, s.threshold, s.staticAsks, ob.Asks)
	s.staticBids = mergeOrders(false, s.threshold, s.staticBids, ob.Bids)

	tradingBid, _ := priceInAmount(s.tradingAmount, ob.Bids)
	tradingAsk, _ := priceInAmount(s.tradingAmount, ob.Asks)

	// identify the nearest known static order within range
	nearStaticBid := nearOldStaticOrder(s.staticBids, s.minAge, tradingBid, s.tooFar)
	nearStaticAsk := nearOldStaticOrder(s.staticAsks, s.minAge, tradingAsk, s.tooFar)

	if s.dumpFile != nil {
		fmt.Fprintf(s.dumpFile, "%3.1f,%6.1f,%6.1f", position, tradingBid, tradingAsk)
		if nearStaticBid != nil {
			fmt.Fprintf(s.dumpFile, ",%6.1f,%3.1f,%v", nearStaticBid.Price, nearStaticBid.Amount, nearStaticBid.Age)
		} else {
			fmt.Fprintf(s.dumpFile, ",,,")
		}

		if nearStaticAsk != nil {
			fmt.Fprintf(s.dumpFile, ",%6.1f,%3.1f,%v", nearStaticAsk.Price, nearStaticAsk.Amount, nearStaticAsk.Age)
		} else {
			fmt.Fprintf(s.dumpFile, ",,,")
		}
	}

	// this logic needs to be improved

	// if close orders on both sides then pick the closest. But what about size ????
	if nearStaticBid != nil && nearStaticAsk != nil {
		bidVicinity := tradingBid - nearStaticBid.Price
		askVicinity := nearStaticAsk.Price - tradingAsk
		if bidVicinity < s.distance && askVicinity < s.distance {
			nearStaticBid = nil
			nearStaticAsk = nil
		}
		if bidVicinity < askVicinity {
			nearStaticAsk = nil
		} else {
			nearStaticBid = nil
		}
	}

	// if static order on the bid but not the ask and position allows then place order ahead of it
	if nearStaticBid != nil && nearStaticAsk == nil && position < s.tradingAmount {
		level := math.Min(nearStaticBid.Price+s.distance, tradingAsk)
		act := PlaceLimitOrderAction(s.exName, s.pair, level, s.tradingAmount-position)
		actions = append(actions, act)
		if s.dumpFile != nil {
			fmt.Fprintf(s.dumpFile, ",%6.1f", level)
		}
	}
	// vice versa
	if nearStaticAsk != nil && nearStaticBid == nil && position > -s.tradingAmount {
		level := math.Max(nearStaticAsk.Price-s.distance, tradingBid)
		act := PlaceLimitOrderAction(s.exName, s.pair, level, -s.tradingAmount+position)
		actions = append(actions, act)
		if s.dumpFile != nil {
			fmt.Fprintf(s.dumpFile, ",%6.1f", level)
		}
	}
	// no static orders within range and we have a position, place an order at mid to close
	if nearStaticAsk == nil && nearStaticBid == nil && math.Abs(position) > 0.0 {
		level := (tradingBid + tradingAsk) / 2.0
		act := PlaceLimitOrderAction(s.exName, s.pair, level, -position)
		actions = append(actions, act)
		if s.dumpFile != nil {
			fmt.Fprintf(s.dumpFile, ",%6.1f", level)
		}

	}
	if s.dumpFile != nil {
		fmt.Fprintf(s.dumpFile, "\n")
	}

	return actions
}

// return the closest static order meeting the age / distance threshold
func nearOldStaticOrder(staticOrders []orderAge, minAge int, refPrice, minDist float64) *orderAge {
	for _, o := range staticOrders {
		if o.Age > minAge && math.Abs(o.Price-refPrice) < minDist {
			return &o
		}
	}
	return nil
}
