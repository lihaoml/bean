package bean

import (
	"math"
	"sort"
	"sync"
)

// OrderBook1 is an implementation of the OrderBookCore interface. Bids and asks are stored as lists of orders
type OrderBook1 struct {
	bids []Order
	asks []Order
	m    sync.Mutex
}

// Bids retrieves a list of bid orders from the orderbook
func (ob *OrderBook1) Bids() []Order {
	return ob.bids
}

// Asks retrieves a list of asks from the orderbook
func (ob *OrderBook1) Asks() []Order {
	return ob.asks
}

// EmptyOrderBook returns an empty orderbook
func EmptyOrderBook() OrderBook {
	return OrderBook{new(OrderBook1)}
}

// NewOrderBook returns a new order book populated by bids and offers
func NewOrderBook(bids, asks []Order) OrderBook {
	ob := OrderBook1{bids: bids, asks: asks}.Sort()
	return OrderBook{&ob}
}

// InsertBid adds a new order into the orderbook. Returns true if the top of book price has changed
func (ob *OrderBook1) InsertBid(order Order) (tob bool) {
	ob.m.Lock()
	defer ob.m.Unlock()
	ob.bids = append(ob.bids, order)
	ob.Sort()
	return order.Price == ob.bids[0].Price
}

// InsertAsk adds a new order into the orderbook. Returns true if the top of book price has changed
func (ob *OrderBook1) InsertAsk(order Order) (tob bool) {
	ob.m.Lock()
	defer ob.m.Unlock()
	ob.asks = append(ob.asks, order)
	ob.Sort()
	return order.Price == ob.asks[0].Price
}

// CancelBid deletes an order from the orderbook. Returns true if the top of book price has changed
func (ob *OrderBook1) CancelBid(order Order) (tob bool) {
	ob.m.Lock()
	defer ob.m.Unlock()
	for i := range ob.bids {
		if ob.bids[i].Price == order.Price {
			ob.bids = append(ob.bids[:i], ob.bids[i+1:]...)
			if i == 0 {
				tob = true
			}
			break
		}
	}
	return
}

// CancelAsk deletes an order from the orderbook. Returns true if the top of book price has changed
func (ob *OrderBook1) CancelAsk(order Order) (tob bool) {
	ob.m.Lock()
	defer ob.m.Unlock()
	for i := range ob.asks {
		if ob.asks[i].Price == order.Price {
			ob.asks = append(ob.asks[:i], ob.asks[i+1:]...)
			if i == 0 {
				tob = true
			}
			break
		}
	}
	return
}

// EditBid replaces an order at a particular level with another. Returns true if the top of book has changed
func (ob *OrderBook1) EditBid(order Order) (tob bool) {
	ob.m.Lock()
	defer ob.m.Unlock()
	for i := range ob.bids {
		if ob.bids[i].Price == order.Price {
			ob.bids[i].Amount = order.Amount
			break
		}
	}
	return
}

// EditAsk replaces an order at a particular level with another. Returns true if the top of book has changed
func (ob *OrderBook1) EditAsk(order Order) (tob bool) {
	ob.m.Lock()
	defer ob.m.Unlock()
	for i := range ob.asks {
		if ob.asks[i].Price == order.Price {
			ob.asks[i].Amount = order.Amount
			break
		}
	}
	return
}

func (ob *OrderBook1) BestBid() Order {
	if ob != nil && len(ob.bids) > 0 {
		return ob.bids[0]
	} else {
		return Order{Price: math.NaN(), Amount: 0.0}
	}
}

func (ob *OrderBook1) BestAsk() Order {
	if ob != nil && len(ob.asks) > 0 {
		return ob.asks[0]
	} else {
		return Order{Price: math.NaN(), Amount: 0.0}
	}
}
func (ob OrderBook1) Sort() OrderBook1 {
	// asks in ascending order
	sort.Slice(ob.asks, func(i, j int) bool { return ob.asks[i].Price < ob.asks[j].Price })
	// bids in descending order
	sort.Slice(ob.bids, func(i, j int) bool { return ob.bids[i].Price > ob.bids[j].Price })
	return ob
}
