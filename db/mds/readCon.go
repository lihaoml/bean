package mds

import (
	. "bean"
	"bean/db/influx"
	"bean/utils"
	"encoding/json"
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
func (mds MDS) GetContractOrderBookTS(con *Contract, start, end time.Time, depth int, sample time.Duration) (OrderBookTS, error) {
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

	askMap := mds.getOrders2(mds.c, con.Name(), "ASK", timeFrom, timeTo, depth, sampleStr)
	bidMap := mds.getOrders2(mds.c, con.Name(), "BID", timeFrom, timeTo, depth, sampleStr)

	for k, askOrders := range askMap {
		tm, _ := time.Parse(time.RFC3339, k)
		bidOrders := bidMap[k]
		ob := NewOrderBook(bidOrders, askOrders)
		obts = append(obts, OrderBookT{ob, tm})
	}

	for k, bidOrders := range bidMap {
		tm, _ := time.Parse(time.RFC3339, k)
		if _, askexists := askMap[k]; !askexists {
			ob := NewOrderBook(bidOrders, nil)
			obts = append(obts, OrderBookT{ob, tm})
		}
	}
	return obts.Sort(), nil
}

// GetMarket retrieves an entire market of contracts as per a specific time
func (mds MDS) GetMarketRaw(exName string, underlying Pair, snap time.Time) (map[string]OrderBookT, error) {
	cmd := "SELECT instrument,side,index,Amount,last(Price) as Price from " + MT_ORDERBOOK + // TODO: change MT_ORDERBOOK to MT_CONTRACT_ORDERBOOK when mds migration is done
		" WHERE time <='" + snap.Format(time.RFC3339) + "'" +
		" and time >='" + snap.Add(-12*time.Hour).Format(time.RFC3339) + "'" +
		" and exchange = '" + exName + "'" +
		" and index='0' " +
		" GROUP BY instrument,side,index"
	resp, err := influx.QueryDB(MDS_DBNAME, mds.c, cmd)
	if err != nil {
		return nil, err
	}
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return nil, err
	}

	mkt := make(map[string]OrderBookT)
	// group result by time
	for _, row := range resp[0].Series {
		for _, d := range row.Values {
			// fmt.Println(d)
			t, _ := time.Parse(time.RFC3339, d[0].(string))
			instr := d[1].(string)
			side := d[2].(string)
			amt, _ := d[4].(json.Number).Float64()
			prc, _ := d[5].(json.Number).Float64()
			if _, exist := mkt[instr]; !exist {
				mkt[instr] = OrderBookT{EmptyOrderBook(), t}
			}
			if side == "BID" {
				mkt[instr].InsertBid(Order{Price: prc, Amount: amt})
				// mkt[instr].Time = t
			}
			if side == "ASK" {
				mkt[instr].InsertAsk(Order{Price: prc, Amount: amt})
			}
		}
	}
	return mkt, nil
}

// internal functions
func (mds MDS) getOrders2(c client.Client, instrument string, side string, timeFrom string, timeTo string, indexLimit int, sample string) map[string][]Order {
	if indexLimit < 1 {
		log.Fatal("index limit should be positive integer")
	}

	orders := make(map[string][]Order)
	for i := 0; i < indexLimit; i++ {
		order := mds.getOrder2(c, instrument, side, timeFrom, timeTo, strconv.Itoa(i), sample)
		for key, val := range order {
			orders[key] = append(orders[key], val)
		}
	}

	return orders
}

func (mds MDS) getOrder2(c client.Client, instrument string, side string, timeFrom string, timeTo string, index string, sample string) map[string]Order {
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
