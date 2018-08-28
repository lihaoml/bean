package bean

import (
	"net/rpc"
	"log"
	"fmt"
	. "bean"
)

//////////////////////////////////////////////////////////////
// the RPC exchange client instance
type RPCExchangeC struct {
	client *rpc.Client
}

func NewRPCExchangeC(network, address string) RPCExchangeC {
	// Create a TCP connection to localhost on port 1234
	// client, err := rpc.DialHTTP("tcp", "localhost:9892")
	client, err := rpc.DialHTTP(network, address)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return RPCExchangeC{client}
}

func (ex RPCExchangeC) Name() string {
	return "NOT IMPLEMENTED"
}

// get the open orders for a currency pair,
// when the exchange query fails, return an empty order book
func (ex RPCExchangeC) GetOrderBook (pair Pair) (OrderBook, error) {
	var err error
	var ob OrderBook
	err = ex.client.Call("RPCExchangeD.GetOrderBook", pair, &ob)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(ob)
	}
	return ob, err
}

func (ex RPCExchangeC) GetPairs(base Coin) ([]Pair, error) {
	var pairs []Pair
	return pairs, nil // not implemented
}
