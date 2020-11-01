package mds

import (
	"bean"
	. "bean"
	"bean/db/influx"
	util "bean/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
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

	if len(mds.cs) == 0 {
		return nil, errors.New("no MDS connection established")
	}
	askMap := getOrders2(mds.cs[0], con.Name(), "ASK", timeFrom, timeTo, depth, sampleStr)
	bidMap := getOrders2(mds.cs[0], con.Name(), "BID", timeFrom, timeTo, depth, sampleStr)

	for k, askOrders := range askMap {
		tm, _ := time.Parse(time.RFC3339, k)
		bidOrders := bidMap[k]
		ob := NewOrderBook(bidOrders, askOrders)
		obts = append(obts, OrderBookT{OrderBook: ob, Time: tm})
	}

	for k, bidOrders := range bidMap {
		tm, _ := time.Parse(time.RFC3339, k)
		if _, askexists := askMap[k]; !askexists {
			ob := NewOrderBook(bidOrders, nil)
			obts = append(obts, OrderBookT{OrderBook: ob, Time: tm})
		}
	}
	return obts.Sort(), nil
}

// GetMarket retrieves an entire market of contracts as per a specific time
func (mds MDS) GetAllContractOrderBooks(exName string, underlying Pair, snap time.Time) (map[string]OrderBookT, error) {
	cmd := "SELECT instrument,side,index,Amount,last(Price) as Price from " + MT_CONTRACT_ORDERBOOK +
		" WHERE time <='" + snap.Format(time.RFC3339) + "'" +
		" and time >='" + snap.Add(-12*time.Hour).Format(time.RFC3339) + "'" +
		" and exchange = '" + exName + "'" +
		" and index='0' " +
		" GROUP BY instrument,side,index"
	if len(mds.cs) == 0 {
		return nil, errors.New("no MDS connection established")
	}
	resp, err := influx.QueryDB(MDS_DBNAME, mds.cs[0], cmd)
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
				mkt[instr] = OrderBookT{EmptyOrderBook(), t, 0} // TODO: review changeID
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

func (mds MDS) GetLatestHourlyAlpha(exName string, instr string, base string, alphaName string) (time.Time, map[Coin]float64, error) {
	return mds.getLatestAlpha(exName, instr, base, MT_HOURLY_ALPHA, alphaName)
}
func (mds MDS) GetLatestMinAlpha(exName string, instr string, base string, nMin int, alphaName string) (time.Time, map[Coin]float64, error) {
	return mds.getLatestAlpha(exName, instr, base, "MIN_" + fmt.Sprint(nMin) + "_ALPHA", alphaName)
}

// GetMarket retrieves an entire market of contracts as per a specific time
func (mds MDS) getLatestAlpha(exName string, instr string, base string, alphaTable, alphaName string) (time.Time, map[Coin]float64, error) {
	cmd := "SELECT *::field from " + alphaTable +
		" WHERE exchange = '" + exName + "' and \"name\" = '" + alphaName + "'" +
		" and instr = '" + instr + "' and base = '" + base + "'" +
		" order by time desc limit 1"
	if len(mds.cs) == 0 {
		return time.Now(), nil, errors.New("no MDS connection established")
	}
	resp, err := influx.QueryDB(MDS_DBNAME, mds.cs[0], cmd)
	if err != nil {
		return time.Now(), nil, err
	}
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return time.Now(), nil, err
	}
	vs := resp[0].Series[0]
	alpha := map[Coin]float64{}
	var t time.Time
	for i, c := range vs.Columns {
		if c == "time" {
			t, _ = time.Parse(time.RFC3339, vs.Values[0][i].(string))
		} else {
			alpha[Coin(strings.ToUpper(c))], _ = vs.Values[0][i].(json.Number).Float64()
		}
	}
	return t, alpha, nil
}

func (mds MDS) GetLatestHourlyData(exName string, instr string, base string, value string) (time.Time, map[Coin]float64, error) {
	return mds.getLatestData(exName, instr, base, value, MT_HOURLY_DATA)
}

func (mds MDS) GetLatestMinData(exName string, instr string, base string, value string, nMin int) (time.Time, map[Coin]float64, error) {
	return mds.getLatestData(exName, instr, base, value, "MIN_" + fmt.Sprint(nMin) + "_DATA")
}

// GetMarket retrieves an entire market of contracts as per a specific time
// value: OPEN | CLOSE | VWAP etc
func (mds MDS) getLatestData(exName string, instr string, base string, value string, dataTable string) (time.Time, map[Coin]float64, error) {
	cmd := "SELECT *::field from " + dataTable +
		" WHERE exchange = '" + exName + "' and value = '" + value + "'" +
		" and instr = '" + instr + "' and base = '" + base + "'" +
		" order by time desc limit 1"
	fmt.Println(cmd)
	if len(mds.cs) == 0 {
		return time.Now(), nil, errors.New("no MDS connection established")
	}
	resp, err := influx.QueryDB(MDS_DBNAME, mds.cs[0], cmd)
	if err != nil {
		return time.Now(), nil, err
	}
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return time.Now(), nil, err
	}
	vs := resp[0].Series[0]
	fmt.Println(vs)
	data := map[Coin]float64{}
	var t time.Time
	for i, c := range vs.Columns {
		if c == "time" {
			t, _ = time.Parse(time.RFC3339, vs.Values[0][i].(string))
		} else {
			data[Coin(strings.ToUpper(c))], _ = vs.Values[0][i].(json.Number).Float64()
		}
	}
	return t, data, nil
}

