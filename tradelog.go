package bean

import (
	util "bean/utils"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

// Define my trade
type TradeLog struct {
	OrderID         string
	Pair            Pair
	Price           float64
	Quantity        float64
	Commission      float64
	CommissionAsset Coin
	Time            time.Time
	Side            Side
	TxnID           string // if any
}

type TradeLogSummary struct {
	Pair       Pair
	SellAmount float64
	SellValue  float64
	BuyAmount  float64
	BuyValue   float64
	Fee        map[Coin]float64
}

func (tls TradeLogSummary) AvgBuyPrice() float64 {
	return tls.BuyValue / tls.BuyAmount
}

func (tls TradeLogSummary) AvgSellPrice() float64 {
	return tls.SellValue / tls.SellAmount
}

func (tls TradeLogSummary) NetExposure() float64 {
	return tls.BuyAmount - tls.SellAmount
}

func (tls TradeLogSummary) RealizedPL() float64 {
	return (tls.AvgSellPrice() - tls.AvgBuyPrice()) * math.Min(tls.BuyAmount, tls.SellAmount)
}

func (tls TradeLogSummary) UnrealizedPL(mid float64) float64 {
	exposure := tls.NetExposure()
	if exposure < 0 {
		return exposure * (mid - tls.AvgSellPrice())
	} else {
		return exposure * (mid - tls.AvgBuyPrice())
	}
}

func (tls TradeLogSummary) AvgCost() float64 {
	return (tls.SellValue - tls.BuyValue) / (tls.BuyAmount - tls.SellAmount)
}

type TradeLogS []TradeLog

func (t TradeLogS) Sort() TradeLogS {
	sort.Slice(t, func(i, j int) bool { return t[i].Time.Before(t[j].Time) })
	return t
}

// requires a single pair
func (trds TradeLogS) PrintStats(p Pair) {
	pv := make([]float64, 1)
	pl := make([]float64, 1)
	baseAmt := 0.0
	assetAmt := 0.0
	for _, v := range trds {
		if v.Pair == p {
			sign := 1.0
			if v.Side == "SELL" {
				sign = -1.0
			}
			baseAmt = baseAmt - v.Price*v.Quantity*sign
			assetAmt = assetAmt + v.Quantity*sign
			if v.CommissionAsset == v.Pair.Base {
				baseAmt = baseAmt - v.Commission
			} else {
				assetAmt = assetAmt - v.Commission
			}
			pl = append(pl, baseAmt+assetAmt*v.Price-pv[len(pv)-1])
			pv = append(pv, baseAmt+assetAmt*v.Price)
		}
	}
	maxdd := MaxDD(pv)
	fmt.Println("MaxDD:\t", maxdd)
	fmt.Println("PL:\t", pv[len(pl)-1])
}

func (trds TradeLogS) ToCSV(filename string) {
	csvFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()
	var data [][]string

	head := []string{
		"Time",
		"Pair",
		"Price",
		"Quantity",
		"Commission",
		"CommissionAsset",
		"Side",
		"BASE",
		"ASSET",
		"PV",
		"PnL",
	}
	data = append(data, head)
	baseAmt := 0.0
	assetAmt := 0.0
	pv := 0.0
	pl := 0.0
	for _, v := range trds {
		sign := 1.0
		if v.Side == "SELL" {
			sign = -1.0
		}
		baseAmt = baseAmt - v.Price*v.Quantity*sign
		assetAmt = assetAmt + v.Quantity*sign
		if v.CommissionAsset == v.Pair.Base {
			baseAmt = baseAmt - v.Commission
		} else {
			assetAmt = assetAmt - v.Commission
		}
		pl = baseAmt + assetAmt*v.Price - pv
		pv = baseAmt + assetAmt*v.Price
		s := []string{
			v.Time.Format(time.RFC3339),
			v.Pair.String(),
			fmt.Sprint(v.Price),
			fmt.Sprint(v.Quantity),
			fmt.Sprint(v.Commission),
			string(v.CommissionAsset),
			string(v.Side),
			fmt.Sprint(baseAmt),
			fmt.Sprint(assetAmt),
			fmt.Sprint(pv),
			fmt.Sprint(pl),
		}
		data = append(data, s)
	}
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.WriteAll(data)
	csvWriter.Flush()
}

func (trades TradeLogS) Summary(pair Pair) (tradesummary TradeLogSummary) {
	sellAmount := 0.0
	sellValue := 0.0
	buyAmount := 0.0
	buyValue := 0.0
	fee := make(map[Coin]float64)
	for _, v := range trades {
		if v.Pair == pair {
			if v.Side == BUY {
				buyAmount += v.Quantity
				buyValue += v.Quantity * v.Price
			} else if v.Side == SELL {
				sellAmount += v.Quantity
				sellValue += v.Quantity * v.Price
			}
			fee[v.CommissionAsset] += v.Commission
		}
	}
	tradesummary.Pair = pair
	tradesummary.BuyValue = buyValue
	tradesummary.BuyAmount = buyAmount
	tradesummary.SellValue = sellValue
	tradesummary.SellAmount = sellAmount
	tradesummary.Fee = fee
	return
}

func (trades TradeLogS) Net() Portfolio {
	port := NewPortfolio()
	for _, t := range trades {
		sign := 1.0
		if t.Side == SELL {
			sign = -1.0
		}
		port.AddBalance(t.Pair.Coin, t.Quantity*sign)
		port.AddBalance(t.Pair.Base, t.Quantity*sign*-1*t.Price)
		port.AddBalance(t.CommissionAsset, t.Commission*-1)
	}
	return port
}

// res = trd1 - trd2
func (trds1 TradeLogS) Minus(trds2 TradeLogS) (res TradeLogS) {
	// force alignment of time
	for i, v := range trds2 {
		trds2[i].Time = time.Unix(v.Time.Unix(), 0)
	}
	for _, v := range trds1 {
		v.Time = time.Unix(v.Time.Unix(), 0)
		if util.Contains(trds2, v) {
			continue
		} else {
			res = append(res, v)
		}
	}
	return
}

func (trades TradeLogS) ToTransactions() (txns Transactions) {
	for _, trd := range trades {
		sign := 1.0
		maker := Buyer
		if trd.Side == SELL {
			sign = -1.0
			maker = Seller
		}
		txn := Transaction{
			Pair:      trd.Pair,
			Price:     trd.Price,
			Amount:    math.Abs(trd.Quantity) * sign,
			TimeStamp: trd.Time,
			Maker:     maker,
			TxnID:     trd.OrderID,
		}
		txns = append(txns, txn)
	}
	return txns
}

func (trades TradeLogS) Since(t time.Time) (position Portfolio, after TradeLogS) {
	var before TradeLogS
	for _, trd := range trades {
		if trd.Time.Before(t) {
			before = append(before, trd)
		} else {
			after = append(after, trd)
		}
	}
	snapts := GenerateSnapshotTS(before.ToTransactions(), NewPortfolio())
	if len(snapts) > 0 {
		position = snapts[len(snapts)-1].Port
	}
	return position, after
}

func (tls TradeLogS) Pairs() []Pair {
	pairs := []Pair{}
	for _, t := range tls {
		if !util.Contains(pairs, t.Pair) {
			pairs = append(pairs, t.Pair)
		}
	}
	return pairs
}
