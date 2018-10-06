package bean

import (
	"time"
)

type Strat interface {
	GetExchangeNames() []string
	GetPairs() []Pair // start with a single pair, TODO: extend to multi pair
	GetTick() time.Duration
	Work(exs map[string]*Exchange) // work to be done for each tick
}
