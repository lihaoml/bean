package influx

import (
	"errors"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
)

func QueryDB(dbName string, clnt client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: dbName,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func WriteBatchPoints(cs []client.Client, bp client.BatchPoints) (err error) {
	errMsgs := []string{}
	for _, c := range cs {
		err = c.Write(bp)
		if err != nil {
			errMsgs = append(errMsgs, "err writing db points "+err.Error())
		}
	}
	if len(errMsgs) == len(cs) {
		if len(cs) == 0 {
			err = errors.New("no influx connection established")
		} else {
			err = errors.New(fmt.Sprint(errMsgs))
		}
	}
	return
}

// get databases of a influx instance
func getDatabases(c client.Client) []string {
	query := "show databases"
	q := client.NewQuery(query, "", "ns")
	resp, err := c.Query(q)
	var dbNames []string
	if err == nil && resp.Results != nil && len(resp.Results) > 0 {
		nms := resp.Results[0].Series[0].Values
		for _, n := range nms {
			dbnm := n[0].(string)
			dbNames = append(dbNames, dbnm)
		}
	}
	return dbNames
}

// check total number of series in a database - useful for monitoring db
// select * from _internal.."database" where "database"='MDS' order by time desc limit 1
