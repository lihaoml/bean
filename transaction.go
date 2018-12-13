package bean

import (
	"encoding/csv"
	"math"
	"os"
	"sort"
	"strconv"
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
	Pair            string
	Price           string
	Quantity        string
	Commission      string
	CommissionAsset string
	Time            string
	Side            string
}

type TradeLogSummary struct {
	Pair         Pair
	SellAmount   float64
	SellValue    float64
	AvgSellPrice float64
	BuyAmount    float64
	BuyValue     float64
	AvgBuyPrice  float64
	AvgCost      float64
	Fee          map[string]float64
}

type TradeLogS []TradeLog

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
			v.Time,
			v.Pair,
			v.Price,
			v.Quantity,
			v.Commission,
			v.CommissionAsset,
			v.Side,
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
	fee := make(map[string]float64, 3)
	for _, v := range trades {
		if v.Pair == string(pair.Coin)+string(pair.Base) {
			if v.Side == "BUY" {
				buy, _ := strconv.ParseFloat(v.Quantity, 64)
				buyAmount += buy
				price, _ := strconv.ParseFloat(v.Price, 64)
				buyValue += buy * price
			} else {
				sell, _ := strconv.ParseFloat(v.Quantity, 64)
				sellAmount += sell
				price, _ := strconv.ParseFloat(v.Price, 64)
				sellValue += sell * price
			}
			comisn, _ := strconv.ParseFloat(v.Commission, 64)
			fee[v.CommissionAsset] += comisn
		}
	}
	tradesummary.Pair = pair
	tradesummary.BuyValue = buyValue
	tradesummary.BuyAmount = buyAmount
	tradesummary.AvgBuyPrice = (buyValue / buyAmount)
	tradesummary.SellValue = sellValue
	tradesummary.SellAmount = sellAmount
	tradesummary.AvgSellPrice = (sellValue / sellAmount)
	tradesummary.Fee = fee
	tradesummary.AvgCost = (sellValue - buyValue) / (buyAmount - sellAmount)
	return
}
