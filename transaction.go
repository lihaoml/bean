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
func (txn Transactions) GetTransactionHistory(t time.Time) Transactions {
	var idx int
	for i, tt := range txn {
		if tt.TimeStamp.Before(t) {
			idx = i
		} else {
			break
		}
	}
	res := txn[0:idx]
	return res
}
