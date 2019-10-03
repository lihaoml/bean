package mds

import (
	. "bean"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"math"
	"strconv"
	"time"
)

func WriteContractOrderBook(exName string, instr string, obt OrderBookT) {
	c, err := connect()
	if err != nil {
		panic(err.Error())
	}
	defer c.Close()
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "ms",
	})
	writeOBBatchPoints(bp, exName, instr, "BID", obt.Bids(), obt.Time, 0)
	writeOBBatchPoints(bp, exName, instr, "ASK", obt.Asks(), obt.Time, 0)
	c.Write(bp)
}

func WriteContractTransactions(exName string, pts []ConTxnPoint) {
	c, err := connect()
	if err != nil {
		panic(err.Error())
	}
	defer c.Close()
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "ms",
	})
	for _, pt := range pts {
		fmt.Println(pt)
		writeTxnBatchPoints(bp, exName, pt.Instrument, pt.Side, pt.Price, pt.Amount, pt.IndexPrice, pt.Vol, pt.TimeStamp)
	}
	fmt.Println("start writing")
	c.Write(bp)
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
	for index, o := range orders {
		if index >= OB_LIMIT {
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

func writeTxnBatchPoints(bp client.BatchPoints, exName, instr string, side Side, price, amount, indexPrice, vol float64, timeStamp time.Time) error {
	fields := map[string]interface{}{
		"Price":      price,
		"Amount":     amount,
		"IndexPrice": indexPrice,
		"Vol":        vol}

	tags := map[string]string{
		"instr":    instr,
		"exchange": exName,
		"side":     string(side),
	}
	pt, err := client.NewPoint(MT_CONTRACT_TRANSACTION, tags, fields, timeStamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)
	return nil
}
