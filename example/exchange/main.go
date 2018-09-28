package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
)

func main() {
	ex := bean.NewRPCExchangeC("tcp", "13.229.125.250:9892") // create an RPC exchange client
	pair := Pair{BTC, USDT}                                  // pair to query for the orderbook
	ob, _ := ex.GetOrderBook(pair)                           // making the query
	fmt.Println(ob)                                          // print it out
}
