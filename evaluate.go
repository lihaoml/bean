package bean

import (
	"fmt"
	"sort"
	"time"
)

//Snapshot represents the state of a portfolio at a specific time
type Snapshot struct {
	Time time.Time
	Port Portfolio
}

//SnapshotTS represents the states of a portfolio over time
type SnapshotTS []Snapshot

//Performance measures the value of a portfolio at a specific time
type Performance struct {
	Time     time.Time
	MtMBase  Coin
	PV       float64
	DailyPnL float64
}

//PerformanceTS measures the value of a portoflio over time
type PerformanceTS []Performance

//ReferenceRate holds the reference rates for one pair at a given time
type ReferenceRate struct {
	Time  time.Time
	Price float64
}

//ReferenceRateTS holds the reference rates for one pair over a period
type ReferenceRateTS []ReferenceRate

func RefRatesFromTxn(txn Transactions) ReferenceRateTS {
	var res ReferenceRateTS
	for _, t := range txn {
		r := ReferenceRate{
			Time:  t.TimeStamp,
			Price: t.Price,
		}
		res = append(res, r)
	}
	return res
}

//ReferenceRateBook holds the reference rates for multiple pairs over a period
type ReferenceRateBook map[Pair]ReferenceRateTS

//Sort ReferenceRateTS by time
func (t ReferenceRateTS) Sort() ReferenceRateTS {
	sort.Slice(t, func(i, j int) bool { return t[i].Time.Before(t[j].Time) })
	return t
}

//GenerateSnapshotTS update portfolio overtime
func GenerateSnapshotTS(ts Transactions, p Portfolio) SnapshotTS {
	ts.Sort()

	var snapts SnapshotTS

	for _, t := range ts {
		pClone := p.Clone()
		snapNew := GenerateSnapshot(t, pClone)
		// update inital snapshot
		if len(snapts) == 0 {
			snapts = append(snapts, Snapshot{Time: snapNew.Time.Add(-1 * time.Second), Port: p})
		}
		// update subsequest snapshot
		if snapts[len(snapts)-1].Time != snapNew.Time {
			snapts = append(snapts, snapNew)
		} else {
			snapts[len(snapts)-1] = snapNew
		}
		p = pClone

	}

	fmt.Printf("SnapshotTS:\n")
	for i, v := range snapts {
		fmt.Printf("%v,%v\n", i, v.Time)
		for key, value := range v.Port.Balances() {
			fmt.Printf("%v:%v\n", key, value)
		}
	}

	return snapts
}

//GenerateSnapshot updates the portfolio status after single transaction
func GenerateSnapshot(t Transaction, p Portfolio) Snapshot {
	var snap Snapshot
	coin := t.Pair.Coin
	base := t.Pair.Base
	var coinChange, baseChange float64

	coinChange = t.Amount
	baseChange = t.Price * t.Amount

	p.AddBalance(coin, coinChange)
	p.RemoveBalance(base, baseChange)

	snap.Port = p
	snap.Time = t.TimeStamp

	return snap

}

//EvaluateSnapshot shows the backtest performance of single snapshot
func EvaluateSnapshot(snap Snapshot, mtmBase Coin, ratesbook ReferenceRateBook) Performance {
	var perf Performance

	port := snap.Port
	perf.Time = snap.Time
	perf.MtMBase = mtmBase

	pv := 0.0

	for k, v := range port.Balances() {
		var rate float64
		if k == mtmBase {
			rate = 1
		} else {
			rate = LookupRate(Pair{Coin: k, Base: mtmBase}, snap.Time, ratesbook)
			fmt.Printf("%v%v,%v:%v\n", k, mtmBase, snap.Time, rate)
		}
		if rate != 0 {
			pv += v * rate
		} else {
			fmt.Printf("%v%v at %v is not availabe\n", k, mtmBase, snap.Time)
		}
	}

	perf.PV = pv

	return perf

}

//EvaluateSnapshotTS shows the backtest performance of a series of snapshots
func EvaluateSnapshotTS(snapts SnapshotTS, mtmBase Coin, ratesbook ReferenceRateBook) PerformanceTS {
	var perfts PerformanceTS

	for i, v := range snapts {
		perf := EvaluateSnapshot(v, mtmBase, ratesbook)
		if i > 0 {
			perf.DailyPnL = perf.PV - perfts[i-1].PV
		}
		perfts = append(perfts, perf)
	}
	return perfts
}

//LookupRate return the MTM exchange rate for given pair at a given time
func LookupRate(pair Pair, tm time.Time, ratesbook ReferenceRateBook) float64 {
	ratests := ratesbook[pair]
	ratests.Sort()

	//filter out price equal 0
	var validRatests ReferenceRateTS
	for _, v := range ratests {
		if v.Price != 0 {
			validRatests = append(validRatests, v)
		}
	}

	var prevTime time.Time
	var prevRate float64
	var currTime time.Time
	var currRate float64

	for i, rates := range validRatests {
		//if rate at tm not found, lookup for the nearest time to tm
		if i == 0 {
			prevTime = rates.Time
			prevRate = rates.Price
		}
		currTime = rates.Time
		currRate = rates.Price

		if currTime.Before(tm) {
			prevTime = currTime
			prevRate = currRate
			continue
		} else {
			break
		}
	}

	if currTime.After(tm) && prevTime.Before(tm) {
		duration1 := tm.Sub(prevTime)
		duration2 := currTime.Sub(tm)
		if duration1 < duration2 {
			return prevRate
		} else {
			return currRate
		}
	} else {
		return currRate
	}
}

// func LookupRate(pair Pair) float64 {
// 	BTCUSDT := Pair{Coin: BTC, Base: USDT}
// 	FTUSDT := Pair{Coin: FT, Base: USDT}
// 	if pair == BTCUSDT {
// 		return 3
// 	} else if pair == FTUSDT {
// 		return 2
// 	} else {
// 		return 1
// 	}

// }
