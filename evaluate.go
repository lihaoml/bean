package bean

import (
	"fmt"
	"github.com/gonum/floats"
	"math"
	"sort"
	"time"
)

func MaxDD(pv []float64) float64 {
	var drawdown []float64
	var maxsofar float64
	for i, v := range pv {
		if i == 0 || v > maxsofar {
			maxsofar = v
		}
		drawdown = append(drawdown, maxsofar-v)
	}
	return floats.Max(drawdown)
}

//Snapshot represents the state of a portfolio at a specific time
type Snapshot struct {
	Time time.Time
	Port Portfolio
}

//SnapshotTS represents the states of a portfolio over time
type SnapshotTS []Snapshot

//Print the contents of the SnapshotTS
func (s SnapshotTS) Print() {
	fmt.Println("SnapshotTS")
	for _, v := range s {
		fmt.Printf("  Snapshot %v: ", v.Time.Local())
		balances := v.Port.Balances()
		for k, e := range balances {
			fmt.Printf("%v,%v  ", k, e)
		}
		fmt.Println("")
	}
}

// minus initial portfolio to get PnL portfolio
func (s SnapshotTS) Minus(initPort Portfolio) SnapshotTS {
	newTS := SnapshotTS{}
	for _, v := range s {
		newTS = append(newTS, Snapshot{v.Time, v.Port.Subtract(initPort)})
	}
	return newTS
}

func (s SnapshotTS) Times() []time.Time {
	ts := []time.Time{}
	for _, v := range s {
		ts = append(ts, v.Time)
	}
	return ts
}

//Performance measures the value of a portfolio at a specific time
type Performance struct {
	Time    time.Time
	MtMBase Coin
	PV      float64
	PnL     float64
}

//PerformanceTS measures the value of a portoflio over time
type PerformanceTS []Performance

//Print the contents of PerformanceTS
func (ps PerformanceTS) Print() {
	fmt.Println("PerformanceTS")
	for _, p := range ps {
		fmt.Printf("  %v:  MtMBase %v  PV %v  PnL %v\n", p.Time, p.MtMBase, p.PV, p.PnL)
	}
}

func (ps PerformanceTS) PnLSince(t time.Time) float64 {
	if len(ps) == 0 {
		return 0.0
	}
	lastPV := ps[len(ps)-1].PV

	idx := 0
	minGap := math.Abs(float64(ps[0].Time.Sub(t)))

	fmt.Println(ps[0].Time)
	for i, p := range ps {
		if math.Abs(float64(p.Time.Sub(t))) < minGap {
			idx = i
			minGap = math.Abs(float64(p.Time.Sub(t)))
		}
	}
	return lastPV - ps[idx].PV
}

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

//Print the contents of ReferenceRateBook
func (t ReferenceRateBook) Print() {
	for k, v := range t {
		fmt.Println("ReferenceRateBook")
		fmt.Printf("  ReferenceRate %v: \n", k)
		for _, elm := range v {
			fmt.Printf("    %v: Price %v\n", elm.Time, elm.Price)
		}
	}
}

//GenerateSnapshotTS update portfolio overtime
func GenerateSnapshotTS(ts Transactions, initPort Portfolio) SnapshotTS {
	ts.Sort()
	var snapts SnapshotTS
	p := initPort
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
		}
		if rate != 0 {
			pv += v * rate
		} else {
			panic(fmt.Sprint("%v%v at %v is not availabe\n", k, mtmBase, snap.Time))
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
			perf.PnL = perf.PV - perfts[i-1].PV
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
