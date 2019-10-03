package mds

import (
	"bean"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

const MDS_DBNAME string = "MDS"
const MDS_PORT string = "8086"
const MT_ORDERBOOK string = "ORDERBOOK"
const MT_TRANSACTION string = "TRANSACTION"
const MT_CONTRACT_ORDERBOOK string = "CONTRACT_ORDERBOOK"
const MT_CONTRACT_TRANSACTION string = "CONTRACT_TRANSACTION"
const MT_TICK string = "TICK"

type MDS struct {
	c client.Client
}

func Connect(dbhost, port string) (MDS, error) {
	c, err := connectTo(dbhost, port)
	return MDS{c}, err
}

func ConnectService() (MDS, error) {
	c, err := connect()
	defer c.Close()
	return MDS{c}, err
}

// remember to defer c.Close() for every call of connect(), otherwise influx will open up too many files and stops working
func connect() (client.Client, error) {
	err := godotenv.Load()
	if err != nil || os.Getenv("MDS_DB_ADDRESS") == "" {
		err = godotenv.Load(bean.BeanexAccountPath() + "db.env")
		if err != nil {
			panic(err.Error())
		}
	}
	dbhost := os.Getenv("MDS_DB_ADDRESS")
	return connectTo(dbhost, MDS_PORT)
}

func connectTo(dbhost, port string) (client.Client, error) {
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
