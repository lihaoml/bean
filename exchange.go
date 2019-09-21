package bean

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// the supported exchanges
// NOTE: the names should all be uppercase
const (
	NameBinance  = "BINANCE"
	NameHuobi    = "HUOBI"
	NameHadax    = "HADAX"
	NameGate     = "GATE"
	NameKucoin   = "KUCOIN"
	NameFcoin    = "FCOIN"
	NameBgogo    = "BGOGO"
	NameBCoin    = "BCOIN"
	NameBittrex  = "BITTREX"
	NameDeribit  = "DERIBIT"
	NameBitMex   = "BITMEX"
	NameAllbit   = "ALLBIT"
	NameUpBit    = "UPBIT"
	NameFcoinC   = "FCOINC"
	NameFcoinM   = "FCOINM" // FCoin margin account, require MARGIN_PAIR in env, only single margin pair is supported
	NameGopax    = "GOPAX"
	NamePiexgo   = "PIEXGO"
	NameCodex    = "CODEX"
	NameElitex   = "ELITEX"
	NameBitMax   = "BITMAX"
	NameHotBit   = "HOTBIT"
	NameBilaxy   = "BILAXY"
	NameLBank    = "LBANK"
	NameBitfinex = "BITFINEX"
)

// Exchange is the interface for all exchanges
// Any exchange struct should implement all these interfaces
type Exchange interface {
	Name() string
	// get the open orders for a currency pair, when the exchange query fails, return an empty order book and log a warning message.
	GetOrderBook(pair Pair) OrderBook
	GetTicker(pair Pair) (Ticker, error) // get best bid/ask, should be a lightweight query compared to GetOrderBook
	GetLastPrice(pair Pair) (float64, error)
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
	GetMakerFee(pair Pair) float64
	GetTakerFee(pair Pair) float64
	GetKline(pair Pair, interval string, limit int) (OHLCVBSTS, error)
	/*  // other candidates
	// GetTransactionHistory returns a slice of past transaction, in ascending order
	MinimumTick(pair models.Pair) float64
	GetPairs(base models.Coin) []models.Pair  // get all pairs of the exchange with the given base
	*/
}

// exchange is a struct for holding common member variables and base functions
type BaseExchange struct {
	name string
	oids map[Pair]([]string)
}

func NewBaseExchange(name string) BaseExchange {
	return BaseExchange{
		name: name,
	}
}

func (ex BaseExchange) OrderIDS() map[Pair]([]string) {
	return ex.oids
}

func (ex BaseExchange) Name() string {
	return ex.name
}

func (ex *BaseExchange) TrackOrderID(pair Pair, oid string) {
	if ex.oids == nil {
		ex.oids = make(map[Pair]([]string))
	}
	ex.oids[pair] = append(ex.oids[pair], oid)
}

// interface function of base class
func (ex BaseExchange) GetMyOrders(pair Pair) []OrderStatus {
	panic(ex.name + "GetMyOrders(pair Pair)" + " not implemented")
}

func (ex BaseExchange) GetAccountOrders(pair Pair) []OrderStatus {
	panic(ex.name + "GetAccountOrders(pair Pair)" + " not implemented")
}

func (ex BaseExchange) GetMyTrades(pair Pair, start, end time.Time) TradeLogS {
	fmt.Println(ex.name + ".GetMyTrades(pair Pair, start, end time.Time)" + " not implemented")
	return TradeLogS{}
}

func (ex BaseExchange) CancelAllOrders(pair Pair) {
	panic(ex.name + " CancelAllOrders(pair Pair) not implemented")
}

func (ex BaseExchange) GetTransactionHistory(pair Pair) Transactions {
	panic(ex.name + " GetTransactionHistory(pair Pair) not implemented")
}

func (ex BaseExchange) GetOrderStatus(orderID string, pair Pair) (OrderStatus, error) {
	var ostatus OrderStatus
	panic(ex.name + " GetOrderStatus(orderID, pair) not implemented")
	return ostatus, nil
}

func (ex BaseExchange) GetMakerFee(pair Pair) float64 {
	panic(ex.name + " GetMakerFee(pair) not implemented")
	return math.NaN()
}

func (ex BaseExchange) GetTakerFee(pair Pair) float64 {
	panic(ex.name + " GetTakerFee(pair) not implemented")
	return math.NaN()
}

//NOTE: 'interval' pattern: 1m, 1h, 1d, 1w, 1M

func (ex BaseExchange) GetKline(pair Pair, interval string, limit int) (OHLCVBSTS, error) {
	panic(ex.name + "GetKline() not implemented")
	return nil, nil
}

func (ex BaseExchange) GetTicker(pair Pair) (Ticker, error) {
	err := errors.New(ex.Name() + ".GetTicker() not implemented")
	return Ticker{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, err
}

func (ex BaseExchange) GetLastPrice(pair Pair) (float64, error) {
	err := errors.New(ex.Name() + ".GetTicker() not implemented")
	return math.NaN(), err
}
