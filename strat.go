package bean

import (
	"time"
	"fmt"
)

type Operation int

const ( // iota is reset to 0
	PlaceLimitOrder Operation = 0
	CancelOpenOrder Operation = 1
)

type TradeAction struct {
	ExName string
	Op     Operation
	Pair   Pair
	Params map[string]interface{}
}

func (t TradeAction) Show() string {
	switch t.Op {
	case PlaceLimitOrder:
		return fmt.Sprint(t.ExName, "Placed order", t.Pair, t.Params["price"], t.Params["amount"])
	case CancelOpenOrder:
		return fmt.Sprint(t.ExName, "Cancel order", t.Pair, t.Params["orderid"])
	default:
		return fmt.Sprint(t)
	}
}

type TradeActionT struct {
	Time   time.Time
	Action TradeAction
}

func CancelOrderAction(exName string, pair Pair, oid string) TradeAction {
	params := make(map[string]interface{})
	params["orderid"] = oid
	return TradeAction{
		ExName: exName,
		Op:     CancelOpenOrder,
		Pair:   pair,
		Params: params,
	}
}

func PlaceLimitOrderAction(exName string, pair Pair, price, amount float64) TradeAction {
	params := make(map[string]interface{})
	params["price"] = price
	params["amount"] = amount
	return TradeAction{
		ExName: exName,
		Op:     PlaceLimitOrder,
		Pair:   pair,
		Params: params,
	}
}

type Strat interface {
	GetExchangeNames() []string                  // get exchange names used
	GetPairs() []Pair                            // get relevant pairs
	GetTick() time.Duration                      // get tick duration
	Grind(exs map[string]Exchange) []TradeAction // called at each tick, generates trading actions
}
