package brew

import (
	. "bean"
	"bean/exchange"
	"bean/rpc"
	util "bean/utils"
	"fmt"
	"net/http"
	"time"

	"beanex/db/mds"
	"github.com/wcharczuk/go-chart"
)

// BackTest data type
type BackTest struct {
	dbhost string
	dbport string
}

type BackTestResult struct {
	Txn            []Transaction
	Orders         []OrderBookTS // my order book at any point in time
	start, end     time.Time
	pairs          []Pair
	dbhost, dbport string
}

func NewBackTest(dbhost, dbport string) BackTest {
	return BackTest{
		dbhost: dbhost,
		dbport: dbport,
	}
}

func (bt BackTest) Simulate(strat Strat, start, end time.Time, initPort Portfolio) BackTestResult {
	// fetch historical data ///////////////////////////////////////////
	exNames := strat.GetExchangeNames()
	pairs := strat.GetPairs()
	exSims := make([]exchange.Simulator, len(exNames))
	exs := make(map[string]Exchange)
	for i, exName := range exNames {
		exSims[i] = exchange.NewSimulator(exName, pairs, bt.dbhost, bt.dbport, start, end, initPort)
		exs[exName] = &exSims[i]
	}
	fmt.Println("ex constructed")

	tick := strat.GetTick()
	// from start to end, call strat's Work
	for t := start; t.Before(end); t = t.Add(tick) {
		// update now in exSIm
		for i, _ := range exNames {
			exSims[i].SetTime(t)
		}
		actions := strat.Grind(exs)
		// Perform actions
		PerformActions(&exs, actions)
	}

	fmt.Println("done simulation")
	result := BackTestResult{start: start, end: end, pairs: pairs, dbhost: bt.dbhost, dbport: bt.dbport}
	for i, _ := range exNames {
		txn := exSims[i].GetTrades()
		result.Txn = append(result.Txn, txn...)
	}
	return result
}

func (bt BackTest) SimulateN(strats []Strat, start, end time.Time, initPort Portfolio) []BackTestResult {
	// fetch historical data ///////////////////////////////////////////
	var exNames []string
	pairs := make(map[string]([]Pair))
	for _, s := range strats {
		exns := s.GetExchangeNames()
		ps := s.GetPairs()
		for _, n := range exns {
			if !util.Contains(exNames, n) {
				exNames = append(exNames, n)
			}
			for _, p := range ps {
				if !util.Contains(pairs[n], p) {
					pairs[n] = append(pairs[n], p)
				}
			}
		}
	}
	fmt.Println("exNames: ", exNames)
	fmt.Println("pairs: ", pairs)
	exSims := make([]exchange.Simulator, len(exNames))
	exs := make(map[string]Exchange)
	for i, exName := range exNames {
		exSims[i] = exchange.NewSimulator(exName, pairs[exName], bt.dbhost, bt.dbport, start, end, initPort)
		exs[exName] = &exSims[i]
	}
	fmt.Println("ex constructed")
	// now we can simulate each strategy
	result := make([]BackTestResult, len(strats))
	for k, strat := range strats {
		tick := strat.GetTick()
		// from start to end, call strat's Work
		for t := start; t.Before(end); t = t.Add(tick) {
			// update now in exSIm
			for i, _ := range exNames {
				exSims[i].SetTime(t)
			}
			actions := strat.Grind(exs)
			// Perform actions
			PerformActions(&exs, actions)
		}

		result[k] = BackTestResult{start: start, end: end, pairs: strat.GetPairs(), dbhost: bt.dbhost, dbport: bt.dbport}
		for nm, _ := range exNames {
			txn := exSims[nm].GetTrades()
			result[k].Txn = append(result[k].Txn, txn...)
		}

		for i, _ := range exs {
			exs[i].(*exchange.Simulator).Reset(start, initPort)
		}

		fmt.Println("done simulation for ", strat.Name(), strat.FormatParams(), len(result[k].Txn))
	}
	return result
}

// TODO: too ad-hoc, make it generic
func (res BackTestResult) Show() TradestatPort {
	//	p := NewPortfolio()
	// mds := bean.NewRPCMDSConnC("tcp", res.dbhost+":"+res.dbport)
	ratesbook := make(ReferenceRateBook)

	// FIXME: think about how to show multi pair result
	var stat TradestatPort
	if len(res.pairs) > 0 {
		p := res.pairs[0]
		txn, _ := mds.GetTransactions2(NameFcoin, p, res.start, res.end)
		ratesbook[p] = RefRatesFromTxn(txn)
		//		snapts.Print()
		//		perfts.Print()

		stat = *Tradestat(p.Base, res.Txn, NewPortfolio(), ratesbook)
		stat.Print()
	}
	return stat
}

func (res BackTestResult) Graph() {
	p := NewPortfolio()
	snapts := GenerateSnapshotTS(res.Txn, p)
	ratesbook := make(ReferenceRateBook)
	perfts := EvaluateSnapshotTS(snapts, res.pairs[0].Base, ratesbook)

	xs := make([]time.Time, len(perfts))
	ys := make([]float64, len(perfts))
	for i, v := range perfts {
		xs[i] = v.Time
		ys[i] = v.PV
	}
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeHourValueFormatter,
			TickPosition:   chart.TickPositionBetweenTicks,
		},
		YAxis: chart.YAxis{
			Style: chart.StyleShow(),
			/*
				Range: &chart.ContinuousRange{
					Max: 5,
					Min: -5,
				},
			*/
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: xs,
				YValues: ys,
			},
		},
	}

	http.HandleFunc("/", func(r http.ResponseWriter, req *http.Request) {
		r.Header().Set("Content-Type", "image/png")
		graph.Render(chart.PNG, r)
	})
	http.ListenAndServe(":8080", nil)
}

//Evaluate shows the performance for a backtest marked to mtmBase
func (res BackTestResult) Evaluate(mtmBase Coin) {
	p := NewPortfolio()
	snapts := GenerateSnapshotTS(res.Txn, p)
	mds := bean.NewRPCMDSConnC("tcp", res.dbhost+":"+res.dbport)
	ratesbook := make(ReferenceRateBook)

	for _, p := range res.pairs {
		if p.Coin != mtmBase {
			tp := Pair{Coin: p.Coin, Base: mtmBase}
			txn, _ := mds.GetTransactions(tp, res.start, res.end)
			ratesbook[tp] = RefRatesFromTxn(txn)
		}
		if p.Base != mtmBase {
			tp := Pair{Coin: p.Base, Base: mtmBase}
			txn, _ := mds.GetTransactions(tp, res.start, res.end)
			ratesbook[tp] = RefRatesFromTxn(txn)
		}
		//ratesbook.Print()
		perfts := EvaluateSnapshotTS(snapts, mtmBase, ratesbook)
		perfts.Print()
	}
}

// quick evaluation for single pair transactions
func (res BackTestResult) QuickEval() TradestatPort {
	if len(res.pairs) == 1 {
		ratesbook := make(ReferenceRateBook)
		ratesbook[res.pairs[0]] = RefRatesFromTxn(res.Txn)
		stats := *Tradestat(res.pairs[0].Base, res.Txn, NewPortfolio(), ratesbook)
		return stats
	} else {
		panic("QuickEval supports single pair only")
	}

}
