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
	strat := strats.NewSimpleMM(NameBinance, pair, 0.01, time.Second*30)

	start := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2018, time.October, 2, 0, 0, 0, 0, time.Local)

	bt := brew.NewBackTest(bean.MDS_HOST_SG40, bean.MDS_PORT)
	result := bt.Simulate(strat, start, end)

	result.Evaluate(ETH)
	result.Show()
}
