package main

import (
	. "bean"
	"bean/brew"
	"bean/rpc"
	"bean/strats"
	"time"
)

func main() {
	pair := Pair{BTC, USDT}

	start := time.Date(2018, time.November, 28, 00, 0, 0, 0, time.Local)
	end := time.Date(2018, time.November, 29, 00, 0, 0, 0, time.Local)
	freq := time.Second * 10

	strat := strats.NewOrderScan(NameBinance, pair, freq, 1.0, true)

	bt := brew.NewBackTest(bean.MDS_HOST_SG40, bean.MDS_PORT)
	result := bt.Simulate(strat, start, end, NewPortfolio())
	result.Evaluate(USDT)
	result.Show()
}
