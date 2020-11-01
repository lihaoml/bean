package mds

import (
	"bean/db/influx"

	"github.com/influxdata/influxdb/client/v2"
)

const MDS_DBNAME string = "MDS"
const MDS_PORT string = "8086"
const MT_ORDERBOOK string = "ORDERBOOK"
const MT_ORDERBOOK_STATS string = "ORDERBOOK_STATS"
const MT_TRANSACTION string = "TRANSACTION"
const MT_CONTRACT_ORDERBOOK string = "CONTRACT_ORDERBOOK"
const MT_CONTRACT_TRANSACTION string = "CONTRACT_TRANSACTION"
const MT_CONTRACT_BOOK_TICKER string = "CONTRACT_BOOK_TICKER"
const MT_CONTRACT_RISK string = "CONTRACT_RISK"
const MT_SMILE string = "SMILE"
const MT_TICK string = "TICK"
const MT_FUNDING_RATE string = "FUNDING_RATE"
const MT_FUNDING_RATE_DISPLAY string = "FUNDING_RATE_DISPLAY"
const MT_AV_OHLC_1m string = "AV_OHLC_1m" // data from alpha vantage
const MT_HOURLY_DATA string = "HOURLY_DATA" // data from alpha vantage
const MT_HOURLY_FACTOR string = "HOURLY_FACTOR" // data from alpha vantage
const MT_HOURLY_ALPHA string = "HOURLY_ALPHA" // data from alpha vantage

const MT_MIN_5_DATA string = "MIN_5_DATA" // data from alpha vantage
const MT_MIN_5_FACTOR string = "MIN_5_FACTOR" // data from alpha vantage
const MT_MIN_5_ALPHA string = "MIN_5_ALPHA" // data from alpha vantage

const MT_MIN_1_DATA string = "MIN_1_DATA" // data from alpha vantage
const MT_MIN_1_FACTOR string = "MIN_1_FACTOR" // data from alpha vantage
const MT_MIN_1_ALPHA string = "MIN_1_ALPHA" // data from alpha vantage

const MT_MIN_15_DATA string = "MIN_15_DATA" // data from alpha vantage
const MT_MIN_15_FACTOR string = "MIN_15_FACTOR" // data from alpha vantage
const MT_MIN_15_ALPHA string = "MIN_15_ALPHA" // data from alpha vantage

const MT_MIN_30_DATA string = "MIN_30_DATA" // data from alpha vantage
const MT_MIN_30_FACTOR string = "MIN_30_FACTOR" // data from alpha vantage
const MT_MIN_30_ALPHA string = "MIN_30_ALPHA" // data from alpha vantage

type MDS struct {
	cs []client.Client // connecting to multiple MDS server (if provided), write - write to multiple servers, read - read from one that is available
}

func Connect(dbhost, port string) (MDS, error) {
	cs, err := influx.ConnectTo(dbhost, port, "MDS_USER", "MDS_PASSWORD")
	//	for _, c := range cs {
	//		defer c.Close()
	//	}
	return MDS{cs}, err
}

func ConnectService() (MDS, error) {
	cs, err := connect()
	//	for _, c := range cs {
	//		defer c.Close()
	//	}
	return MDS{cs}, err
}

func (m MDS) WriteBatchPoints(bp client.BatchPoints) (err error) {
	return influx.WriteBatchPoints(m.cs, bp)
}

// remember to defer c.Close() for every call of connect(), otherwise influx will open up too many files and stops working
func connect() ([]client.Client, error) {
	return influx.ConnectService("MDS_DB_ADDRESS", MDS_PORT, "MDS_USER", "MDS_PASSWORD")
}

func (m MDS) Close() {
	for _, c := range m.cs {
		c.Close()
	}
}
