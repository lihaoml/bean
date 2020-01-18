package tds

import (
	. "bean"
	"bean/utils"
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"math"
	"time"
)

type DealerInfo struct {
	Dealer string
	Param  string
	Remark string
}

// GetTransactionHistory : get transaction time series
func GetTradeLogS(filter map[string]string, start, end time.Time) (TradeLogS, error) {
	var trds TradeLogS
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	if err == nil && len(cs) > 0 {
		timeFrom := start.Format(time.RFC3339)
		timeTo := end.Format(time.RFC3339)
		trds = getTrades(cs[0], TDS_DBNAME, filter, timeFrom, timeTo)
	}
	return trds, err
}

// return a orderID -> DealerInfo
func GetDealerInfo(filter map[string]string, start, end time.Time) map[string]DealerInfo {
	cs, err := connect()
	var res map[string]DealerInfo
	for _, c := range cs {
		defer c.Close()
	}
	if err == nil && len(cs) > 0 {
		timeFrom := start.Format(time.RFC3339)
		timeTo := end.Format(time.RFC3339)
		res = getDealers(cs[0], TDS_DBNAME, filter, timeFrom, timeTo)
	}
	return res
}

// get portfolio at time T for a particular account at a particular exchange
func GetPortfolio(exNames []string, coins []Coin, acct string, t time.Time) Portfolio {
	cs, _ := connect()
	for _, c := range cs {
		defer c.Close()
	}
	timeAt := t.Format(time.RFC3339)
	port := NewPortfolio()
	for _, exName := range exNames {
		port = port.Add(getBalances(cs[0], BALANCE_DBNAME, acct, exName, coins, timeAt))
	}
	return port
}

func GetLatestTotalPortfolio(exNames []string, coins Coins, acctName string) Portfolio {
	cs, _ := connect()
	for _, c := range cs {
		defer c.Close()
	}
	port := NewPortfolio()
	for _, exName := range exNames {
		for _, coin := range coins {
			v := getLatestBalance(cs[0], coin, exName, acctName)
			port.AddBalance(coin, v)
		}
	}
	return port
}

func GetInitPortfolio(coins Coins, acctName string) Portfolio {
	cs, _ := connect()
	for _, c := range cs {
		defer c.Close()
	}
	port := NewPortfolio()
	for _, coin := range coins {
		v := getInitBalance(cs[0], coin, acctName)
		port.AddBalance(coin, v)
	}
	return port
}

