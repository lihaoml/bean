package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
	"time"
)

func main() {
	mds := bean.NewRPCMDSConnC("tcp", bean.MDS_HOST_SG40+":"+bean.MDS_PORT)
	pair := Pair{BTC, USDT}

	end := time.Now()
	start := end.Add(time.Duration(-10) * time.Second)
	fmt.Println("Orderbook history from", start.Format("15:04:05"), "to", end.Format("15:04:05"))

	// open book history
	obts, _ := mds.GetOrderBookTS(pair, start, end, 20)
	fmt.Println("retrieval time: ", time.Now().Sub(end), "-------------")
	obts.ShowBrief()

	end = time.Now()
	// transactino history
	fmt.Println("Transaction history from", start.Format("15:04:05"), "to", end.Format("15:04:05"))
	txn, _ := mds.GetTransactions(pair, end.Add(time.Duration(-1)*time.Minute), end)

	fmt.Println("retrieval time: ", time.Now().Sub(end), "-------------")
	fmt.Println(txn)
}
