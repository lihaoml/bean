package main

import (
	. "bean"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	//database = "mydb"
	//username = "lu"
	//password = "password123"

	database = "BEANEX_BINANCE"
	username = ""
	password = ""
)

// DBPoint the structure to store one point from DB
type DBPoint struct {
	time   string
	amount float64
	price  float64
	index  string
}

func main() {

	// Create a new HTTPClient
	c := getDBClient()
	defer c.Close()

	measurement := "BTC_USDT_ASK"
	indexLimit := 10
	timeFrom := "2018-09-04T01:53:17Z"
	timeTo := "2018-09-04T01:53:25Z"

	orders := getOrders(c, measurement, timeFrom, timeTo, indexLimit)

	for key, val := range orders {
		fmt.Printf("%v:\n", key)
		for i, iv := range val {
			fmt.Printf("%v: %v,%v\n", i, iv.Price, iv.Amount)
		}

	}

	fmt.Println("OK")
}

// queryDB convenience function to query the database
func queryDB(clnt client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
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

func getDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://ss.w4ip.com:8086",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func getOrders(c client.Client, measurement string, timeFrom string, timeTo string, indexLimit int) map[string][]Order {
	if indexLimit < 1 {
		log.Fatal("index limit should be positive integer")
	}

	orders := make(map[string][]Order)
	for i := 0; i < indexLimit; i++ {
		order := getOrder(c, measurement, timeFrom, timeTo, strconv.Itoa(i))
		for key, val := range order {
			orders[key] = append(orders[key], val)
		}
	}

	return orders
}

func getOrder(c client.Client, measurement string, timeFrom string, timeTo string, index string) map[string]Order {
	query := "select * from " + measurement + " where time >='" + timeFrom + "' and time <='" + timeTo + "' and index = '" + index + "' limit 10"
	resp, err := queryDB(c, query)
	if err != nil {
		log.Fatal(err)
	}

	row := resp[0].Series[0]

	fmt.Printf("row.Name %s\n", row.Name)

	//row.Name BINANCE_BTC_USDT_ASK
	//row.Columns 0 : time
	//row.Columns 1 : Amount
	//row.Columns 2 : IDX
	//row.Columns 3 : Price
	//row.Columns 4 : index

	for i, col := range row.Columns {
		fmt.Printf("row.Columns %s : %s\n", strconv.Itoa(i), col)
	}

	for k, v := range row.Tags {
		fmt.Printf("k=%s, v=%s\n", k, v)
	}

	fmt.Printf("Partial = %v\n", row.Partial)

	for i, iv := range row.Values {
		for j, jv := range iv {
			fmt.Printf("[%v,%v]: %v\n", i, j, jv)
		}
	}

	var feed = make([]DBPoint, len(row.Values))
	for i, d := range row.Values {
		//fmt.Printf("%T,%T,%T,%T,%T\n", d[0], d[1], d[2], d[3], d[4])
		t1, _ := d[1].(json.Number).Float64()
		t2, _ := d[3].(json.Number).Float64()
		feed[i] = DBPoint{time: d[0].(string), amount: t1, price: t2, index: d[4].(string)}
	}

	dborder := convertToOrders(feed)
	for key, val := range dborder {
		fmt.Printf("%v: %v %v\n", key, val.Price, val.Amount)
	}

	fmt.Println("OK")
	return dborder
}

func convertToOrders(feed []DBPoint) map[string]Order {

	dborders := make(map[string]Order)
	for _, v := range feed {
		dborders[v.time] = Order{Amount: v.amount, Price: v.price}
	}

	return dborders

}
