package main

import (
	"beanex/exchange"
	"bean"
	"fmt"
)

func main()  {
	ex := exchange.NewExchange(bean.NameBinance)
	cum := ex.GetOrderBook(bean.Pair{bean.BTC, bean.USDT}).CumPctOB()
	fmt.Println("cum:", cum)
	fmt.Println("cumask:", cum.CumPctAsks)
	fmt.Println("cumbid:", cum.CumPctBids)
}
