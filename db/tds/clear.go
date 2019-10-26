package tds

import (
	"bean/db/influx"
	"fmt"
	"time"
)

// GetTransactionHistory : get transaction time series
func ClearOpenOrders(accountName string, start, end time.Time, filter map[string]string) error {
	cs, err := connect()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	timeFrom := start.Format(time.RFC3339)
	timeTo := end.Format(time.RFC3339)
	cmd := "delete from " + MT_OPEN_ORDER + " where time >='" + timeFrom + "' and time <='" + timeTo + "'" + " and account = '" + accountName + "'"
	fmt.Println(cmd)
	for k, v := range filter {
		cmd += " and " + k + "='" + v + "'"
	}
	for _, c := range cs {
		c.Close()
		_, err = influx.QueryDB(TDS_DBNAME, c, cmd)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return err
}
