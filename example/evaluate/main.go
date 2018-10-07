package main

import (
	. "bean"
	"fmt"
	"time"
)

func main() {

	tm := time.Now()
	var ts1 ReferenceRateTS
	var ts2 ReferenceRateTS
	fmt.Printf("ts1=======\n")
	for i := 0; i < 5; i++ {
		unit := float64(i)
		time := tm.Add(3 * time.Duration(i) * time.Second)
		price := 3 * 100 * unit
		rate := ReferenceRate{Time: time, Price: price}
		ts1 = append(ts1, rate)
		fmt.Printf("%v:%v\n", time, price)

	}
	fmt.Printf("ts2=======\n")
	for i := 0; i < 5; i++ {
		unit := float64(i)
		time := tm.Add(5 * time.Duration(i) * time.Second)
		price := 5 * 10 * unit
		rate := ReferenceRate{Time: time, Price: price}
		ts2 = append(ts2, rate)
		fmt.Printf("%v:%v\n", time, price)
	}
	ratesbook := ReferenceRateBook{Pair{Coin: BTC, Base: USDT}: ts1, Pair{Coin: FT, Base: USDT}: ts2}

	// pair := Pair{Coin: IOTX, Base: USDT}
	// for i := 0; i < 16; i++ {
	// 	tmCheck := tm.Add(time.Duration(i) * time.Second)
	// 	lookup := LookupRate(pair, tmCheck, ratesbook)
	// 	fmt.Printf("BTCUSDT: %v\n", lookup)

	// }

	t1 := Transaction{
		Pair:      Pair{Coin: BTC, Base: USDT},
		Price:     100,
		Amount:    2,
		TimeStamp: tm.Add(4 * time.Second),
		Maker:     Buyer,
	}

	t2 := Transaction{
		Pair:      Pair{Coin: FT, Base: USDT},
		Price:     5,
		Amount:    10,
		TimeStamp: tm.Add(4 * time.Second),
		Maker:     Seller,
	}

	t3 := Transaction{
		Pair:      Pair{Coin: BTC, Base: USDT},
		Price:     50,
		Amount:    10,
		TimeStamp: tm.Add(8 * time.Second),
		Maker:     Seller,
	}

	p := NewPortfolio()
	p.AddBalance(BTC, 50)
	p.AddBalance(USDT, 1000)

	var t Transactions
	t = append(t, t1, t2, t3)
	snapts := GenerateSnapshotTS(t, p)
	perfts := EvaluateSnapshotTS(snapts, USDT, ratesbook)

	for _, v := range perfts {
		fmt.Printf("%v,%v,%v,%v\n", v.MtMBase, v.Time, v.PV, v.DailyPnL)
	}

	// negative balance?
	// if reference rate unavailable?
}
