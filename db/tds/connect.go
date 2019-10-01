package tds

import (
	"bean/db/influx"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/joho/godotenv"
	"os"
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

const TDS_DBNAME = "TDS"
const BALANCE_DBNAME = "BALANCE"

func connect() (client.Client, error) {
	godotenv.Load()
	// TODO: add https and password support for tds
	dbhost := "http://" + os.Getenv("TDS_DB_ADDRESS")
	fmt.Println("TDS connecting to ", dbhost)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: dbhost + ":" + TDS_PORT,
		// 		Username: username,
		//		Password: password,
	})
	return c, err
}

// FIXME: move it to a common module (used also in mds)
func queryDB(clnt client.Client, dbName, cmd string) (res []client.Result, err error) {
	res, err = influx.QueryDB(dbName, clnt, cmd)
	return
}
