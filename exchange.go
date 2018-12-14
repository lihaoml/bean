package bean

import "time"

// the supported exchanges
// NOTE: the names should all be uppercase
const (
	NameBinance = "BINANCE"
	NameHuobi   = "HUOBI"
	NameHadax   = "HADAX"
	NameGate    = "GATE"
	NameKucoin  = "KUCOIN"
	NameFcoin   = "FCOIN"
	NameBgogo   = "BGOGO"
	NameBCoin   = "BCOIN"
	NameDeribit = "DERIBIT"
)

// Exchange is the interface for all exchanges
// Any exchange struct should implement all these interfaces
type Exchange interface {
	Name() string
	// get the open orders for a currency pair, when the exchange query fails, return an empty order book and log a warning message.
	GetOrderBook(pair Pair) OrderBook
	GetTransactionHistory(pair Pair) Transactions
	// get coin blanaces on the exchange
	GetPortfolio() Portfolio
	GetPortfolioByCoins(coins Coins) Portfolio
	// if amount is positive then it's a buy order
	// if amount is negative then it's a sell order
	PlaceLimitOrder(pair Pair, price float64, amount float64) (string, error) // return the orderid of the trade
	CancelOrder(pair Pair, orderID string) error                              // cancel the order
	GetOrderStatus(orderID string, pair Pair) (OrderStatus, error)

	// get our open orders for a currency pair, when the exchange query fails, return an empty order book and log a warning message.
	GetMyOrders(pair Pair) []OrderStatus      // returns the current open orders of the exchange instance (tracked by TrackOrderID() ), use with care
	GetAccountOrders(pair Pair) []OrderStatus // returns the open orders of the account

	CancelAllOrders(pair Pair)
	GetMyTrades(pair Pair, start, end time.Time) TradeLogS
	TrackOrderID(pair Pair, oid string)
	//////////////////////////////////////////////////////////////////////////////////////
	/*  // other candidates
	// GetTransactionHistory returns a slice of past transaction, in ascending order
	GetMakerFee(pair models.Pair) float64
	GetTakerFee(pair models.Pair) float64
	MinimumTick(pair models.Pair) float64
	GetPairs(base models.Coin) []models.Pair  // get all pairs of the exchange with the given base
	*/
}
