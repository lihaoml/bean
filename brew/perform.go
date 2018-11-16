package brew

import (
	. "bean"
)

func PerformActions(exs *(map[string]Exchange), actions []TradeAction) {
	// exchange to deal with the actions
	for _, act := range actions {
		switch act.Op {
		case PlaceLimitOrder:
			// place order,
			(*exs)[act.ExName].PlaceLimitOrder(
				act.Pair,
				act.Params["price"].(float64),
				act.Params["amount"].(float64),
			)
			break
		case CancelOpenOrder:
			(*exs)[act.ExName].CancelOrder(
				act.Pair,
				act.Params["orderid"].(string),
			)
			break
		}
	}
}
