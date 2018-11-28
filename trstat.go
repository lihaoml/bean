package bean

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/gonum/stat"
	"gonum.org/v1/gonum/floats"
)

// stat for single coin
type TradestatCoin struct{}

type CoinPerformanceStat struct {
	NumofTr     int
	NetPnL      float64
	AvgPnL      float64
	AnnReturn   float64
	DrawdownTS  []float64
	MaxDrawdown float64
	Sharpe      float64
	AvgWLRatio  float64
	WinRate     float64
	LossRate    float64
	WLRatio     float64
}

//stat for whole portfolio
type TradestatPort struct {
	mtmBase   Coin
	ts        Transactions
	p         Portfolio
	ratesBook ReferenceRateBook
	permTS    PerformanceTS
}

type PortPerformanceStat struct {
	AllCoins    Coins
	NetPnL      float64
	AvgPnL      float64
	AnnReturn   float64
	DrawdownTS  []float64
	MaxDrawdown float64
	Sharpe      float64
	AvgWLRatio  float64
	WinRate     float64
	LossRate    float64
	WLRatio     float64
}

// get all stat for a single coin
func (tst TradestatCoin) GetCoinStat(coin Coin, mtmBase Coin, ratesbook ReferenceRateBook, ts Transactions, p Portfolio) CoinPerformanceStat {
	var coinstat CoinPerformanceStat
	coinstat.NumofTr = tst.GetTrNumber(coin, ts)
	coinstat.NetPnL, _ = tst.GetNetPnL(coin, mtmBase, ratesbook, ts, p)
	coinstat.AvgPnL = tst.GetAvgPnL(coin, mtmBase, ratesbook, ts, p)
	coinstat.AnnReturn, _ = tst.GetAnnReturn(coin, mtmBase, ratesbook, ts, p)
	coinstat.DrawdownTS, _ = tst.GetMaxDrawdown(coin, mtmBase, ratesbook, ts, p)
	_, coinstat.MaxDrawdown = tst.GetMaxDrawdown(coin, mtmBase, ratesbook, ts, p)
	coinstat.Sharpe = tst.GetSharpe(coin, mtmBase, ratesbook, ts, p)
	coinstat.AvgWLRatio, _, _, _ = tst.GetWLRatio(coin, mtmBase, ratesbook, ts, p)
	_, coinstat.WinRate, _, _ = tst.GetWLRatio(coin, mtmBase, ratesbook, ts, p)
	_, _, coinstat.LossRate, _ = tst.GetWLRatio(coin, mtmBase, ratesbook, ts, p)
	_, _, _, coinstat.WLRatio = tst.GetWLRatio(coin, mtmBase, ratesbook, ts, p)

	return coinstat
}

// get all stat for portfolio
func (pfst TradestatPort) PortStat() PortPerformanceStat {
	var portstat PortPerformanceStat
	portstat.AllCoins = pfst.AllCoin()
	portstat.NetPnL = pfst.NetPnL()
	portstat.AvgPnL = pfst.AveragePnL()
	_, portstat.AnnReturn = pfst.AnnReturn()
	portstat.DrawdownTS, portstat.MaxDrawdown = pfst.MaxDrawdown()
	portstat.Sharpe = pfst.Sharpe()
	portstat.AvgWLRatio, portstat.WinRate, portstat.LossRate, portstat.WLRatio = pfst.WLRatio()
	return portstat
}

////////////////////////////////// functions for portfolio /////////////////////////////////

// list all the coins within trading
func (pfst TradestatPort) AllCoin() Coins {
	var coins Coins
	for _, v := range pfst.ts {
		indC := 0
		indB := 0
		for _, coin := range coins {
			if coin != v.Pair.Coin {
				indC += 0
			} else {
				indC += 1
			}
			if coin != v.Pair.Base {
				indB += 0
			} else {
				indB += 1
			}
		}
		if indC == 0 {
			coins = append(coins, v.Pair.Coin)
		}
		if indB == 0 {
			coins = append(coins, v.Pair.Base)
		}
	}
	return coins
}

func Tradestat(mtmbase Coin, ts Transactions, p Portfolio, ratesbook ReferenceRateBook) *TradestatPort {
	tradestat := TradestatPort{
		mtmBase:   mtmbase,
		ts:        ts,
		p:         p,
		ratesBook: ratesbook,
	}
	ssTS := GenerateSnapshotTS(ts, p)
	tradestat.permTS = EvaluateSnapshotTS(ssTS, mtmbase, ratesbook)
	return &tradestat
}

func (pfst TradestatPort) Print() {
	fmt.Println("all coin:", pfst.AllCoin())
	_, maxdd := pfst.MaxDrawdown()
	fmt.Println("MaxDrawdown:", maxdd)
	fmt.Println("NetPnL:", pfst.NetPnL())
	// fmt.Println("AnnReturn:", sta.AnnReturn)
	// fmt.Println("Sharpe:", sta.Sharpe)
	AvgWL, WR, _, WL := pfst.WLRatio()
	fmt.Println("Win/Loss:", WL)
	fmt.Println("Win/NumofTrade:", WR)
	fmt.Println("AvgWLRatio:", AvgWL)
	fmt.Println("AverageWin:", pfst.AveragePnL())
	fmt.Println("Total Transaction Amount:", pfst.TotalTransactionAmount())
}

