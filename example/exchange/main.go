package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
)

func main() {
	ex := bean.NewRPCExchangeC("tcp", "13.229.125.250:9892")
	pair := Pair{BTC, USDT}
	ob, _ := ex.GetOrderBook(pair)
	fmt.Println(ob)
}
