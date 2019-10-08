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
	OB_LIMIT     = 150
	OB_OPT_LIMIT = 15 // for options we save only 20 ticks in the orderbook, otherwise the DB series is too much
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
	return m.WriteBatchPoints(bp)
}

func (mds MDS) WriteOBPoints(ob OrderBookT, exName string, pair Pair) error {
	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "ms",
	})
	writeSpotOBBatchPoints(bp, exName, pair, ob.OrderBook, ob.Time)
	return mds.WriteBatchPoints(bp)
}

func (mds MDS) WriteTXNPoints(trans Transactions, exName string) error {
	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "us",
	})

	for _, t := range trans {
		writeSpotTxnBatchPoints(bp, exName, t)
	}
	return mds.WriteBatchPoints(bp)
}

func writeSpotTxnBatchPoints(bp client.BatchPoints, exName string, txn Transaction) error {
	fields := map[string]interface{}{
		"Price":  txn.Price,
		"Amount": txn.Amount,
	}
	side := BUY
	if txn.Maker == Buyer {
		side = SELL
	}
	tags := map[string]string{
		"LHS":      string(txn.Pair.Coin),
		"RHS":      string(txn.Pair.Base),
		"exchange": exName,
		"side":     string(side),
	}
	// inject last digi of transaction index to time stamp to differenciate transacitons happening at same milisecond
	ts := txn.TimeStamp
	if len(txn.TxnID) > 0 {
		ts = txn.TimeStamp.Add(time.Duration(util.SafeFloat64(txn.TxnID[len(txn.TxnID)-1:])))
	}
	newpt, err := client.NewPoint(MT_TRANSACTION, tags, fields, ts)
	if err != nil {
		return err
	}
	bp.AddPoint(newpt)
	return nil
}

func writeSpotOBBatchPoints(bp client.BatchPoints, exName string, pair Pair, ob OrderBook, timeStamp time.Time) error {
	for index, bid := range ob.Bids() {
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
		pt, err := client.NewPoint(MT_ORDERBOOK, tags, fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	for index, ask := range ob.Asks() {
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
		pt, err := client.NewPoint(MT_ORDERBOOK, tags, fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	return nil
}
