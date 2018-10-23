package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
)

func main() {
	ex := bean.NewRPCExchangeC("tcp", bean.MDS_HOST_SG40+":"+bean.MDS_PORT) // create an RPC exchange client
	pair := Pair{BTC, USDT}                                                 // pair to query for the orderbook
	ob, _ := ex.GetOrderBook(pair)                                          // making the query
	fmt.Println(ob)                                                         // print it out
}
