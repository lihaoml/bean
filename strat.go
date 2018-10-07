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

type Strat interface {
	GetExchangeNames() []string                  // get exchange names used
	GetPairs() []Pair                            // get relevant pairs
	GetTick() time.Duration                      // get tick duration
	Grind(exs map[string]Exchange) []TradeAction // called at each tick, generates trading actions
}