// get net PnL (endPV - initPV)
func (pfst TradestatPort) NetPnL() float64 {
	var netPnL float64
	for _, v := range pfst.permTS {
		netPnL += v.PnL
	}
	return netPnL
}

func (pfst TradestatPort) TotalTransactionAmount() float64 {
	totalAmount := 0.0
	for _, tx := range pfst.ts {
		totalAmount += math.Abs(tx.Amount)
	}
	return totalAmount
}

// get AvgPnL (netPnL / numofTr)
func (pfst TradestatPort) AveragePnL() float64 {
	netPnL := pfst.NetPnL()
	return netPnL / float64(len(pfst.permTS))
}

// get AnnReturn (ln return)
func (pfst TradestatPort) AnnReturn() (RtnTS []float64, annrtn float64) {
	var returnTS []float64
	var PV []float64
	for i, v := range pfst.permTS {
		if i == 0 {
			PV = append(PV, v.PV)
		} else {
			PV = append(PV, v.PV)
			returnTS = append(returnTS, math.Log(PV[i]/PV[i-1]))
		}
	}
	var rtnTS []float64
	if PV[0] == 0 {
		for i, _ := range PV {
			PV[i] = PV[i] + 2e-8
			if i > 0 {
				if PV[i]*PV[i-1] > 0 {
					if PV[i]-PV[i-1] > 0 {
						rtnTS = append(rtnTS, math.Log(PV[i]/PV[i-1]))
					} else {
						rtnTS = append(rtnTS, -1*math.Log(PV[i]/PV[i-1]))
					}
				} else {
					if PV[i-1] > 0 {
						rtnTS = append(rtnTS, -1*math.Log((math.Abs(PV[i])+2*PV[i-1])/PV[i-1]))
					} else {
						rtnTS = append(rtnTS, math.Log((PV[i]+2*math.Abs(PV[i-1]))/math.Abs(PV[i-1])))
					}
				}
			}
		}
		tmperiod := (pfst.permTS[len(pfst.permTS)-1].Time.Sub(pfst.permTS[0].Time)).Seconds() / (24 * 60 * 60)
		annRtn := floats.Sum(rtnTS) / (tmperiod / 365)
		return rtnTS, annRtn
	} else {
		tmperiod := (pfst.permTS[len(pfst.permTS)-1].Time.Sub(pfst.permTS[0].Time)).Seconds() / (24 * 60 * 60)
		annReturn := floats.Sum(returnTS) / (tmperiod / 365)
		return returnTS, annReturn
	}
}

// get Drawdown series and MaxDrawdown
func (pfst TradestatPort) MaxDrawdown() (DD []float64, MaxDD float64) {
	var drawdown []float64
	var maxsofar float64
	for i, v := range pfst.permTS {
		if i == 0 {
			maxsofar = v.PV
			drawdown = append(drawdown, 0)
		}
		if v.PV > maxsofar {
			drawdown = append(drawdown, maxsofar-v.PV)
			maxsofar = v.PV
		} else {
			drawdown = append(drawdown, maxsofar-v.PV)
		}
	}
	cloneDD := drawdown
	sort.Float64s(cloneDD)
	// find the max of DD series
	return drawdown, cloneDD[len(drawdown)-1]
}

// get Sharpe Ratio (use ln return)
func (pfst TradestatPort) Sharpe() float64 {
	// annualized return and annualized volatility
	rtnTS, annreturn := pfst.AnnReturn()
	stddev := stat.StdDev(rtnTS, nil)
	annvol := stddev * math.Sqrt(365)
	return (annreturn - 0.02) / annvol // here, set risk free rate as 2%
}

// get AvgWinLoss ratio; Win rate; Loss rate; WL ratio
func (pfst TradestatPort) WLRatio() (AvgWinLoss, WR, LR, WL float64) {
	var PV []float64     // record PV series
	winNum := float64(0) // record number of win transaction
	lossNum := float64(0)
	var winAmount float64 // record total win amount
	var lossAmount float64
	for i, v := range pfst.permTS {
		if i == 0 {
			PV = append(PV, v.PV)
		} else {
			PV = append(PV, v.PV)
			// find PnL
			change := PV[i] - PV[i-1]
			if change > 0 {
				winNum += 1
				winAmount += change
			} else if change < 0 {
				lossNum += 1
				lossAmount += change
			}
		}
	}
	return (winAmount / winNum) / (lossAmount / lossNum), (winNum / float64(len(pfst.permTS))), (lossNum / float64(len(pfst.permTS))), (winNum / lossNum)

}

////////////////////////////////// functions for single coin performance /////////////////////////////////