func (mds MDS) GetSmilesRaw(exName string, underlying Pair, snap time.Time) (smiles []SmilePoint, err error) {
	cmd := "SELECT expiry,last(Atm),RR25,Fly25,RR10,Fly10,Spot,Swaps from " + MT_SMILE +
		" WHERE time <='" + snap.Format(time.RFC3339) + "'" +
		" and time >='" + snap.Add(-45*time.Minute).Format(time.RFC3339) + "'" +
		" and pair = '" + underlying.String() + "'" +
		" GROUP BY expiry"
	if len(mds.cs) == 0 {
		return nil, errors.New("no MDS connection established")
	}
	resp, err := influx.QueryDB(MDS_DBNAME, mds.cs[0], cmd)
	if err != nil {
		return nil, err
	}
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return nil, err
	}

	// group result by time
	for _, row := range resp[0].Series {
		for _, d := range row.Values {
			// fmt.Println(d)
			t, _ := time.Parse(time.RFC3339, d[0].(string))
			expiry := d[1].(string)
			atm, _ := d[2].(json.Number).Float64()
			rr25, _ := d[3].(json.Number).Float64()
			fly25, _ := d[4].(json.Number).Float64()
			rr10, _ := d[5].(json.Number).Float64()
			fly10, _ := d[6].(json.Number).Float64()
			spot, _ := d[7].(json.Number).Float64()
			swaps, _ := d[8].(json.Number).Float64()

			exp, err := time.Parse(bean.ContractDateFormat, expiry)
			if err != nil {
				return nil, err
			}
			smiles = append(smiles,
				SmilePoint{
					TimeStamp: t,
					Pair:      underlying,
					Expiry:    exp,
					Atm:       atm / 100.0, RR25: rr25 / 100.0, RR10: rr10 / 100.0,
					Fly25: fly25 / 100.0, Fly10: fly10 / 100.0,
					Spot: spot, Swaps: swaps,
				})
		}
	}
	return
}

func (mds MDS) GetContractTXNs(exName string, instr string, start, end time.Time) (ContractTXNs, error) {
	var txns []ContractTXN
	timeFrom := start.Format(time.RFC3339)
	timeTo := end.Format(time.RFC3339)

	cmd := "select Amount,Price,side from " + MT_CONTRACT_TRANSACTION +
		" where time >='" + timeFrom + "' and time <='" + timeTo +
		"' and exchange = '" + exName +
		"' and instr = '" + instr + "'"
	if len(mds.cs) == 0 {
		return nil, errors.New("no MDS connection established")
	}
	resp, err := influx.QueryDB(MDS_DBNAME, mds.cs[0], cmd)
	if err != nil {
		panic(err.Error())
	}
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return txns, err
	}

	row := resp[0].Series[0]
	var feed = make([]TransactPoint, len(row.Values))
	for i, d := range row.Values {
		// fmt.Println(d)
		t1, _ := d[1].(json.Number).Float64()
		t2, _ := d[2].(json.Number).Float64()
		var side string
		// this example works!
		if m, ok := d[3].(string); ok {
			side = m
		} else if len(d) >= 5 {
			if m, ok := d[4].(string); ok {
				side = m
			}
		}
		feed[i] = TransactPoint{time: d[0].(string), amount: t1, price: t2, traderType: side} // FIXME: traderType is not side
	}

	for _, v := range feed {
		price := v.price
		amount := v.amount
		timestamp, _ := time.Parse(time.RFC3339, v.time)
		var maker TraderType
		if v.traderType == "BUY" {
			maker = Seller
		} else {
			maker = Buyer
		}
		txns = append(txns, ContractTXN{Instrument: instr, Price: price, Amount: amount, TimeStamp: timestamp, Maker: maker})
	}
	// sort by TimeStamp
	sort.Slice(txns, func(i, j int) bool { return txns[i].TimeStamp.Before(txns[j].TimeStamp) })
	return txns, nil
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
		query = "select Amount,Price,index from " + MT_CONTRACT_ORDERBOOK +
			" where instrument='" + instrument + "'" +
			" and time >='" + timeFrom + "' and time <='" + timeTo + "'" +
			" and index = '" + index + "'" +
			" and side='" + side + "'"
	} else {
		query = "select last(Amount),Price,index from " + MT_CONTRACT_ORDERBOOK +
			" where instrument='" + instrument + "'" +
			" and time >='" + timeFrom + "' and time <='" + timeTo + "'" +
			" and index = '" + index + "'" +
			" and side='" + side + "'" +
			" group by time(" + sample + ")"
	}

	resp, err := influx.QueryDB(MDS_DBNAME, c, query)
	fmt.Println(query)

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
