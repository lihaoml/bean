package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
	"time"
)

func main() {
	mds := bean.NewRPCMDSConnC("tcp", "13.229.125.250:9892")
	pair := Pair{BTC, USDT}
	for {
		end := time.Now()
		start := end.Add(time.Duration(-10)*time.Minute)
		fmt.Println("Orderbook history from", start, "to", end)
		obts, _ := mds.GetOrderBookTS(pair, start, end, 20)
		fmt.Println(time.Now().Sub(end))

		printOrderBookTS(obts)

		fmt.Println("Transaction history from", start, "to", end)
		txn, _ := mds.GetTransactions(pair, start, end)

		fmt.Println(txn)
		time.Sleep(time.Second * 2)
	}
}

func printOrderBookTS (obts OrderBookTS) {
	for _, ob := range obts {
		fmt.Println(ob.Time, len(ob.OB.Asks), "bestBid:", ob.OB.Bids[0].Price, "bestAsk:", ob.OB.Asks[0].Price)
	}
}