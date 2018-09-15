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
)

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
