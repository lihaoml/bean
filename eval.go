package bean

import (
	"fmt"
	"github.com/gonum/floats"
	"github.com/gonum/stat"
	"math"
	"time"
)

type Perf struct {
	Inception time.Time   // starting time
	Interval  time.Duration  // time inverval between each snapshot
	Port []Portfolio  // portfolio snapshot for each interval
	MtMBTC  []float64  // mtm snapshot for each interval, in BTC
	MtMUSD  []float64  // mtm snapshot for each interval, in USD
}

func (perf Perf) Sharpe() float64 {
	pnl := make([]float64, len(perf.MtMUSD) - 1)
	for i := 1; i < len(perf.MtMUSD); i++ {
		pnl[i-1] = perf.MtMUSD[i] - perf.MtMUSD[i-1]
	}
	fmt.Println(pnl)
	ret := perf.MtMUSD[len(pnl)] / float64(len(pnl))
	vol := stat.StdDev(pnl, nil)
	return ret / vol * math.Sqrt(365 * float64(time.Hour) * 24 / float64(perf.Interval))
}

// get Drawdown series and MaxDrawdown
func (perf Perf) MaxDrawdown() float64 {
	var drawdown []float64
	var maxsofar float64
	for i, v := range perf.MtMUSD {
		if i == 0 || v > maxsofar {
			maxsofar = v
		}
		drawdown = append(drawdown, maxsofar-v)
	}
	return floats.Max(drawdown)
}

// evaluate performance of TradeLogS, assuming initial portoflio is empty.
// dividing trades into intervals, and generate Perf stats, mtmBase is normally USDT and/or BTC
func (tls TradeLogS) GenPerf0(interval time.Duration) (perf Perf) {
	init := NewPortfolio()
	return tls.GenPerf(init, interval)
}

// evaluate performance of TradeLogS
// dividing trades into intervals, and generate Perf stats, mtmBase is normally USDT and/or BTC
func (tls TradeLogS) GenPerf(init Portfolio, interval time.Duration) (perf Perf) {
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

	// TODO: re-engineer below
	for i, p := range perf.Port {
		mtmBTC := 0.0
		for c, v := range p.Balances() {
			// FIXME: feed rates to the function would be better
			mtmBTC += v * ts.inBTC(c, perf.Inception.Add(interval * time.Duration(i)))
		}
		if (i == 0) {
			fmt.Println(mtmBTC)
			fmt.Println(p.Balances())
			fmt.Println(init.Balances())
		}
		perf.MtMBTC = append(perf.MtMBTC, mtmBTC)
		if mtmBTC == 0 {
			perf.MtMUSD = append(perf.MtMUSD, 0.0)
		} else {
			perf.MtMUSD = append(perf.MtMUSD, mtmBTC / ts.inBTC(USDT, perf.Inception.Add(interval * time.Duration(i))))
		}
	}
	return
}

// FIXME: we can do better
// assuming tls is sorted
func (tls TradeLogS) inBTC(coin Coin, cut time.Time) float64 {
	if coin == BTC {
		return 1.0
	} else {
		rate := math.NaN()
		for _, t := range tls {
			if t.Pair.Coin == coin && t.Pair.Base == BTC {
				rate = t.Price
			}
			if t.Pair.Base == coin && t.Pair.Coin == BTC {
				rate = 1.0 / t.Price
			}
			if t.Time.After(cut) {
				break
			}
		}
		return rate
	}
}