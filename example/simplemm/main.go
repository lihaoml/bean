package main

import (
	. "bean"
	"bean/brew"
	"bean/rpc"
	"bean/strats"
	"time"
)

func main() {
	pair := Pair{IOTX, ETH}
	strat := strats.NewSimpleMM(NameBinance, pair, 0.01, time.Second*60)

	start := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2018, time.October, 2, 0, 0, 0, 0, time.Local)

	bt := brew.NewBackTest(bean.MDS_HOST_SG40, bean.MDS_PORT)
	port := NewPortfolio()
	result := bt.Simulate(strat, start, end, port)

	result.Show()
	//	result.Evaluate(USDT) // if we were to evaluate in different base, untested.
}
