package bean

import (
	"bean"
	"log"
	"net/rpc"
	"time"
)

const (
	MDS_HOST_SG40 = "178.128.210.218"
	MDS_PORT      = "9892"
)

//////////////////////////////////////////////////////////////
// the RPC MDS connection client instance
type RPCMDSConnC struct {
	client *rpc.Client
}

// arguments to retrieve orderbook time series
type ArgOB struct {
	Pair  bean.Pair
	Start time.Time
	End   time.Time
	Depth int
}

type ArgTxn struct {
	Pair  bean.Pair
	Start time.Time
	End   time.Time
}

func NewRPCMDSConnC(network, address string) RPCMDSConnC {
	// Create a TCP connection to localhost on port 1234
	// client, err := rpc.DialHTTP("tcp", "localhost:9892")
	client, err := rpc.DialHTTP(network, address)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return RPCMDSConnC{client}
}

func (ex RPCMDSConnC) GetOrderBookTS(pair bean.Pair, start, end time.Time, depth int) (bean.OrderBookTS, error) {
	var err error
	var ob bean.OrderBookTS
	arg := ArgOB{pair, start, end, depth}
	err = ex.client.Call("RPCMDSConnD.GetOrderBookTS", arg, &ob)
	return ob, err
}

func (ex RPCMDSConnC) GetTransactions(pair bean.Pair, start, end time.Time) (bean.Transactions, error) {
	var err error
	var txn bean.Transactions
	arg := ArgTxn{pair, start, end}
	err = ex.client.Call("RPCMDSConnD.GetTransactions", arg, &txn)
	return txn, err
}
