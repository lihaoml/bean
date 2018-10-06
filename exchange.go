package bean

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
)

// Exchange is the interface for all exchanges
// Any exchange struct should implement all these interfaces
type Exchange interface {
	Name() string
	// get the open orders for a currency pair, when the exchange query fails, return an empty order book and log a warning message.
	GetOrderBook(pair Pair) OrderBook
	GetTransactionHistory(pair Pair) Transactions
	/*
		// get our open orders for a currency pair, when the exchange query fails, return an empty order book and log a warning message.
		GetMyOrders(pair Pair) OrderBook
		GetTransactionHistory(pair models.Pair) []models.Transaction
		// GetTransactionHistory returns a slice of past transaction, in ascending order
		// if amount is positive then it's a buy order
		// if amount is negative then it's a sell order
		PlaceLimitOrder(pair models.Pair, price float64, amount float64, dealer string, remark string) (string, float64) // returns the orderID and filled percentage
		CancelOrder(orderID string, side string, pair models.Pair)
		CancelAllOrders(pair models.Pair)
		GetOrderStatus(orderID string, side string, pair models.Pair) (models.OrderStatus, error)
		// if coins is empty get all.
		GetPortfolio() models.Portfolio
		GetPortfolioByCoins(models.Coins) models.Portfolio

		GetMakerFee(pair models.Pair) float64
		GetTakerFee(pair models.Pair) float64
		MinimumTick(pair models.Pair) float64
		GetPairs(base models.Coin) []models.Pair  // get all pairs of the exchange with the given base
	*/
}

/*
// Exchange is the interface for all exchanges
// Any exchange struct should implement all these interfaces
type Exchange interface {
	// get name of exchange
	Name() string

	// get the open orders for a currency pair,
	// when the exchange query fails, return an empty order book.
	GetOrderBook(pair Pair) (OrderBook, error)

	// get all pairs of the exchange with the given base
	GetPairs(base Coin) ([]Pair, error)
}
*/
