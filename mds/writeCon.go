package mds

import (
	"bean"
	. "bean"
	"github.com/influxdata/influxdb/client/v2"
	"math"
	"strconv"
	"time"
)

func writeOBBatchPoints(bp client.BatchPoints, exName, instr string, side Side, orders []bean.Order, timeStamp time.Time) error {

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
		pt, err := client.NewPoint(MT_ORDERBOOK, tags, fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	for index, bid := range orders {
		if index >= OB_LIMIT {
			break
		}
		fields := make(map[string]interface{})
		fields["Price"] = bid.Price
		fields["Amount"] = bid.Amount
		tags := map[string]string{
			"index":      strconv.Itoa(index),
			"instrument": instr,
			"exchange":   exName,
			"side":       string(side),
		}
		pt, err := client.NewPoint(MT_ORDERBOOK, tags, fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	return nil
}
