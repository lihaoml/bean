package mds

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

const MDS_DBNAME string = "MDS"
const MDS_PORT string = "8086"
const MT_ORDERBOOK string = "ORDERBOOK"
const MT_TRANSACTION string = "TRANSACTION"
const MT_TICK string = "TICK"

type MDS struct {
	c client.Client
}

func ConnectService() (MDS, error) {
	c, err := connect()
	defer c.Close()
	return MDS{c}, err
}

func connect() (client.Client, error) {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}
	dbhost := os.Getenv("MDS_DB_ADDRESS")
	if !strings.HasPrefix(dbhost, "http") {
		dbhost = "http://" + dbhost
	}
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     dbhost + ":" + MDS_PORT,
		Username: os.Getenv("MDS_USER"),
		Password: os.Getenv("MDS_PASSWORD"),
	})
	return c, err
}
