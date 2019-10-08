package mds

import (
	"bean"
	util "bean/utils"
	"errors"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

const MDS_DBNAME string = "MDS"
const MDS_PORT string = "8086"
const MT_ORDERBOOK string = "ORDERBOOK"
const MT_ORDERBOOK_STATS string = "ORDERBOOK_STATS"
const MT_TRANSACTION string = "TRANSACTION"
const MT_CONTRACT_ORDERBOOK string = "CONTRACT_ORDERBOOK"
const MT_CONTRACT_TRANSACTION string = "CONTRACT_TRANSACTION"
const MT_TICK string = "TICK"

type MDS struct {
	cs []client.Client // connecting to multiple MDS server (if provided), write - write to multiple servers, read - read from one that is available
}

func Connect(dbhost, port string) (MDS, error) {
	cs, err := connectTo(dbhost, port)
	return MDS{cs}, err
}

func ConnectService() (MDS, error) {
	cs, err := connect()
	for _, c := range cs {
		defer c.Close()
	}
	return MDS{cs}, err
}

func (m MDS) WriteBatchPoints(bp client.BatchPoints) (err error) {
	errMsgs := []string{}
	for _, c := range m.cs {
		err = c.Write(bp)
		if err != nil {
			errMsgs = append(errMsgs, "err writing db points "+err.Error())
		}
	}
	if len(errMsgs) == len(m.cs) {
		if len(m.cs) == 0 {
			err = errors.New("no MDS connection established")
		} else {
			err = errors.New(fmt.Sprint(errMsgs))
		}
	}
	return
}

// remember to defer c.Close() for every call of connect(), otherwise influx will open up too many files and stops working
func connect() ([]client.Client, error) {
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

func connectTo(dbhost, port string) (cs []client.Client, err error) {
	hosts := strings.Split(dbhost, ",")
	usrs := strings.Split(os.Getenv("MDS_USER"), ",")
	if len(usrs) <= 1 {
		usrs = util.ReplicateString(os.Getenv("MDS_USER"), len(hosts))
	}

	pwds := strings.Split(os.Getenv("MDS_PASSWORD"), ",")
	if len(pwds) <= 1 {
		pwds = util.ReplicateString(os.Getenv("MDS_PASSWORD"), len(hosts))
	}
	if !(len(hosts) == len(usrs) && len(hosts) == len(pwds)) {
		return nil, errors.New("number of MDS hosts do not match with number of users & pwds")
	}
	errMsg := []string{}
	for i, host := range hosts {
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		c, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     host + ":" + MDS_PORT,
			Username: usrs[i],
			Password: pwds[i],
		})
		if err != nil {
			errMsg = append(errMsg, host+": "+err.Error())
		} else {
			cs = append(cs, c)
		}
	}
	if len(errMsg) > 0 {
		err = errors.New(fmt.Sprint(errMsg))
	}
	return cs, err
}
