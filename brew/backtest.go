package brew

import (
	. "bean"
	"bean/exchange"
	"bean/rpc"
	"fmt"
	"github.com/wcharczuk/go-chart"
	"math"
	"net/http"
	"time"
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

	result := BackTestResult{start: start, end: end, pairs: pairs, dbhost: bt.dbhost, dbport: bt.dbport}
	for i, _ := range exNames {
		txn := exSims[i].GetTrades()
		result.Txn = append(result.Txn, txn...)
	}
	return result
}

// TODO: too ad-hoc, make it generic
func (res BackTestResult) Show() {
	p := NewPortfolio()
	snapts := GenerateSnapshotTS(res.Txn, p)
	mds := bean.NewRPCMDSConnC("tcp", res.dbhost+":"+res.dbport)
	ratesbook := make(ReferenceRateBook)

	// FIXME: think about how to show multi pair result
	if len(res.pairs) > 0 {
		p := res.pairs[0]
		txn, _ := mds.GetTransactions(p, res.start, res.end)
		ratesbook[p] = RefRatesFromTxn(txn)
		perfts := EvaluateSnapshotTS(snapts, p.Base, ratesbook)
		snapts.Print()
		perfts.Print()

		var pfst TradestatPort
		sta := pfst.GetPortStat(p.Base, res.Txn, NewPortfolio(), ratesbook)
		fmt.Println("all coin:", sta.AllCoins)
		fmt.Println("MaxDrawdown:", sta.MaxDrawdown)
		fmt.Println("NetPnL:", sta.NetPnL)
		// fmt.Println("AnnReturn:", sta.AnnReturn)
		// fmt.Println("Sharpe:", sta.Sharpe)
		fmt.Println("Win/Loss:", sta.WLRatio)
		fmt.Println("Win/NumofTrade:", sta.WinRate)
		fmt.Println("AvgWLRatio:", sta.AvgWLRatio)

		totalAmount := 0.0
		for _, tx := range res.Txn {
			totalAmount += math.Abs(tx.Amount)
		}
		fmt.Println("Total Transaction Amount:", totalAmount)
		fmt.Println("Final Portfolio:", snapts[len(snapts)-1])

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
