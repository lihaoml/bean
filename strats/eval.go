package strats

import (
	. "bean"
	"bean/rpc"
	"fmt"
	"github.com/gonum/stat"
	"math"
	"os"
	"time"
)

type Perf struct {
	Inception time.Time     // starting time
	Interval  time.Duration // time inverval between each snapshot
	Port      []Portfolio   // portfolio snapshot for each interval
	MtMBTC    []float64     // mtm snapshot for each interval, in BTC
	MtMUSD    []float64     // mtm snapshot for each interval, in USD
}

func (perf Perf) Sharpe() float64 {
	pnl := make([]float64, len(perf.MtMUSD)-1)
	for i := 1; i < len(perf.MtMUSD); i++ {
		pnl[i-1] = perf.MtMUSD[i] - perf.MtMUSD[i-1]
	}
	fmt.Println(pnl)
	ret := perf.MtMUSD[len(pnl)] / float64(len(pnl))
	vol := stat.StdDev(pnl, nil)
	return ret / vol * math.Sqrt(365*float64(time.Hour)*24/float64(perf.Interval))
}

// get Drawdown series and MaxDrawdown
func (perf Perf) MaxDrawdown() float64 {
	return MaxDD(perf.MtMUSD)
}

// evaluate performance of TradeLogS, assuming initial portoflio is empty.
// dividing trades into intervals, and generate Perf stats, mtmBase is normally USDT and/or BTC
func GenPerf0(tls TradeLogS, interval time.Duration) (perf Perf) {
	init := NewPortfolio()
	pairs := tls.Pairs()
	mds := bean.NewRPCMDSConnC("tcp", os.Getenv("MDS_DB_ADDRESS")+":"+bean.MDS_PORT)
	// get start and end of trade logs
	if len(tls) == 0 {
		return
	}
	start := tls[0].Time
	end := tls[0].Time
	for _, t := range tls {
		if t.Time.Before(start) {
			start = t.Time
		}
		if t.Time.After(end) {
			end = t.Time
		}
	}
	// FIXME: try to sample the transaction
	txn, _ := mds.GetTransactions(Pair{BTC, USDT}, start, end)
	for _, c := range AllCoins(pairs) {
		if c != BTC && c != USDT {
			txn_, _ := mds.GetTransactions(Pair{c, BTC}, start, end)
			txn = append(txn, txn_...)
		}
	}
	// sample transaction by interval
	return GenPerf(tls, init, interval, txn.Sort())
}

// evaluate performance of TradeLogS
// dividing trades into intervals, and generate Perf stats
func GenPerf(tls TradeLogS, init Portfolio, interval time.Duration, txn Transactions) (perf Perf) {
	if len(tls) == 0 {
		return
	}
	ts := tls.Sort() // make sure trade log is sorted
	perf.Inception = ts[0].Time.Truncate(interval)
	perf.Interval = interval
	perf.Port = []Portfolio{init.Clone()}
	next := perf.Inception.Add(interval)
	port := init.Clone()
	// generate portfolio snapshots
	for _, t := range ts {
		if !t.Time.Before(next) {
			perf.Port = append(perf.Port, port.Clone())
			next = next.Add(interval)
		}
		port = port.Age([]TradeLog{t})
	}
	perf.Port = append(perf.Port, port.Clone())

	for i, p := range perf.Port {
		mtmBTC := 0.0
		for c, v := range p.Balances() {
			mtmBTC += v * inBTC(txn, c, perf.Inception.Add(interval*time.Duration(i)))
		}
		if i == 0 {
			fmt.Println(mtmBTC)
			fmt.Println(p.Balances())
			fmt.Println(init.Balances())
		}
		perf.MtMBTC = append(perf.MtMBTC, mtmBTC)
		if mtmBTC == 0 {
			perf.MtMUSD = append(perf.MtMUSD, 0.0)
		} else {
			perf.MtMUSD = append(perf.MtMUSD, mtmBTC/inBTC(txn, USDT, perf.Inception.Add(interval*time.Duration(i))))
		}
	}
	return
}

// FIXME: we can do better
// assuming tls is sorted
func inBTC(txn Transactions, coin Coin, cut time.Time) float64 {
	if coin == BTC {
		return 1.0
	} else {
		rate := math.NaN()
		for _, t := range txn {
			if t.Pair.Coin == coin && t.Pair.Base == BTC {
				rate = t.Price
			}
			if t.Pair.Base == coin && t.Pair.Coin == BTC {
				rate = 1.0 / t.Price
			}
			if t.TimeStamp.After(cut) {
				break
			}
		}
		return rate
	}
}
