package bean

import (
	"time"
)

type Operation int

const ( // iota is reset to 0
	PlaceLimitOrder Operation = 0
	CancelOrder     Operation = 1
)

type TradeAction struct {
	ExName string
	Op     Operation
	Params map[string]interface{}
}

func CancelOrderAction (exName string, pair Pair, oid string) TradeAction {
	params := make(map[string]interface{})
	params["pair"] = pair
	params["orderid"] = oid
	return TradeAction{
		ExName: exName,
		Op: CancelOrder,
		Params: params,
	}
}

func PlaceLimitOrderAction (exName string, pair Pair, price, amount float64) TradeAction {
	params := make(map[string]interface{})
	params["pair"] = pair
	params["price"] = price
	params["amount"] = amount
	return TradeAction{
		ExName: exName,
		Op: PlaceLimitOrder,
		Params: params,
	}
}

type Strat interface {
	GetExchangeNames() []string                  // get exchange names used
	GetPairs() []Pair                            // get relevant pairs
	GetTick() time.Duration                      // get tick duration
	Grind(exs map[string]Exchange) []TradeAction // called at each tick, generates trading actions
}
