package mds

import (
	. "bean"
	"bean/db/influx"
	"encoding/json"
	"time"
)

// OrderPoint the structure to store one order point from DB
type OrderPoint struct {
	time   string
	amount float64
	price  float64
	index  string
}

// TransactPoint the structure to store one transaction point from DB
type TransactPoint struct {
	time       string
	amount     float64
	price      float64
	traderType string
}

// GetOrderBookTS : get orderbook time series (sorted)
func GetOrderBookTS2(exName string, pair Pair, start, end time.Time) (OrderBookTS, error) {
	var obts OrderBookTS
	c, err := connect()
	if err != nil {
		return obts, err
	}
	defer c.Close()

	timeFrom := start.Format(time.RFC3339)
	timeTo := end.Format(time.RFC3339)

	cmd := "select Amount,Price,side from " + MT_ORDERBOOK +
		" where time >='" + timeFrom + "' and time <='" + timeTo +
		"' and exchange = '" + exName +
		"' and LHS = '" + string(pair.Coin) +
		"' and RHS = '" + string(pair.Base) + "'"

	resp, err := influx.QueryDB(MDS_DBNAME, c, cmd)
	if err != nil {
		return obts, err
	}
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return obts, err
	}

	tmp := make(map[time.Time]OrderBook)
	// group result by time
	row := resp[0].Series[0]
	for _, d := range row.Values {
		// fmt.Println(d)
		t, _ := time.Parse(time.RFC3339, d[0].(string))
		amt, _ := d[1].(json.Number).Float64()
		prc, _ := d[2].(json.Number).Float64()
		side := d[3].(string)
		if _, exist := tmp[t]; !exist {
			tmp[t] = OrderBook{}
		}
		if side == "BID" {
			tmp[t] = OrderBook{append(tmp[t].Bids, Order{prc, amt}), tmp[t].Asks}
		} else if side == "ASK" {
			tmp[t] = OrderBook{tmp[t].Bids, append(tmp[t].Asks, Order{prc, amt})}
		} else {
			panic("unknown side: " + side)
		}
	}
	for t, ob := range tmp {
		obts = append(obts, OrderBookT{t, ob.Sort()})
	}
	return obts.Sort(), nil
}

func GetTransactions2(exName string, pair Pair, start, end time.Time) (Transactions, error) {
	var txns Transactions
	c, err := connect()
	if err != nil {
		return txns, err
	}
	defer c.Close()

	timeFrom := start.Format(time.RFC3339)
	timeTo := end.Format(time.RFC3339)

	cmd := "select Amount,Price,side from " + MT_TRANSACTION +
		" where time >='" + timeFrom + "' and time <='" + timeTo +
		"' and exchange = '" + exName +
		"' and LHS = '" + string(pair.Coin) +
		"' and RHS = '" + string(pair.Base) + "'"

	resp, err := influx.QueryDB(MDS_DBNAME, c, cmd)
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
		txns = append(txns, Transaction{Pair: pair, Price: price, Amount: amount, TimeStamp: timestamp, Maker: maker})
	}
	return txns.Sort(), nil
}

func convertToOrders(feed []OrderPoint) map[string]Order {
	dborders := make(map[string]Order)
	for _, v := range feed {
		if v.amount > 0 {
			dborders[v.time] = Order{Amount: v.amount, Price: v.price}
		}
	}
	return dborders
}
