package strats

import (
	. "bean"
	"time"
)

type SimpleMM struct {
	BaseStrat
	pair   Pair
	exName string
	spread float64
}

func NewSimpleMM(exName string, pair Pair, spread float64, tick time.Duration) SimpleMM {
	return SimpleMM{
		exName:    exName,
		pair:      pair,
		BaseStrat: BaseStrat{tick},
		spread:    spread,
	}
}

func (s SimpleMM) GetExchangeNames() []string {
	return []string{s.exName}
}

func (s SimpleMM) GetPairs() []Pair {
	return []Pair{s.pair}
}

func (s SimpleMM) Grind(exs map[string]Exchange) []TradeAction {
	ex := exs[s.exName]
	ob := ex.GetOrderBook(s.pair)
	// ob.ShowBrief()
	/*
		txn := ex.GetTransactionHistory(s.pair)
		fmt.Println(len(txn))
	*/
	// mid, vol, volume := exchange.GetHistStats(s.ex, s.pair, s.tick * 5)
	amt := 1000.0

	var actions []TradeAction
	if ob.Valid() {
		// TODO: need to denoise
		bestBid := ob.Bids[0].Price
		bestAsk := ob.Asks[0].Price
		mid := (bestAsk + bestBid) / 2.0
		buyPrice := mid * (1 - s.spread)
		sellPrice := mid * (1 + s.spread)
		actBuy := PlaceLimitOrderAction(s.exName, s.pair, buyPrice, amt)
		actSell := PlaceLimitOrderAction(s.exName, s.pair, sellPrice, -amt)

		actions = append(actions, actBuy)
		actions = append(actions, actSell)
	}
	return actions
}