func GetTodayPL(acctName string) float64 {
	cs, _ := connect()
	for _, c := range cs {
		defer c.Close()
	}
	return getTodayPL(cs[0], acctName)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// select sum(PL)-sum(PL0) as dPL from (SELECT LAST("USDT") as PL, FIRST("USDT") as PL0 FROM PNL_BALANCE where  ("account" != 'SUPERMU') AND $timeFilter and aggregated = 'YES' GROUP by account)
// get today's PL (SGT)
func getTodayPL(client client.Client, acctName string) float64 {
	mt := MT_PNL_BALANCE
	loc, _ := time.LoadLocation("Asia/Singapore")
	//set timezone,
	dte := time.Now().In(loc).Format("2006-01-02")
	query := "select LAST(USDT) - FIRST(USDT) as dPL from \"" + mt + "\" where account='" + acctName + "' and aggregated = 'YES' and time >= '" + dte + "' TZ('Asia/Singapore')"
	fmt.Println(query)
	resp, err := queryDB(client, BALANCE_DBNAME, query)
	if err != nil {
		log.Fatal(err)
	}
	if len(resp) > 0 && len(resp[0].Series) > 0 && len(resp[0].Series[0].Values) > 0 {
		return util.SafeFloat64(resp[0].Series[0].Values[0][1])
	} else {
		fmt.Println("failing to query pl")
		return math.NaN()
	}
}

func GetLatestMTM(acctName string) float64 {
	cs, _ := connect()
	for _, c := range cs {
		defer c.Close()
	}
	mt := MT_MTM
	query := "select LAST(MTMUSD) from \"" + mt + "\" where account='" + acctName + "'"
	resp, err := queryDB(cs[0], BALANCE_DBNAME, query)
	if err != nil {
		log.Fatal(err)
	}
	if len(resp) > 0 && len(resp[0].Series) > 0 && len(resp[0].Series[0].Values) > 0 {
		return util.SafeFloat64(resp[0].Series[0].Values[0][1])
	} else {
		fmt.Println("failing to query pl")
		return math.NaN()
	}
}

func getLatestBalance(client client.Client, c Coin, exName string, acctName string) float64 {
	mt := MT_COIN_BALANCE
	query := "select LAST(" + string(c) + ") from \"" + mt + "\" where exchange = '" + exName + "' and account = '" + acctName + "' and aggregated = 'NO'" + " limit 1"
	// fmt.Println(query)
	resp, err := queryDB(client, BALANCE_DBNAME, query)
	if err != nil {
		log.Fatal(err)
	}
	lastBal := 0.0
	if len(resp) > 0 && len(resp[0].Series) > 0 && len(resp[0].Series[0].Values) > 0 {
		lastBal += util.SafeFloat64(resp[0].Series[0].Values[0][1])
	}

	// now we need to add margin account balance
	// RHS first
	// example query: select sum(V) - sum(L) from (select last(RHS_BAL) as V, last(RHS_LOAN) as L from "MARGIN_ACCOUNT_INFO" where account='UHL' and PAIR =~ /USDT$/ and time > now() - 5m group by PAIR)
	query = "select sum(V) - sum(L) from (select LAST(RHS_BAL) as V, LAST(RHS_LOAN) as L from \"" +
		MT_MARGIN_ACCOUNT_INFO + "\" where exchange = '" + exName +
		"' and account = '" + acctName +
		"' and PAIR =~ /" + string(c) + "$/ and time > now() - 8m group by PAIR)"
	// fmt.Println(query)
	resp, err = queryDB(client, BALANCE_DBNAME, query)
	if err != nil {
		log.Fatal(err)
	}
	if len(resp) > 0 && len(resp[0].Series) > 0 && len(resp[0].Series[0].Values) > 0 {
		lastBal += util.SafeFloat64(resp[0].Series[0].Values[0][1])
	}
	// now LHS
	query = "select sum(V) - sum(L) from (select LAST(LHS_BAL) as V, LAST(LHS_LOAN) as L from \"" +
		MT_MARGIN_ACCOUNT_INFO + "\" where exchange = '" + exName +
		"' and account = '" + acctName +
		"' and PAIR =~ /^" + string(c) + "/ and time > now() - 8m group by PAIR)"
	// fmt.Println(query)
	resp, err = queryDB(client, BALANCE_DBNAME, query)
	if err != nil {
		log.Fatal(err)
	}
	if len(resp) > 0 && len(resp[0].Series) > 0 && len(resp[0].Series[0].Values) > 0 {
		lastBal += util.SafeFloat64(resp[0].Series[0].Values[0][1])
	}
	return lastBal
}

func getInitBalance(client client.Client, c Coin, acctName string) float64 {
	query := "select sum(" + string(c) + ") from \"" + MT_PRINCIPAL + "\" where account = '" + acctName + "'"
	fmt.Println(query)
	resp, err := queryDB(client, BALANCE_DBNAME, query)
	if err != nil {
		log.Fatal(err)
	}
	if len(resp) == 0 || len(resp[0].Series) == 0 {
		return 0.0
	}
	row := resp[0].Series[0]
	if len(row.Values) < 1 {
		return 0.0
	}
	return util.SafeFloat64(row.Values[0][1])
}

// internal functions
func getTrades(c client.Client, dbName string, filter map[string]string, timeFrom string, timeTo string) TradeLogS {
	query := "select Price,Amount,CommissionAsset,Commission,OrderID,LHS,RHS,dealer from " + MT_TRADE +
		" where time >='" + timeFrom + "' and time <='" + timeTo + "'"
	for k, v := range filter {
		query += " and " + k + "='" + v + "'"
	}

	fmt.Println(query)
	resp, err := queryDB(c, dbName, query)
	if err != nil {
		panic(err.Error())
	}
	var trades TradeLogS
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return trades
	}
	row := resp[0].Series[0]
	for _, d := range row.Values {
		//		fmt.Println(d[0], d[1], d[2], d[3])
		price, _ := d[1].(json.Number).Float64()      // Price
		amt, _ := d[2].(json.Number).Float64()        // Amount
		commissionAsset := Coin(d[3].(string))        // commission asset
		commission, _ := d[4].(json.Number).Float64() // commission
		oid := d[5].(string)
		lhs := Coin(d[6].(string))
		rhs := Coin(d[7].(string))
		var side Side
		if amt > 0 {
			side = BUY
		} else {
			side = SELL
		}
		t, _ := time.Parse(time.RFC3339, d[0].(string))
		trd := TradeLog{
			oid,
			Pair{lhs, rhs},
			price,
			math.Abs(amt),
			commission,
			commissionAsset,
			t,
			side,
			"",
		}
		trades = append(trades, trd)
	}
	return trades
}

func getDealers(c client.Client, dbName string, filter map[string]string, timeFrom string, timeTo string) map[string]DealerInfo {
	query := "select OrderID,dealer,param,remark from " + MT_PLACED_ORDER +
		" where time >='" + timeFrom + "' and time <='" + timeTo + "'"
	for k, v := range filter {
		query += " and " + k + " = '" + v + "'"
	}
	fmt.Println(query)
	resp, err := queryDB(c, dbName, query)
	if err != nil {
		fmt.Println("getDearler error: " + err.Error())
		return nil
	}
	dealerInfos := make(map[string]DealerInfo)
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return dealerInfos
	}
	row := resp[0].Series[0]
	for _, d := range row.Values {
		oid, _ := d[1].(string)
		dealer, _ := d[2].(string)
		param, _ := d[3].(string)
		remark, _ := d[4].(string)
		dealerInfos[oid] = DealerInfo{dealer, param, remark}
	}
	return dealerInfos
}

func getBalances(c client.Client, dbName string, acct, exName string, coins []Coin, timeAt string) (res Portfolio) {
	res = NewPortfolio()
	if len(coins) == 0 {
		return
	}
	cn := "LAST(" + string(coins[0]) + ")"
	for i := 1; i < len(coins); i++ {
		cn += ",LAST(" + string(coins[i]) + ")"
	}
	query := "select " + cn + " from " + MT_COIN_BALANCE + " where time <='" + timeAt + "' and exchange = '" + exName + "' and aggregated = 'NO' and account = '" + acct + "'"
	fmt.Println(query)
	resp, err := queryDB(c, dbName, query)

	if err != nil {
		panic(err.Error())
	}
	if len(resp) <= 0 || len(resp[0].Series) <= 0 {
		return
	}
	row := resp[0].Series[0]
	if len(row.Values) != 1 {
		return
	}
	d := row.Values[0]
	for i, cn := range coins {
		amt, err := d[i+1].(json.Number).Float64()
		if err != nil {
			amt = 0.0
		}
		res.SetBalance(cn, amt)
	}
	return
}
