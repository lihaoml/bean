package tds

import (
	. "bean"
	"bean/brew"
	"bean/db/influx"
	"bean/logger"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"math"
	"time"
)

// dealer is normally the name of a strategy
// param is the formatted parameter set of the strategy
// remark is whatever you want to remark for the order
// the function should never panic,
// if there is problem with DB or orders it should just return an error and leave it to the caller to deal with the error
func RecordPlacedOrder(oids []brew.ExNameWithOID, acct, dealer, param, remark string) error {
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	if err != nil {
		return err
	}
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  TDS_DBNAME,
		Precision: "s",
	})
	addTRPoints(acct, dealer, param, remark, oids, &bp)
	return influx.WriteBatchPoints(cs, bp)
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

// record portfolio to influx db as of now, agg: is it aggregated MTM?
func RecordPortfolios(ports map[string]Portfolio, acctName string, timeStamp time.Time, agg bool) error {
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	if err != nil {
		return err
	}
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  BALANCE_DBNAME,
		Precision: "s",
	})
	for exName, port := range ports {
		for c, v := range port.Balances() {
			port_fields := make(map[string]interface{})
			port_fields[string(c)] = v

			tags := make(map[string]string)
			tags["account"] = acctName
			tags["exchange"] = exName
			if agg {
				tags["aggregated"] = "YES"
			} else {
				tags["aggregated"] = "NO"
			}
			pt, err := client.NewPoint(MT_COIN_BALANCE, tags, port_fields, timeStamp)
			if err != nil {
				fmt.Println(err.Error())
			}
			bp.AddPoint(pt)
		}
	}
	return influx.WriteBatchPoints(cs, bp)
}

// record aggregated portfolio to influx db as of now, isPL: true: write to MT_PNL_BALANCE, false: write to MT_TOTAL_BALANCE
func RecordTotalPortfolio(port Portfolio, acctName string, timeStamp time.Time, isPL bool, agg bool) error {
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	if err != nil {
		return err
	}
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  BALANCE_DBNAME,
		Precision: "s",
	})
	for c, v := range port.Balances() {
		port_fields := make(map[string]interface{})
		port_fields[string(c)] = v
		tags := make(map[string]string)
		tags["account"] = acctName
		if agg {
			tags["aggregated"] = "YES"
		} else {
			tags["aggregated"] = "NO"
		}

		mt := MT_TOTAL_BALANCE
		if isPL {
			mt = MT_PNL_BALANCE
		}
		pt, err := client.NewPoint(mt, tags, port_fields, timeStamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)

		// ad hoc MTM_USDT in PNL_BALANCE measurement
		if c == USDT && !isPL && agg {
			hack_fields := make(map[string]interface{})
			hack_fields["MTM_"+string(c)] = v
			pt, err := client.NewPoint(MT_PNL_BALANCE, tags, hack_fields, timeStamp)
			if err != nil {
				return err
			}
			bp.AddPoint(pt)
		}
	}
	return influx.WriteBatchPoints(cs, bp)
}

type MarginACPoint struct {
	LHS_BAL      float64
	RHS_BAL      float64
	LHS_IN_ORDER float64
	RHS_IN_ORDER float64
	LHS_LOAN     float64
	RHS_LOAN     float64
	RISK_RATE    float64
	Pair         Pair
	ExName       string
	MTM          float64 // mtm in USDT
	UMTM         float64 // net PV in usdt
}

func RecordMarginPoints(acName string, mpts []MarginACPoint, timeStamp time.Time) error {
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	if err != nil {
		return err
	}
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  BALANCE_DBNAME,
		Precision: "s",
	})
	for _, mpt := range mpts {
		ac_fields := make(map[string]interface{})
		ac_fields["LHS_BAL"] = mpt.LHS_BAL
		ac_fields["RHS_BAL"] = mpt.RHS_BAL
		ac_fields["LHS_IN_ORDER"] = mpt.LHS_IN_ORDER
		ac_fields["RHS_IN_ORDER"] = mpt.RHS_IN_ORDER
		ac_fields["LHS_LOAN"] = mpt.LHS_LOAN
		ac_fields["RHS_LOAN"] = mpt.RHS_LOAN
		if mpt.RISK_RATE > 0 {
			ac_fields["RISK_RATE"] = mpt.RISK_RATE
		}

		tags := map[string]string{
			"PAIR":     mpt.Pair.String(),
			"account":  acName,
			"exchange": mpt.ExName,
		}
		ac_fields["MTM"] = mpt.MTM // MTM of leveraged asset
		ac_fields["uMTM"] = mpt.UMTM

		if !math.IsNaN(mpt.MTM) && !math.IsNaN(mpt.UMTM) {
			pt, err := client.NewPoint(MT_MARGIN_ACCOUNT_INFO, tags, ac_fields, timeStamp)
			if err != nil {
				panic(err.Error()) // TODO: deal with errors
			}
			bp.AddPoint(pt)
		}
	}
	return influx.WriteBatchPoints(cs, bp)
}

/////////////////////////////////////////////////////////////////////
// record trades from the last hour
func RecordTrades(exTrades map[string]TradeLogS, acctName string) error {
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	if err != nil {
		logger.Warn().Msg(err.Error())
		return err
	}
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
	return influx.WriteBatchPoints(cs, bp)
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
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	if err != nil {
		logger.Warn().Msg(err.Error())
		return err
	}
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
	return influx.WriteBatchPoints(cs, bp)
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
