package bean

import (
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

// To be completed
type Transactions []Transaction

func (t Transactions) IsValid() bool {
	return len(t) > 0
}

func (t Transactions) Volume() {
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
			if t.Price > price {
				return true
			}
		}
	} else {
		// buying, so need to check if the lowest transaction is lower than the order price
		for _, t := range txn {
			if t.Price < price {
				return true
			}
		}
	}
	return false
}
