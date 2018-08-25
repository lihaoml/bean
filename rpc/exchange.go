package rpc

//////////////////////////////////////////////////////////////
// the RPC exchange instance
type rpcExchange struct {
	name string
	url string
	port int
}

// TODO: implement a NewExchange function that takes a exchange name and returns an rpcExchange instance
/*
func (ex Exchange) Name() string {
	return ex.name
}

// get the open orders for a currency pair,
// when the exchange query fails, return an empty order book
func (ex Exchange) GetOrderBook (pair Pair) (OrderBook, error) {
	// make the RPC call
	var ob OrderBook
	return ob, nil
}


func (ex Exchange) GetPairs(base Coin) ([]Pair, error) {
	var pairs []Pair
	return pairs, nil
}
*/
