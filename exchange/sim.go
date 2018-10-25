package exchange

import (
	. "bean"
	"bean/rpc"
	"fmt"
	"time"
)

type Simulator struct {
	exName string
	now    time.Time

	// historical data
	obts map[Pair]OrderBookTS
	txn  map[Pair]Transactions

	// simulated actions and deals
	myActions []TradeActionT
	// to review
	myOrders map[Pair]([]simOrder)
	// consider using a list of orderstatus
	myTransactions []Transaction
	oid            int
	myPortfolio    Portfolio
}

func NewSimulator(exName string, pairs []Pair, dbhost, dbport string, start, end time.Time, initPortfolio Portfolio) Simulator {
	// create MDS
	// mds := mds.NewMDS(exName, dbhost, dbport)
	// FIXME: be able to pass exName to RPC MDS, right now we only use Binance
	mds := bean.NewRPCMDSConnC("tcp", dbhost+":"+dbport)
	// get historical data
	obts := make(map[Pair]OrderBookTS, len(pairs))
	for _, p := range pairs {
		obts[p], _ = mds.GetOrderBookTS(p, start, end, 10) // TODO: hard code 10 depth for now
		// obts[p].ShowBrief()
	}
	txn := make(map[Pair]Transactions, len(pairs))
	for _, p := range pairs {
		txn[p], _ = mds.GetTransactions(p, start, end)
	}
	// construct exSim
	myOrders := make(map[Pair]([]simOrder))
	for _, p := range pairs {
		myOrders[p] = make([]simOrder, 0)
	}
	return Simulator{
		exName:   exName,
		now:      start,
		obts:     obts,
		txn:      txn,
		myOrders: myOrders,
		oid:      0,
	}
}

func (sim Simulator) Name() string {
	return sim.exName
}

func (sim Simulator) GetOrderBook(pair Pair) OrderBook {
	return sim.obts[pair].GetOrderBook(sim.now)
}

func (sim Simulator) GetTransactionHistory(pair Pair) Transactions {
	return sim.txn[pair].Upto(sim.now)
}

func (sim *Simulator) SetTime(t time.Time) {
	// update now
	for p, txn := range sim.txn {
		recentTxn := txn.Between(sim.now, t)
		for i, o := range sim.myOrders[p] {
			if o.status == ALIVE {
				// determine if recentTxn crosses the order
				if recentTxn.Cross(o.price, o.amount) {
					sim.myOrders[p][i].status = FILLED
					// add it to myTransactions
					var maker TraderType
					if o.amount > 0 {
						maker = Buyer
					} else {
						maker = Seller
					}
					newTxn := Transaction{
						Pair:      p,
						Price:     o.price,
						Amount:    o.amount,
						TimeStamp: t,
						Maker:     maker,
						TxnID:     fmt.Sprint(len(sim.myTransactions)),
					}
					sim.myTransactions = append(sim.myTransactions, newTxn)
				}
			}
		}
	}
	sim.now = t
}

type simOrder struct {
	oid       string
	price     float64
	amount    float64
	status    OrderState
	timeStamp time.Time
}

func (sim *Simulator) PlaceLimitOrder(pair Pair, price float64, amount float64) (string, error) {
	// record the action
	act := TradeActionT{
		Time:   sim.now,
		Action: PlaceLimitOrderAction(sim.exName, pair, price, amount),
	}
	sim.myActions = append(sim.myActions, act)

	// add a live order in to myOrders
	oid := fmt.Sprint(sim.oid)
	order := simOrder{
		oid:       oid,
		price:     price,
		amount:    amount,
		timeStamp: sim.now,
		status:    ALIVE,
	}
	sim.myOrders[pair] = append(sim.myOrders[pair], order)
	sim.oid++
	return oid, nil
}

func (sim *Simulator) CancelOrder(pair Pair, oid string) error {
	// record the action
	act := TradeActionT{
		Time:   sim.now,
		Action: CancelOrderAction(sim.exName, pair, oid),
	}
	sim.myActions = append(sim.myActions, act)

	// mark the live order as cancelled
	for i, _ := range sim.myOrders[pair] {
		if sim.myOrders[pair][i].oid == oid {
			sim.myOrders[pair][i].status = CANCELLED
		}
		// TODO: might need to handel invlid input
	}
	return nil
}

func (sim Simulator) GetTrades() Transactions {
	return sim.myTransactions
}

// GetPortfolio not implemented yet
func (sim Simulator) GetPortfolioByCoins(coins Coins) Portfolio {
	p := sim.GetPortfolio()
	return p.Filter(coins)
}

// GetPortfolio -
func (sim Simulator) GetPortfolio() Portfolio {
	return sim.myPortfolio
}
