package tds

import (
	"bean/db/influx"
	"github.com/influxdata/influxdb/client/v2"
)

const TDS_PORT string = "8086"

// TDS DB names and measurements
const MT_PLACED_ORDER = "PLACED_ORDER"
const MT_OPEN_ORDER = "OPEN_ORDER"
const MT_COIN_BALANCE = "COIN_BALANCE"
const MT_TRADE = "TRADE"
const MT_PRINCIPAL = "PRINCIPAL"
const MT_TOTAL_BALANCE = "TOTAL_BALANCE"
const MT_MARGIN_ACCOUNT_INFO = "MARGIN_ACCOUNT_INFO"
const MT_PNL_BALANCE = "PNL_BALANCE"
const MT_OPEN_POSITION = "OPEN_POSITION"
const MT_MTM = "MTM"

const TDS_DBNAME = "TDS"
const BALANCE_DBNAME = "BALANCE"

type TDS struct {
	cs []client.Client // connecting to multiple TDS server (if provided), write - write to multiple servers, read - read from one that is available
}

// remember to defer c.Close() for every call of connect(), otherwise influx will open up too many files and stops working
func connect() ([]client.Client, error) {
	return influx.ConnectService("TDS_DB_ADDRESS", TDS_PORT, "TDS_USER", "TDS_PASSWORD")
}

// FIXME: move it to a common module (used also in mds)
func queryDB(clnt client.Client, dbName, cmd string) (res []client.Result, err error) {
	res, err = influx.QueryDB(dbName, clnt, cmd)
	return
}
