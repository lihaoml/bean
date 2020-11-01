package mds

import (
	. "bean"
	"bean/db/influx"
	"bean/utils"
	"fmt"
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
		Precision: "ns",
	})

	for _, t := range trans {
		writeSpotTxnBatchPoints(bp, exName, t)
	}
	return mds.WriteBatchPoints(bp)
}

func (mds MDS) WritePoints(pts []influx.Point, measurement string) error {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "s",
	})
	for i, p := range pts {
		pt, err := client.NewPoint(measurement, p.Tags, p.Fields, p.TimeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
		if i % 10000 == 0 || i == len(pts) - 1{
			err := mds.WriteBatchPoints(bp)
			if err != nil {
				return err
			}else {
				bp, _ = client.NewBatchPoints(client.BatchPointsConfig{
					Database:  MDS_DBNAME,
					Precision: "s",
				})
			}
		}
	}
	return nil
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
	if len(txn.TxnID) > 1 {
		ts = txn.TimeStamp.Add(time.Duration(util.SafeFloat64(txn.TxnID[len(txn.TxnID)-2:])))
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

	// extracting orderbook stats and record them
	if ob.Valid() {
		cumpctOB := ob.CumPctOB()
		tags := map[string]string{
			"LHS":      string(pair.Coin),
			"RHS":      string(pair.Base),
			"exchange": exName,
		}
		fields := map[string]interface{}{}
		pcts := []int{0, 1, 5, 10}
		for _, p := range pcts {
			if len(cumpctOB.CumPctBids) > p {
				fields["bid_vwap_"+fmt.Sprint(p)] = cumpctOB.CumPctBids[p].Price
				fields["bid_camt_"+fmt.Sprint(p)] = cumpctOB.CumPctBids[p].Amount
			}
			if len(cumpctOB.CumPctAsks) > p {
				fields["ask_vwap_"+fmt.Sprint(p)] = cumpctOB.CumPctAsks[p].Price
				fields["ask_camt_"+fmt.Sprint(p)] = cumpctOB.CumPctAsks[p].Amount
			}
		}
		mid := (cumpctOB.CumPctAsks[0].Price + cumpctOB.CumPctBids[0].Price) / 2
		fields["mid"] = mid
		fields["spread"] = (cumpctOB.CumPctAsks[0].Price - cumpctOB.CumPctBids[0].Price) / mid
		// price in amounts

		piaBaseAmt := []float64{1, 5, 10}
		sizeScaler := map[Coin]float64{
			BTC: 0.1 / mid, ETH: 5 / mid,
			USD: 1000 / mid, USDT: 1000 / mid, USDC: 1000 / mid, PAX: 1000 / mid, TUSD: 1000 / mid,
		}
		for _, p := range piaBaseAmt {
			size := sizeScaler[pair.Base] * p
			bid, ask, bidSize, askSize := ob.PriceIn(size)
			fields["bid_pia_price_"+fmt.Sprint(p)] = bid
			fields["bid_pia_amount_"+fmt.Sprint(p)] = bidSize
			fields["ask_pia_price_"+fmt.Sprint(p)] = ask
			fields["ask_pia_amount_"+fmt.Sprint(p)] = askSize
		}
		pt, err := client.NewPoint(MT_ORDERBOOK_STATS, tags, fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	return nil
}
