package mds

import (
	. "bean"
	"bean/utils"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"strconv"
	"time"
)

const (
	OB_LIMIT = 150
)

func (m MDS) WriteTick(lhs, rhs string, source string, bid, ask float64, t time.Time) error {
	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "ms",
	})
	fields := make(map[string]interface{})
	fields["bid"] = bid
	fields["ask"] = ask
	tags := map[string]string{
		"LHS":    lhs,
		"RHS":    rhs,
		"source": source,
	}
	pt, err := client.NewPoint(MT_TICK, tags, fields, t)
	if err != nil {
		log.Fatal(err) // TODO: deal with errors
	}
	bp.AddPoint(pt)
	return m.c.Write(bp)
}

func WriteOBPoints(ob OrderBookT, exName string, pair Pair) error {
	c, err := connect()
	if err != nil {
		return err
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "ms",
	})

	for index, bid := range ob.OrderBook.Bids() {
		if index >= OB_LIMIT {
			break
		}
		fields := make(map[string]interface{})
		fields["Price"] = bid.Price
		fields["Amount"] = bid.Amount
		tags := map[string]string{
			"index":    strconv.Itoa(index),
			"LHS":      string(pair.Coin),
			"RHS":      string(pair.Base),
			"exchange": exName,
			"side":     "BID",
		}
		pt, err := client.NewPoint(MT_ORDERBOOK, tags, fields, ob.Time)
		if err != nil {
			log.Fatal(err) // TODO: deal with errors
		}
		bp.AddPoint(pt)
	}
	for index, ask := range ob.OrderBook.Asks() {
		if index >= OB_LIMIT {
			break
		}
		fields := make(map[string]interface{})
		fields["Price"] = ask.Price
		fields["Amount"] = ask.Amount
		tags := map[string]string{
			"index":    strconv.Itoa(index),
			"LHS":      string(pair.Coin),
			"RHS":      string(pair.Base),
			"exchange": exName,
			"side":     "ASK",
		}
		pt, err := client.NewPoint(MT_ORDERBOOK, tags, fields, ob.Time)
		if err != nil {
			log.Fatal(err) // TODO: deal with errors
		}
		bp.AddPoint(pt)
	}
	return c.Write(bp)
}

func WriteTXNPoints(trans Transactions, exName string) error {
	c, err := connect()
	if err != nil {
		return err
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "us",
	})

	for _, t := range trans {
		txn_fields := make(map[string]interface{})
		txn_fields["Price"] = t.Price
		txn_fields["Amount"] = t.Amount

		tags := map[string]string{
			"LHS":      string(t.Pair.Coin),
			"RHS":      string(t.Pair.Base),
			"exchange": exName,
		}
		if t.Maker == Buyer {
			tags["side"] = "SELL"
		} else {
			tags["side"] = "BUY"
		}
		// inject last digi of transaction index to time stamp to differenciate transacitons happening at same milisecond
		ts := t.TimeStamp
		if len(t.TxnID) > 0 {
			ts = t.TimeStamp.Add(time.Duration(util.SafeFloat64(t.TxnID[len(t.TxnID)-1:])))
		}
		pt, err := client.NewPoint(MT_TRANSACTION, tags, txn_fields, ts)
		if err != nil {
			log.Fatal(err) // TODO: deal with errors
		}
		bp.AddPoint(pt)
	}
	return c.Write(bp)
}
