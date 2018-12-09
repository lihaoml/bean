package brew

import (
	. "bean"
	"time"
)

type ExNameWithOID struct {
	ExName  string
	Pair    Pair
	OrderID string
}

func PerformActions(exs *(map[string]Exchange), actions []TradeAction) (cancelled, placed []ExNameWithOID) {
	// exchange to deal with the actions
	for _, act := range actions {
		time.Sleep(time.Millisecond * 300)
		switch act.Op {
		case PlaceLimitOrder:
			// place order,
			oid, _ := (*exs)[act.ExName].PlaceLimitOrder(
				act.Pair,
				act.Params["price"].(float64),
				act.Params["amount"].(float64),
			)
			placed = append(placed, ExNameWithOID{act.ExName, act.Pair, oid})
			break
		case CancelOpenOrder:
			oid := act.Params["orderid"].(string)
			(*exs)[act.ExName].CancelOrder(
				act.Pair,
				oid,
			)
			cancelled = append(cancelled, ExNameWithOID{act.ExName, act.Pair, oid})
			break
		}

	}
	return
}
