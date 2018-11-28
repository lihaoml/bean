package main

import (
	. "bean"
	"bean/brew"
	"bean/rpc"
	"bean/strats"
	"fmt"
	"time"
)

func main() {
	pair := Pair{BTC, USDT}

	start := time.Date(2018, time.November, 10, 0, 0, 0, 0, time.Local)
	end := time.Date(2018, time.November, 11, 0, 0, 0, 0, time.Local)
	freq := time.Second * 60

	//	for largebias := 0.0; largebias <= 1.0; largebias += 0.25 {
	//		for positionbias := 0.0; positionbias <= 1.0; positionbias += 0.25 {
	largebias := 0.25
	positionbias := 0.25
	strat := strats.NewStackBiasMM(NameBinance, pair, freq, 1.0, 5.0, largebias, positionbias, true)

	bt := brew.NewBackTest(bean.MDS_HOST_SG40, bean.MDS_PORT)
	result := bt.Simulate(strat, start, end, NewPortfolio())
	fmt.Printf("stack bias %0.2f, position bias %0.2f\n", largebias, positionbias)
	result.Show()
	//		}
	//	}

	//	result.Evaluate(BTC)
}
