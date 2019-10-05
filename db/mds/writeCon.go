package mds

import (
	. "bean"
	util "bean/utils"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"math"
	"strconv"
	"time"
)

func (mds MDS) WriteContractOrderBook(exName string, instr string, obt OrderBookT) error {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "ms",
	})
	writeOBBatchPoints(bp, exName, instr, "BID", obt.Bids(), obt.Time, 0)
	writeOBBatchPoints(bp, exName, instr, "ASK", obt.Asks(), obt.Time, 0)
	return mds.WriteBatchPoints(bp)
}

func (mds MDS) WriteContractTransactions(pts []ConTxnPoint) error {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "us",
	})
	for _, pt := range pts {
		fmt.Println(pt)
		writeConTxnBatchPoints(bp, pt)
	}
	fmt.Println("start writing")
	return mds.WriteBatchPoints(bp)
}

func writeOBBatchPoints(bp client.BatchPoints, exName, instr string, side Side, orders []Order, timeStamp time.Time, lag time.Duration) error {

	if orders == nil {
		tags := map[string]string{
			"index":      "0",
			"instrument": instr,
			"exchange":   exName,
			"side":       string(side),
		}

		fields := make(map[string]interface{})
		fields["Price"] = math.NaN()
		fields["Amount"] = 0.0
		fields["Lag"] = lag.Seconds()
		pt, err := client.NewPoint(MT_CONTRACT_ORDERBOOK, tags, fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	con, err := ContractFromName(instr)
	if err != nil {
		return err
	}
	limit := OB_LIMIT
	if con.IsOption() {
		limit = OB_OPT_LIMIT
	}

	for index, o := range orders {
		if index >= limit {
			break
		}
		fields := make(map[string]interface{})
		fields["Price"] = o.Price
		fields["Amount"] = o.Amount
		fields["Lag"] = lag.Seconds()
		tags := map[string]string{
			"index":      strconv.Itoa(index),
			"instrument": instr,
			"exchange":   exName,
			"side":       string(side),
		}

		pt, err := client.NewPoint(MT_CONTRACT_ORDERBOOK, tags, fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	return nil
}

func writeConTxnBatchPoints(bp client.BatchPoints, pt ConTxnPoint) error {
	fields := map[string]interface{}{
		"Price":      pt.Price,
		"Amount":     pt.Amount,
		"IndexPrice": pt.IndexPrice,
		"Vol":        pt.Vol}

	// FIXME: could miss transactions if two different transactions on the same instrument have identical time stamp, side, and exchange
	tags := map[string]string{
		"instr":    pt.Instrument,
		"exchange": pt.ExName,
		"side":     string(pt.Side),
	}

	// inject last digi of transaction index to time stamp to differenciate transacitons happening at same milisecond
	ts := pt.TimeStamp
	if len(pt.TxnID) > 0 {
		ts = pt.TimeStamp.Add(time.Duration(util.SafeFloat64(pt.TxnID[len(pt.TxnID)-1:])))
	}

	newpt, err := client.NewPoint(MT_CONTRACT_TRANSACTION, tags, fields, ts)
	if err != nil {
		return err
	}
	bp.AddPoint(newpt)
	return nil
}
