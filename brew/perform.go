package brew

import (
	. "bean"
	"time"
)

type ExNameWithOID struct {
	ExName    string
	Pair      Pair
	OrderID   string
	TimeStamp time.Time
}

func PerformActions(exs *(map[string]Exchange), actions []TradeAction, sep ...time.Duration) (cancelled, placed []ExNameWithOID) {
	// exchange to deal with the actions
	for _, act := range actions {
		if len(sep) > 0 {
			time.Sleep(sep[0])
		}
		switch act.Op {
		case PlaceLimitOrder:
			// place order,
			oid, _ := (*exs)[act.ExName].PlaceLimitOrder(
				act.Pair,
				act.Params["price"].(float64),
				act.Params["amount"].(float64),
			)
			(*exs)[act.ExName].TrackOrderID(act.Pair, oid)
			placed = append(placed, ExNameWithOID{act.ExName, act.Pair, oid, time.Now()})
			break
		case CancelOpenOrder:
			oid := act.Params["orderid"].(string)
			(*exs)[act.ExName].CancelOrder(
				act.Pair,
				oid,
			)
			cancelled = append(cancelled, ExNameWithOID{act.ExName, act.Pair, oid, time.Now()})
			break
		}
	}
	return
}
