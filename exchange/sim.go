package exchange

import (
	. "bean"
	"bean/rpc"
	"fmt"
	"math"
	"strconv"
	"time"
)

type Simulator struct {
	exName string
	now    time.Time
	last   time.Time

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
		obts[p], _ = mds.GetOrderBookTS(p, start, end, 20) // TODO: hard code 10 depth for now
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
		exName:      exName,
		now:         start,
		obts:        obts,
		txn:         txn,
		myOrders:    myOrders,
		oid:         0,
		myPortfolio: initPortfolio,
	}
}

func (sim Simulator) Name() string {
	return sim.exName
}

func (sim Simulator) GetOrderBook(pair Pair) OrderBook {
	return sim.obts[pair].GetOrderBook(sim.now)
}

func (sim Simulator) GetTransactionHistory(pair Pair) Transactions {
	return sim.txn[pair].Between(sim.now.Add(-10*time.Minute), sim.now) // exchanges typically give about 10mins of trade data
}

func (sim *Simulator) SetTime(t time.Time) {
	// update now
	for p := range sim.myOrders {
		for i, myOrder := range sim.myOrders[p] {
			if myOrder.status == ALIVE {
				// first see if the order can be filled against the immediate order book
				obFill := sim.GetOrderBook(p).Match(Order{Amount: myOrder.amount, Price: myOrder.price})

				txnFill := 0.0
				if math.Abs(obFill.Amount-myOrder.amount) > 0.0 {
					// if not then see if it can be filled against subsequent transactions
					recentTxn := sim.txn[p].Between(sim.now, t)
					txnFill = recentTxn.Fill(myOrder.price, myOrder.amount-obFill.Amount)
				}

				fillAmount := obFill.Amount + txnFill
				fillPrice := (obFill.Price*obFill.Amount + myOrder.price*txnFill) / fillAmount

				// determine if recentTxn crosses the order - updated for partial fills
				if fillAmount != 0.0 {
					if fillAmount == myOrder.amount {
						sim.myOrders[p][i].status = FILLED
					} else {
						sim.myOrders[p][i].amount -= fillAmount
					}
					// add it to myTransactions
					var maker TraderType
					if fillAmount > 0 {
						maker = Buyer
						currentLockedBase := sim.myPortfolio.Balance(p.Base) - sim.myPortfolio.AvailableBalance(p.Base)
						sim.myPortfolio.SetLockedBalance(p.Base, currentLockedBase-math.Abs(fillAmount)*fillPrice)
					} else {
						maker = Seller
						currentLockedCoin := sim.myPortfolio.Balance(p.Coin) - sim.myPortfolio.AvailableBalance(p.Coin)
						sim.myPortfolio.SetLockedBalance(p.Coin, currentLockedCoin-math.Abs(fillAmount))
					}
					sim.myPortfolio.AddBalance(p.Coin, fillAmount)
					sim.myPortfolio.AddBalance(p.Base, -fillAmount*fillPrice)
					newTxn := Transaction{
						Pair:      p,
						Price:     fillPrice,
						Amount:    fillAmount,
						TimeStamp: t,
						Maker:     maker,
						TxnID:     fmt.Sprint(len(sim.myTransactions)),
					}
					sim.myTransactions = append(sim.myTransactions, newTxn)

				}
			}
		}
	}
	sim.last = sim.now
	sim.now = t
}

type simOrder struct {
	oid       string
	price     float64
	amount    float64
	status    OrderState
	timeStamp time.Time
}

func (sim *Simulator) PlaceLimitOrder(pair Pair, price_ float64, amount float64) (string, error) {
	price, _ := strconv.ParseFloat(pair.OrderPricePrec(price_), 64)
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

	if amount > 0 {
		currentLockedBase := sim.myPortfolio.Balance(pair.Base) - sim.myPortfolio.AvailableBalance(pair.Base)
		sim.myPortfolio.SetLockedBalance(pair.Base, currentLockedBase+price*math.Abs(amount))
	} else {
		currentLockedCoin := sim.myPortfolio.Balance(pair.Coin) - sim.myPortfolio.AvailableBalance(pair.Coin)
		sim.myPortfolio.SetLockedBalance(pair.Coin, currentLockedCoin+math.Abs(amount))
	}

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

func (ex *Simulator) CancelAllOrders(pair Pair) {
	panic("not implemented")
}

func (sim Simulator) GetTrades() Transactions {
	return sim.myTransactions
}

func (sim Simulator) GetMyOrders(pair Pair) []OrderStatus {
	var ostatus []OrderStatus
	for _, o := range sim.myOrders[pair] {
		if o.status == ALIVE {
			os := OrderStatus{
				OrderID:      o.oid,
				PlacedTime:   o.timeStamp,
				Side:         AmountToSide(o.amount),
				FilledAmount: 0.0, // simulator cannot simulate partial fill for now
				LeftAmount:   math.Abs(o.amount),
				PlacedPrice:  o.price,
				Price:        o.price, // filled price, not applicable here
				State:        o.status,
			}
			ostatus = append(ostatus, os)
		}
	}
	return ostatus
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

func (sim Simulator) GetOrderStatus(orderID string, pair Pair) (OrderStatus, error) {
	var ostatus OrderStatus
	panic("not implemented")
	return ostatus, nil
}

func (ex Simulator) GetMyTrades(pair Pair, start, end time.Time) TradeLogS {
	panic("tradelog")
}

// dummy function, simulator doesn't need to trace the orders for each strategy separately
func (ex *Simulator) TrackOrderID(pair Pair, oid string) {
}

func (sim Simulator) GetAccountOrders(pair Pair) []OrderStatus {
	return sim.GetMyOrders(pair)
}