// get the trading frequency for a specific coin
func (tst TradestatCoin) GetTrNumber(coin Coin, txn Transactions) int {
	count := 0
	for _, v := range txn {
		if v.Pair.Coin == coin || v.Pair.Base == coin {
			count += 1
		}
	}
	return count
}

// get the net PnL for a specific coin with respect to mtmbase
type CoinPV struct {
	Time time.Time
	PV   float64
}

func (tst TradestatCoin) GetNetPnL(coin Coin, mtmBase Coin, ratesbook ReferenceRateBook, ts Transactions, p Portfolio) (NetPnL float64, CoinpvTS []CoinPV) {
	ssTS := GenerateSnapshotTS(ts, p)
	var coinpv CoinPV
	var coinpvTS []CoinPV
	for _, v := range ssTS {
		if v.Port.Balance(coin) != 0 {
			rate := LookupRate(Pair{Coin: coin, Base: mtmBase}, v.Time, ratesbook)
			coinpv.Time = v.Time
			coinpv.PV = rate * v.Port.Balance(coin)
			coinpvTS = append(coinpvTS, coinpv)
		}
	}
	return coinpvTS[len(coinpvTS)-1].PV - coinpvTS[0].PV, coinpvTS
}

// get the average PnL for a specific coin with respect to mtmbase
func (tst TradestatCoin) GetAvgPnL(coin Coin, mtmBase Coin, ratesbook ReferenceRateBook, ts Transactions, p Portfolio) float64 {
	netPnL, _ := tst.GetNetPnL(coin, mtmBase, ratesbook, ts, p)
	num := tst.GetTrNumber(coin, ts)
	return netPnL / float64(num)
}

// get AnnReturn with respect to mtmbase
func (tst TradestatCoin) GetAnnReturn(coin Coin, mtmBase Coin, ratesbook ReferenceRateBook, ts Transactions, p Portfolio) (annRtn float64, RtnTS []float64) {
	_, coinpvTS := tst.GetNetPnL(coin, mtmBase, ratesbook, ts, p)
	var returnTS []float64
	for i, _ := range coinpvTS {
		if i > 0 {
			returnTS = append(returnTS, math.Log(coinpvTS[i].PV/coinpvTS[i-1].PV))
		}
	}
	tmperiod := coinpvTS[len(coinpvTS)-1].Time.Sub(coinpvTS[0].Time).Seconds() / (24 * 60 * 60)
	return floats.Sum(returnTS) / (tmperiod / 365), returnTS
}

// get Drawdown series and MaxDrawdown
func (tst TradestatCoin) GetMaxDrawdown(coin Coin, mtmBase Coin, ratesbook ReferenceRateBook, ts Transactions, p Portfolio) (DD []float64, MaxDD float64) {
	_, coinpvTS := tst.GetNetPnL(coin, mtmBase, ratesbook, ts, p)
	// find the drawdown series
	var drawdown []float64
	var maxsofar float64
	for i, v := range coinpvTS {
		if i == 0 {
			maxsofar = v.PV
			drawdown = append(drawdown, 0)
		}
		if v.PV > maxsofar {
			drawdown = append(drawdown, 1-v.PV/maxsofar)
			maxsofar = v.PV
		} else {
			drawdown = append(drawdown, 1-v.PV/maxsofar)
		}
	}
	cloneDD := drawdown
	sort.Float64s(cloneDD)
	// find the max of DD series
	return drawdown, cloneDD[len(drawdown)-1]
}

// get AnnSharpe Ratio
func (tst TradestatCoin) GetSharpe(coin Coin, mtmBase Coin, ratesbook ReferenceRateBook, ts Transactions, p Portfolio) float64 {
	// annualized return and annualized volatility
	annurtn, rtnTS := tst.GetAnnReturn(coin, mtmBase, ratesbook, ts, p)
	stddev := stat.StdDev(rtnTS, nil)
	annvol := stddev * math.Sqrt(365)
	return (annurtn - 0.02) / annvol
}

// get AvgWinLoss ratio; Win rate; Loss rate; WL ratio
func (tst TradestatCoin) GetWLRatio(coin Coin, mtmBase Coin, ratesbook ReferenceRateBook, ts Transactions, p Portfolio) (AvgWinLoss, WR, LR, WL float64) {
	// find return series
	_, pvTS := tst.GetNetPnL(coin, mtmBase, ratesbook, ts, p)
	winNum := float64(0)
	lossNum := float64(0)
	var winAmount float64
	var lossAmount float64
	for i, _ := range pvTS {
		for i > 0 {
			change := pvTS[i].PV - pvTS[i-1].PV
			if change > 0 {
				winNum += 1
				winAmount += change
			} else if change < 0 {
				lossNum += 1
				lossAmount += change
			}
		}
	}
	trnum := tst.GetTrNumber(coin, ts)
	return (winAmount / winNum) / (lossAmount / lossNum), (winNum / float64(trnum)), (lossNum / float64(trnum)), (winNum / lossNum)
}
