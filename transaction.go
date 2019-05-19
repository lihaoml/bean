package bean

import (
	"bean/utils"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

type TraderType int

const (
	Buyer  TraderType = 0
	Seller TraderType = 1
)

type Transaction struct {
	Pair      Pair
	Price     float64
	Amount    float64
	TimeStamp time.Time
	Maker     TraderType // buyer or seller
	TxnID     string
}

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
	TxnID           string
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

// To be completed
type Transactions []Transaction

func (t Transactions) IsValid() bool {
	return len(t) > 0
}

// this function assume all transactions belong to the same pair
func (t Transactions) Volume(pair Pair) (float64, float64) {
	volCoin := 0.0
	volBase := 0.0
	for _, txn := range t {
		if pair == txn.Pair {
			volCoin += math.Abs(txn.Amount)
			volBase += math.Abs(txn.Amount * txn.Price)
		}
	}
	return volCoin, volBase
}

func (t Transactions) OHLC() {
}

func (t Transactions) Sort() Transactions {
	sort.Slice(t, func(i, j int) bool { return t[i].TimeStamp.Before(t[j].TimeStamp) })
	return t
}

// get transactions up to t, assuming txn is sorted
func (txn Transactions) Upto(t time.Time) Transactions {
	idx := 0
	for i, tt := range txn {
		if tt.TimeStamp.Before(t) {
			idx = i
		} else {
			break
		}
	}
	res := txn[0 : idx+1]
	return res
}

func (txn Transactions) Since(t time.Time) Transactions {
	idx := 0
	for i, tt := range txn {
		if !tt.TimeStamp.Before(t) {
			idx = i
			break
		}
	}
	res := txn[idx:]
	return res
}

// get transactions in a time interval, assuming txn is sorted
func (txn Transactions) Between(from, to time.Time) Transactions {
	var res Transactions
	startIdx := len(txn)
	for i, tt := range txn {
		if !tt.TimeStamp.Before(from) {
			startIdx = i
			break
		}
	}
	if startIdx < len(txn) {
		endIdx := startIdx
		for i := startIdx; i < len(txn); i++ {
			if txn[i].TimeStamp.Before(to) {
				endIdx = i
			} else {
				break
			}
		}
		res = txn[startIdx : endIdx+1]
	}
	return res
}

func (txn Transactions) Cross(price, amount float64) bool {
	if amount < 0 {
		// selling, so need to check if the highest transaction is larger than the order price
		for _, t := range txn {
			if t.Price > price*1.001 {
				return true
			}
		}
	} else {
		// buying, so need to check if the lowest transaction is lower than the order price
		for _, t := range txn {
			if t.Price < price*0.999 {
				return true
			}
		}
	}
	return false
}

func (txn Transactions) Fill(price, orderAmount float64) float64 {
	fillAmount := 0.0
	if orderAmount < 0 {
		// selling - add all transactions above the order price. fill is negative
		for _, t := range txn {
			if t.Price > price {
				fillAmount -= t.Amount
			}
		}
		return math.Max(orderAmount, fillAmount)
	} else {
		// buying - add all transactions below the order price. fill is positive
		for _, t := range txn {
			if t.Price < price {
				fillAmount += t.Amount
			}
		}
		return math.Min(orderAmount, fillAmount)
	}
}

func (txn Transactions) ToCSV(pair Pair, filename string) {
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
	}
	data = append(data, head)

	for _, v := range txn {
		s := []string{
			fmt.Sprint(v.TimeStamp),
			v.Pair.String(),
			fmt.Sprint(v.Price),
			fmt.Sprint(v.Amount),
			fmt.Sprint(v.Maker),
		}
		data = append(data, s)
	}
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.WriteAll(data)
	csvWriter.Flush()
}

func (trades TradeLogS) ToCSV(pair Pair, filename string) {
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
	}
	data = append(data, head)

	for _, v := range trades {
		s := []string{
			fmt.Sprint(v.Time),
			v.Pair.String(),
			fmt.Sprint(v.Price),
			fmt.Sprint(v.Quantity),
			fmt.Sprint(v.Commission),
			string(v.CommissionAsset),
			string(v.Side),
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

func (trades TradeLogS) toTransactions() (txns Transactions) {
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
	snapts := GenerateSnapshotTS(before.toTransactions(), NewPortfolio())
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
