package main

import (
	. "bean"
	"bean/rpc"
	"flag"
	"fmt"
	"time"
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

	fut := "BTC-28DEC18"
	// open book history
	obts, _ = mds.GetFutOrderBookTS(fut, start, end, 20)
	fmt.Println("FutOB retrieval time: ", time.Now().Sub(end), "-------------")
	obts.ShowBrief()

	opt := "BTC-28DEC18-4250-C"
	// open book history
	start2 := end.Add(time.Duration(-60) * time.Second) // option OB doesn't update that frequently
	obts, _ = mds.GetOptOrderBookTS(opt, start2, end, 20)
	fmt.Println("OptOB retrieval time: ", time.Now().Sub(end), "-------------")
	obts.ShowBrief()

	end = time.Now()
	// transactino history
	fmt.Println("Transaction history from", start.Format("15:04:05"), "to", end.Format("15:04:05"))
	txn, _ := mds.GetTransactions(pair, end.Add(time.Duration(-1)*time.Minute), end)

	fmt.Println("retrieval time: ", time.Now().Sub(end), "-------------")
	fmt.Println("#Transaction: ", len(txn))
}
