package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
	"time"
)

func main() {
	ex := bean.NewRPCExchangeC("tcp", "ss.w4ip.com:9892")
	pair := Pair{BTC, USDT}

	for {
		ob, _ := ex.GetOrderBook(pair)
		fmt.Println(ob)
		time.Sleep(time.Second * 5)
	}
}
