package influx

import (
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
