package strats

import (
	. "bean"
	"fmt"
	"math"
	"os"
	"time"
)

// we look at the price in a large amount using the stack and use this to bias our price
// in a small amount. Add a bias based on the accumulated position

type StackBiasMM struct {
	pair               Pair
	exName             string
	tick               time.Duration
	largeAmount        float64 // this is the notional used to inspect the stack
	largeBiasFactor    float64 // this is how much the price is biased by the stack. 0 is unaffected. 1 means mid
	tradingAmount      float64 // size we will trade in i.e. the small size in the stack
	maxPosition        float64 // this is the large position long or short it will take
	positionBiasFactor float64 // 0 to 1, value 1 means when position is maximum, price is biased by 1x bid offer spread
	widener            float64 // factor applied to averagespread to represent our width
	wideSpread         float64 // do not deal if the wide spread is wider than this
	dumpFile           *os.File
	spreadHistory      []float64
}

func NewStackBiasMM(exName string, pair Pair, tick time.Duration, tradingamount, maxposition, largebiasfactor, positionbiasfactor float64, dump bool) *StackBiasMM {
	var f *os.File
	if dump {
		f, _ = os.Create("stackbiasanal.csv")
		fmt.Fprintf(f, "Pos,LargeBid,LargeAsk,LargeBidAmount,LargeAskAmount,")
		fmt.Fprintf(f, "TradingBid,TradingAsk,TradingBidAmount,TradingAskAmount,")
		fmt.Fprintf(f, "LargeBias,PositionBias,myBid,myAsk\n")
	} else {
		f = nil
	}
	spreadhistory := make([]float64, 0, 10)
	return &StackBiasMM{
		exName:             exName,
		pair:               pair,
		tick:               tick,
		largeAmount:        10.0,
		largeBiasFactor:    largebiasfactor,
		tradingAmount:      tradingamount,
		maxPosition:        maxposition,
		positionBiasFactor: positionbiasfactor,
		widener:            1.5,
		wideSpread:         10.0,
		dumpFile:           f,
		spreadHistory:      spreadhistory,
	}
}

func averageSlice(sl []float64) float64 {
	var i int
	var v float64
	vs := 0.0
	for i, v = range sl {
		vs = vs + v
	}
	return vs / float64(i+1)
}

func max(i, j int) int {
	if i >= j {
		return i
	}
	return j
}

func (s StackBiasMM) GetExchangeNames() []string {
	return []string{s.exName}
}

func (s StackBiasMM) GetPairs() []Pair {
	return []Pair{s.pair}
}

func (s StackBiasMM) GetTick() time.Duration {
	return s.tick
}

func (s *StackBiasMM) Grind(exs map[string]Exchange) []TradeAction {
	ex := exs[s.exName]
	var actions []TradeAction

	// cancel existing orders
	myorders := ex.GetMyOrders(s.pair)
	for _, ord := range myorders {
		if ord.State == ALIVE {
			actions = append(actions, CancelOrderAction(s.exName, s.pair, ord.OrderID))
		}
	}

	// get position
	port := ex.GetPortfolio()
	position := port.Balance(s.pair.Coin)

	// get price stack
	ob := ex.GetOrderBook(s.pair)
	if !ob.Valid() {
		return actions
	}

	if s.dumpFile != nil {
		fmt.Fprintf(s.dumpFile, "%.0f,", position)
	}

	// find 2way price from the stack for both trading size and large size. large size will determine a bias on the trading price
	largeBid, largeBidAmount := priceInAmount(s.largeAmount, ob.Bids)
	largeAsk, largeAskAmount := priceInAmount(s.largeAmount, ob.Asks)
	if s.dumpFile != nil {
		fmt.Fprintf(s.dumpFile, "%7.2f,%7.2f,%.0f,%.0f,", largeBid, largeAsk, largeBidAmount, largeAskAmount)
	}
	tradingBid, tradingBidAmount := priceInAmount(s.tradingAmount, ob.Bids)
	tradingAsk, tradingAskAmount := priceInAmount(s.tradingAmount, ob.Asks)
	if s.dumpFile != nil {
		fmt.Fprintf(s.dumpFile, "%7.2f,%7.2f,%.0f,%.0f,", tradingBid, tradingAsk, tradingBidAmount, tradingAskAmount)
	}

	// bias the trading bias according to the price in the large amount. if the stack does not have the full size, use a weighted measure
	tradingMid := (tradingBid + tradingAsk) / 2.0
	largeBias := ((largeBid*largeAskAmount+largeAsk*largeBidAmount)/(largeBidAmount+largeAskAmount) - tradingMid) * s.largeBiasFactor

	// track 10 sample moving average spread. our price and bias will be based on this
	s.spreadHistory = append(s.spreadHistory, tradingAsk-tradingBid)
	averageSpread := averageSlice(s.spreadHistory[max(0, len(s.spreadHistory)-10):len(s.spreadHistory)])

	var positionBias float64
	if largeAsk-largeBid < s.wideSpread {
		positionBias = -position / s.maxPosition * averageSpread * s.positionBiasFactor
	} else {
		positionBias = 0.0
	}

	// Add biases to Mid. But avoid mid crossing the small size 2way
	ourMid := tradingMid + largeBias + positionBias
	ourMid = math.Max(ourMid, tradingBid)
	ourMid = math.Min(ourMid, tradingAsk)

	// our spread is based on moving average spread * widener
	ourBid := ourMid - averageSpread/2.0*s.widener
	ourAsk := ourMid + averageSpread/2.0*s.widener

	if s.dumpFile != nil {
		fmt.Fprintf(s.dumpFile, "%3.1f,%3.1f,%7.2f,%7.2f\n", largeBias, positionBias, ourBid, ourAsk)
	}

	actSell := PlaceLimitOrderAction(s.exName, s.pair, ourAsk, -s.tradingAmount)
	actions = append(actions, actSell)
	actBuy := PlaceLimitOrderAction(s.exName, s.pair, ourBid, s.tradingAmount)
	actions = append(actions, actBuy)

	return actions
}
