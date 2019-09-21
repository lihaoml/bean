package tds

import (
	. "bean"
	"bean/brew"
	"bean/logger"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"time"
)

// dealer is normally the name of a strategy
// param is the formatted parameter set of the strategy
// remark is whatever you want to remark for the order
// the function should never panic,
// if there is problem with DB or orders it should just return an error and leave it to the caller to deal with the error
func RecordPlacedOrder(oids []brew.ExNameWithOID, acct, dealer, param, remark string) error {
	c, err := connect()
	if err != nil {
		return err
	}
	defer c.Close()
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  TDS_DBNAME,
		Precision: "s",
	})
	addTRPoints(acct, dealer, param, remark, oids, &bp)
	//Write the batch
	if err = c.Write(bp); err != nil {
		return err
	}
	return nil
}

func addTRPoints(acctName, dealer, param, remark string, oids []brew.ExNameWithOID, bp *client.BatchPoints) error {
	cnt := 0
	for _, o := range oids {
		fmt.Println(o.TimeStamp, o.ExName, o.Pair, o.OrderID)
		//		addTRPoints(acctName, trades, ex.Name(), &bp)
		txn_fields := make(map[string]interface{})
		txn_fields["OrderID"] = o.OrderID
		txn_fields["REMARK"] = remark
		txn_fields["PARAM"] = param
		tags := map[string]string{
			"COUNT":    fmt.Sprint(cnt), // use count to prevent order placed at the same time from being overwritten
			"account":  acctName,
			"exchange": o.ExName,
			"dealer":   dealer,
			"LHS":      string(o.Pair.Coin),
			"RHS":      string(o.Pair.Base),
		}
		pt, err := client.NewPoint(MT_PLACED_ORDER, tags, txn_fields, o.TimeStamp)
		if err != nil {
			return err
		}
		(*bp).AddPoint(pt)
		cnt++
	}
	return nil
}

// record portfolio to influx db as of now
func RecordPortfolios(ports map[string]Portfolio, acctName string, timeStamp time.Time) error {
	c, err := connect()
	if err != nil {
		return err
	}
	defer c.Close()
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  BALANCE_DBNAME,
		Precision: "s",
	})
	for exName, port := range ports {
		for c, v := range port.Balances() {
			port_fields := make(map[string]interface{})
			port_fields[string(c)] = v

			tags := make(map[string]string)
			tags["exchange"] = exName
			tags["account"] = acctName
			tags["aggregated"] = "NO"

			pt, err := client.NewPoint(MT_COIN_BALANCE, tags, port_fields, timeStamp)
			if err != nil {
				return err
			}
			bp.AddPoint(pt)
		}
	}
	//Write the batch
	err = c.Write(bp)
	return err
}

/////////////////////////////////////////////////////////////////////
// record trades from the last hour
func RecordTrades(exTrades map[string]TradeLogS, acctName string) error {
	c, err := connect()
	if err != nil {
		logger.Warn().Msg(err.Error())
		return err
	}
	defer c.Close()

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  TDS_DBNAME,
		Precision: "s",
	})

	for exName, trades := range exTrades {
		err = addTradePoints(acctName, trades, exName, nil, &bp)
		if err != nil {
			logger.Warn().Msg(err.Error())
			return err
		}
	}
	//Write the batch
	err = c.Write(bp)
	if err != nil {
		logger.Warn().Msg(err.Error())
	}
	return err
}

func addTradePoints(acctName string, trades TradeLogS, exName string, dealerInfo map[string]DealerInfo, bp *client.BatchPoints) (err error) {
	tagCount := make(map[int64]int)
	for _, t := range trades {
		txn_fields := make(map[string]interface{})
		txn_fields["Price"] = t.Price
		if t.Side == BUY {
			txn_fields["Amount"] = t.Quantity
		} else {
			txn_fields["Amount"] = -t.Quantity
		}
		txn_fields["CommissionAsset"] = t.CommissionAsset
		txn_fields["Commission"] = t.Commission
		txn_fields["OrderID"] = t.OrderID
		txn_fields["TxnID"] = t.TxnID
		if dealerInfo != nil {
			if info, exists := dealerInfo[t.OrderID]; exists {
				txn_fields["dealer"] = info.Dealer
				txn_fields["param"] = info.Param
				txn_fields["remark"] = info.Remark
			}
		}
		tt := t.Time.Unix()
		cnt := 0
		if c, exists := tagCount[tt]; exists {
			tagCount[tt]++
			cnt = c
		} else {
			tagCount[tt] = 1
		}
		tags := map[string]string{
			"COUNT":    fmt.Sprint(cnt),
			"account":  acctName,
			"exchange": exName,
			"LHS":      string(t.Pair.Coin),
			"RHS":      string(t.Pair.Base),
		}
		pt, err := client.NewPoint(MT_TRADE, tags, txn_fields, t.Time)
		if err != nil {
			logger.Warn().Msg(err.Error()) // TODO: deal with errors
		}
		(*bp).AddPoint(pt)
	}
	return
}

////////////////////////////////////////////////////////////////////
// exchangeOrders: exName -> Pair -> order list
func RecordOpenOrders(exchangeOrders map[string](map[Pair]([]OrderStatus)), acctName string) error {
	c, err := connect()
	if err != nil {
		logger.Warn().Msg(err.Error())
		return err
	}
	defer c.Close()

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  TDS_DBNAME,
		Precision: "s",
	})
	err = addOpenOrderPoints(acctName, exchangeOrders, &bp)
	if err != nil {
		logger.Warn().Msg(err.Error())
		return err
	}
	//Write the batch
	err = c.Write(bp)
	if err != nil {
		logger.Warn().Msg(err.Error())
	}
	return err
}

func addOpenOrderPoints(acctName string, exchangeOrders map[string](map[Pair]([]OrderStatus)), bp *client.BatchPoints) (err error) {
	tt := time.Now()
	for exName, pairOrders := range exchangeOrders {
		for p, orders := range pairOrders {
			for i, o := range orders {
				tags := map[string]string{
					"account":  acctName,
					"exchange": exName,
					"LHS":      string(p.Coin),
					"RHS":      string(p.Base),
					"IDX":      fmt.Sprint(i),
					"SIDE":     string(o.Side),
				}
				trd_fields := make(map[string]interface{})
				trd_fields["Price"] = o.Price
				amt := o.LeftAmount
				if o.Side == SELL {
					amt = amt * -1
				}
				trd_fields["Amount"] = amt
				trd_fields["OrderID"] = o.OrderID
				pt, err := client.NewPoint(MT_OPEN_ORDER, tags, trd_fields, tt)
				if err != nil {
					logger.Warn().Msg(err.Error())
				}
				(*bp).AddPoint(pt)
			}
		}
	}
	return
}
