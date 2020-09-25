package bean

import (
	"fmt"
	"math"
	"time"
)

type Operation int

const ( // iota is reset to 0
	PlaceLimitOrder Operation = 0
	CancelOpenOrder Operation = 1
	Wait            Operation = 2
)

// exchange is a struct for holding common member variables and base functions
type BaseStrat struct {
	Tick time.Duration
}

func (s BaseStrat) GetTick() time.Duration {
	return s.Tick
}

func (s BaseStrat) Name() string {
	return "bean"
}

func (s BaseStrat) FormatParams() string {
	return ""
}

type TradeAction struct {
	ExName string
	Op     Operation
	Pair   Pair
	Params map[string]interface{}
}

func (t TradeAction) Show() string {
	switch t.Op {
	case PlaceLimitOrder:
		if t.Params["amount"].(float64) > 0 {
			return "`b[" + t.Pair.String() + "]:" + t.ExName[0:3] + " " + t.Pair.FormatPrice(t.Params["price"].(float64)) + " " + fmt.Sprint(math.Round(t.Params["amount"].(float64)*100)/100) + "`\n"
		} else {
			return "`s[" + t.Pair.String() + "]:" + t.ExName[0:3] + " " + t.Pair.FormatPrice(t.Params["price"].(float64)) + " " + fmt.Sprint(math.Round(t.Params["amount"].(float64)*100)/100) + "`\n"
		}
	case CancelOpenOrder:
		return fmt.Sprint(t.ExName[0:2], " Cancel order ", t.Pair, t.Params["orderid"])
	case Wait:
		return fmt.Sprint("Wait for ", t.Params["time"], " seconds")
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

func WaitAction(nSec int) TradeAction {
	params := make(map[string]interface{})
	params["time"] = nSec
	return TradeAction{
		Op:     Wait,
		Params: params,
	}
}

type Strat interface {
	GetExchangeNames() []string                  // get exchange names used
	GetPairs() []Pair                            // get relevant pairs
	GetTick() time.Duration                      // get tick duration
	Grind(exs map[string]Exchange) []TradeAction // called at each tick, generates trading actions

	Name() string         // name of the strategy, for reporting purpose
	FormatParams() string // a compact way to format key params of the strategy, to save to TDS for reference
}

// the strat info to push to the bus for dashboard to subscribe
type StratDashBoardInfo struct {
	ColNames []string   `json:"ColNames"`
	Rows     [][]string `json:"Rows"`
}