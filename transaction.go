package bean

import (
	"encoding/csv"
	"errors"
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

type OHLCVBS struct {
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
	BuyVolume  float64
	SellVolume float64
	Start      time.Time
	End        time.Time
}

type OHLCVBSTS []OHLCVBS


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

// assuming transactions is sorted
func (t Transactions) OHLCVBS() (OHLCVBS, error) {
	var res OHLCVBS
	var err error
	if len(t) > 0 {
		res.Start = t[0].TimeStamp
		res.End = t[len(t)-1].TimeStamp
		res.Open = t[0].Price
		res.Close = t[len(t)-1].Price
		res.High = res.Open
		res.Low = res.Open
		res.Volume = 0
		res.BuyVolume = 0
		res.SellVolume = 0
		for _, txn := range t {
			if txn.Price > res.High {
				res.High = txn.Price
			}
			if txn.Price < res.Low {
				res.Low = txn.Price
			}
			res.Volume += math.Abs(txn.Amount) * txn.Price
			if txn.Maker == Buyer {
				res.SellVolume += math.Abs(txn.Amount) * txn.Price
			} else {
				res.BuyVolume += math.Abs(txn.Amount) * txn.Price
			}
		}
	} else {
		err = errors.New("OHLCVBS: empty transactions")
	}
	return res, err
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
		if tt.TimeStamp.After(from) {
			startIdx = i
			break
		}
	}
	if startIdx < len(txn) {
		endIdx := startIdx
		for i := startIdx; i < len(txn); i++ {
			if !txn[i].TimeStamp.After(to) {
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
		"Maker",
	}
	data = append(data, head)

	for _, v := range txn {
		s := []string{
			v.TimeStamp.Format(time.RFC3339),
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
