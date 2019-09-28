package mds

import (
	. "bean"
	"bean/db/influx"
	"bean/utils"
	"errors"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"strconv"
	"time"
)

// Read2 implements functions for reading the market contract information using the ALTERNATIVE db schema
// Instead of

// GetContractOrderBookTS gets the history of a specific contract from period start to end
// sample dictates the sample rate of the data. set to zero for the full dataset
func GetContractOrderBookTS(contractName string, start, end time.Time, depth int, sample time.Duration) (OrderBookTS, error) {
	c, err := connect()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var obts OrderBookTS
	timeFrom := start.Format(time.RFC3339)
	timeTo := end.Format(time.RFC3339)

	var sampleStr string
	switch sample {
	case time.Minute:
		sampleStr = "1m"
	case 5 * time.Minute:
		sampleStr = "5m"
	case 30 * time.Minute:
		sampleStr = "30m"
	case time.Hour:
		sampleStr = "1h"
	case 24 * time.Hour:
		sampleStr = "1d"
	default:
		return nil, errors.New("Unknown sample frequency")
	}

	askMap := getOrders2(c, contractName, "ASK", timeFrom, timeTo, depth, sampleStr)
	bidMap := getOrders2(c, contractName, "BID", timeFrom, timeTo, depth, sampleStr)

	for k, v := range askMap {
		tm, _ := time.Parse(time.RFC3339, k)
		ob := NewOrderBook(bidMap[k], v)
		obts = append(obts, OrderBookT{ob, tm})
	}

	for k, v := range bidMap {
		tm, _ := time.Parse(time.RFC3339, k)
		if len(askMap[k]) == 0 {
			ob := NewOrderBook(v, askMap[k])
			obts = append(obts, OrderBookT{ob, tm})
		}
	}
	return obts.Sort(), nil
}

// internal functions
func getOrders2(c client.Client, instrument string, side string, timeFrom string, timeTo string, indexLimit int, sample string) map[string][]Order {
	if indexLimit < 1 {
		log.Fatal("index limit should be positive integer")
	}

	orders := make(map[string][]Order)
	for i := 0; i < indexLimit; i++ {
		order := getOrder2(c, instrument, side, timeFrom, timeTo, strconv.Itoa(i), sample)
		for key, val := range order {
			orders[key] = append(orders[key], val)
		}
	}

	return orders
}

func getOrder2(c client.Client, instrument string, side string, timeFrom string, timeTo string, index string, sample string) map[string]Order {
	var query string
	if sample == "" {
		query = "select Amount,Price,index from \"" + instrument +
			"\" where time >='" + timeFrom + "' and time <='" + timeTo +
			"' and index = '" + index + "' " +
			" and side='" + side + "'"
	} else {
		query = "select last(Amount),Price,index from " + MT_ORDERBOOK +
			" where instrument='" + instrument + "'" +
			" and time >='" + timeFrom + "' and time <='" + timeTo + "'" +
			" and index = '" + index + "'" +
			" and side='" + side + "'" +
			" group by time(" + sample + ")"
	}

	resp, err := influx.QueryDB(MDS_DBNAME, c, query)
	if err != nil {
		log.Fatal(err)
	}
	if len(resp) == 0 || len(resp[0].Series) == 0 {
		return make(map[string]Order)
	}
	row := resp[0].Series[0]

	var feed = make([]OrderPoint, len(row.Values))
	for i, d := range row.Values {
		//fmt.Printf("%T,%T,%T,%T,%T\n", d[0], d[1], d[2], d[3])
		if d[1] != nil || d[2] != nil || d[3] != nil {
			t1 := util.SafeFloat64(d[1])
			t2 := util.SafeFloat64(d[2])
			feed[i] = OrderPoint{time: d[0].(string), amount: t1, price: t2, index: d[3].(string)}
		}
	}

	dborder := convertToOrders(feed)

	return dborder
}
