# bean
A multi cryptocurrency exchange interface that supports
* retrieval of historical transaction data and limit order book
* retrieval of real time limit order book

and 

* backtest and simulation of trading strategies
* real trading strategies 

Bean is a lightweight wrapper of a bunch of remote protocal calls (RPC). 

# Get Real Time OrderBook

* Get real time order book [example](example/exchange/main.go):
```go
package main

import (
	. "bean"
	"bean/rpc"
	"fmt"
)

func main() {
	ex := bean.NewRPCExchangeC("tcp", "13.229.125.250:9892") // create an RPC exchange client
	pair := Pair{BTC, USDT} // pair to query for the orderbook
	ob, _ := ex.GetOrderBook(pair) // making the query
	fmt.Println(ob) // print it out
}
```

To run it:

``` 
go run example/orderbook/main.go 
```

# Retrieve historical orderbook and transactions [example](example/mds/main.go):
``` go run example/mds/main.go ```

# To implement trading strategy, look at [simplemm](strats/simplemm.go)

# To backtest a trading strategy, look at [run_backtest](example/simplemm/main.go)
``` go run example/simplemm/main.go ```

After backtesting the result is visualized on webpage: http://localhost:8080
