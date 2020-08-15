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
const MT_CONTRACT_RISK string = "CONTRACT_RISK"
const MT_SMILE string = "SMILE"
const MT_TICK string = "TICK"
const MT_FUNDING_RATE string = "FUNDING_RATE"
const MT_FUNDING_RATE_DISPLAY string = "FUNDING_RATE_DISPLAY"
const MT_AV_OHLC_1m string = "AV_OHLC_1m" // data from alpha vantage

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
