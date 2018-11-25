package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
	"time"
	"flag"
)

func main() {
	var dbhost string
	flag.StringVar(&dbhost, "db", bean.MDS_HOST_SG40, "db host address")
	flag.Parse()

	mds := bean.NewRPCMDSConnC("tcp", dbhost+":"+bean.MDS_PORT)
	pair := Pair{BTC, USDT}


	end := time.Now()
	start := end.Add(time.Duration(-10) * time.Second)
	fmt.Println("Orderbook history from", start.Format("15:04:05"), "to", end.Format("15:04:05"))

	// open book history
	obts, _ := mds.GetOrderBookTS(pair, start, end, 20)
	fmt.Println("OB retrieval time: ", time.Now().Sub(end), "-------------")
	obts.ShowBrief()
/*
	fut := "BTC-28DEC18"
	// open book history
	obts, _ = mds.GetFutOrderBookTS(fut, start, end, 20)
	fmt.Println("OBFut retrieval time: ", time.Now().Sub(end), "-------------")
	obts.ShowBrief()
*/
	end = time.Now()
	// transactino history
	fmt.Println("Transaction history from", start.Format("15:04:05"), "to", end.Format("15:04:05"))
	txn, _ := mds.GetTransactions(pair, end.Add(time.Duration(-1)*time.Minute), end)

	fmt.Println("retrieval time: ", time.Now().Sub(end), "-------------")
	fmt.Println(txn)
}
