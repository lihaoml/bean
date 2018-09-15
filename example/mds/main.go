package mds

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
		start := time.Now()
		time.Sleep(time.Second * 10)
		end := time.Now()
		ob, _ := mds.GetOrderBookTS(pair, start, end, 2)
		txn, _ := mds.GetTransactions(pair, start, end)
		fmt.Println(ob)
		fmt.Println(txn)
	}
}
