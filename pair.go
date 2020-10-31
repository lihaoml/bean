package bean

import (
	util "bean/utils"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Pair is representing a trading pair
type Pair struct {
	Coin Coin
	Base Coin
}

func (p Pair) String() string {
	return string(p.Coin) + string(p.Base)
}

func Parse(mpair string, sep string) Pair {
	cs := strings.Split(mpair, sep)
	return Pair{Coin(cs[0]), Coin(cs[1])}
}

// use
// go run exchange/example/binance_info/binance_info.go
// as guideline to set the minimum amount of a pair,
// look at LOT_SIZE and MIN_NOTIONAL / a conservatively low price
func (pair Pair) MinimumTradingAmount() float64 {
	switch pair {
	case Pair{IOTX, ETH}:
		return 300 // binance requires 0.01 ETH as minimum notional
	case Pair{IOTX, BTC}:
		return 500 // binance requires 0.001 BTC as minimum notional
	case Pair{ZRX, ETH}:
		return 8 // binance requires 0.01 ETH as minimum notional, approx price 0.0012
	case Pair{ETH, USDT}:
		return 0.05 // // binance requires 10 USDT as minimum notional, approx price 400
	case Pair{BTC, USDT}:
		return 0.002 // // binance requires 10 USDT as minimum notional, approx price 5000
	case Pair{ETH, BTC}:
		return 0.016 // // binance requires 0.001 BTC as minimum notional, approx price 0.06
	case Pair{TRX, BTC}:
		return 200 // // binance requires 0.001 BTC as minimum notional, approx price 500s
	case Pair{ONT, ETH}:
		return 6 // // binance requires 0.001 BTC as minimum notional, approx price 500s
	case Pair{ONT, USDT}:
		return 8 // // binance requires 0.001 BTC as minimum notional, approx price 500s

	case Pair{ETC, ETH}:
		return 0.3 // // binance requires 0.01 ETH as minimum notional
	case Pair{ETC, USDT}:
		return 1.0 // // binance requires 10 USDT as minimum notional, approx price 500s
	case Pair{ETC, BTC}:
		return 0.8 // // binance requires 0.001 BTC as minimum notional, approx price 500s

	case Pair{EOS, ETH}:
		return 2.0 // // binance requires 0.01 ETH as minimum notional
	case Pair{EOS, USDT}:
		return 3.0 // // binance requires 10 USDT as minimum notional, approx price 500s
	case Pair{EOS, BTC}:
		return 2.5 // // binance requires 0.001 BTC as minimum notional, approx price 500s

	case Pair{NEO, ETH}:
		return 0.6 // // binance requires 0.01 ETH as minimum notional
	case Pair{NEO, USDT}:
		return 0.8 // // binance requires 10 USDT as minimum notional, approx price 500s
	case Pair{NEO, BTC}:
		return 0.8 // // binance requires 0.001 BTC as minimum notional, approx price 500s

		// below are made up -------------------------
	case Pair{NKN, ETH}:
		return 1 // made up, not on binance
	case Pair{MITH, ETH}:
		return 1 // made up, not on binance
	case Pair{MDT, ETH}:
		return 1000 // made up, not on binance
	case Pair{HT, ETH}:
		return 1 // didn't check, not on binance
	case Pair{XMX, ETH}:
		return 100 // didn't check, not on binance
	case Pair{MFT, ETH}:
		return 200 // didn't check, not on binance
	case Pair{MFT, BTC}:
		return 200 // didn't check, not on binance
	case Pair{FT, USDT}:
		return 1.0 // // binance requires 0.001 BTC as minimum notional, approx price 500s

	case Pair{ADA, USDT}:
		return 15.0 // // binance requires 0.001 BTC as minimum notional, approx price 500s

	default:
		//logger.Warn().Interface("pair", pair).Msg("minimum trading amount not implemented use 1.0 by default")
		return 1.0
	}
}

// Format price for reporting purpose (not for trading)
func (pair Pair) FormatPrice(price float64) string {
	var prec int = 8
	switch pair {
	case Pair{ETH, USDT}, Pair{ETH, USDC}, Pair{ETH, PAX}:
		prec = 2
	case Pair{EOS, USDT}, Pair{EOS, USDC}, Pair{EOS, PAX}:
		prec = 3
	case Pair{PAX, USDT}, Pair{USDC, USDT}, Pair{TUSD, USDT}:
		prec = 4
	case Pair{BTC, USDT}, Pair{BTC, USDC}, Pair{BTC, PAX}, Pair{BTC, TUSD}:
		prec = 2

	case Pair{ETH, BTC}:
		prec = 6
	case Pair{ONT, USDT}:
		prec = 3
	case Pair{ONT, ETH}:
		prec = 6
	case Pair{IOTX, ETH}:
		return fmt.Sprintf("%.3e", price)
	}
	return strconv.FormatFloat(price, 'f', prec, 64)
}

// Format price for reporting purpose (not for trading)
func (pair Pair) FormatAvgPrice(price float64) string {
	return fmt.Sprintf("%.4e", price)
}

func orderPricePrec(pair Pair) (prec int) {
	switch pair {
	case Pair{ETH, USDT}, Pair{ETH, PAX}:
		prec = 2
	case Pair{BTC, USDT}, Pair{BTC, PAX}:
		prec = 2
	case Pair{ONT, USDT}:
		prec = 3
	case Pair{NEO, USDT}:
		prec = 3
	case Pair{ETC, USDT}:
		prec = 4
	case Pair{ETH, BTC}:
		prec = 6
	case Pair{ONT, ETH}:
		prec = 6
	case Pair{NEO, ETH}:
		prec = 6
	case Pair{NEO, BTC}:
		prec = 6
	case Pair{ETC, BTC}:
		prec = 6
	case Pair{ETC, ETH}:
		prec = 6
	case Pair{ONT, BTC}:
		prec = 7
	case Pair{BNB, BTC}:
		prec = 7
	case Pair{IOTX, ETH}:
		prec = 8
	case Pair{IOTX, BTC}:
		prec = 8
	case Pair{IOTX, USDT}:
		prec = 6
	case Pair{ZRX, ETH}:
		prec = 8
	case Pair{ZRX, BTC}:
		prec = 8
	case Pair{ZRX, USDT}:
		prec = 4
	case Pair{MFT, ETH}:
		prec = 8
	case Pair{TRX, BTC}:
		prec = 8
	case Pair{MFT, BTC}:
		prec = 8
	case Pair{FT, BTC}:
		prec = 8
	case Pair{FT, USDT}:
		prec = 5
	case Pair{FT, ETH}:
		prec = 8
	case Pair{EOS, BTC}:
		prec = 7
	case Pair{EOS, USDT}, Pair{EOS, PAX}:
		prec = 4 // fcoin uses 3
	case Pair{EOS, ETH}:
		prec = 6
	case Pair{XRP, USDT}:
		prec = 4
	case Pair{XRP, ETH}:
		prec = 8
	case Pair{XRP, BTC}:
		prec = 8
	case Pair{ADA, ETH}:
		prec = 8
	case Pair{ADA, BTC}:
		prec = 8
	case Pair{ADA, USDT}:
		prec = 5

	case Pair{DASH, USDT}:
		prec = 2
	case Pair{DASH, BTC}:
		prec = 6
	case Pair{DASH, ETH}:
		prec = 5
	case Pair{IOTX, KRW}:
		prec = 2

	case Pair{LTC, USDT}, Pair{LTC, PAX}:
		prec = 2
	case Pair{LTC, BTC}:
		prec = 6

	case Pair{PAX, USDT}, Pair{USDC, USDT}, Pair{BUSD, USDT}:
		prec = 4
	case Pair{DGX, USD}:
		prec = 2
	case Pair{DGX, BTC}:
		prec = 6
	case Pair{DGX, ETH}:
		prec = 5
	case Pair{DGX, VND}:
		prec = 0
	case Pair{USDT, VND}:
		prec = 0
	case Pair{ETH, VND}:
		prec = 0
	case Pair{BTC, VND}:
		prec = 0
	case Pair{BCH, USDT}:
		prec = 2
	case Pair{BCH, BTC}:
		prec = 5
	case Pair{XLM, USDT}:
		prec = 4
	case Pair{XLM, BTC}:
		prec = 8
	case Pair{XLM, ETH}:
		prec = 6
	case Pair{ZEC, USDT}:
		prec = 2
	case Pair{ZEC, BTC}:
		prec = 5
	case Pair{ZEC, ETH}:
		prec = 6

	case Pair{DUO, BTC}:
		prec = 8
	case Pair{DUO, USDT}:
		prec = 5
	case Pair{BTC, USD}:
		prec = 2
	case Pair{AITFACE, USDT}:
		prec = 4
	case Pair{GNTO, USDT}:
		prec = 3
	case Pair{DOGE, USDT}:
		prec = 6
	case Pair{HT, USDT}:
		prec = 4
	case Pair{BNB, USDT}:
		prec = 4
	case Pair{TRX, USDT}:
		prec = 5
	case Pair{MCO, USDT}:
		prec = 2
	case Pair{OMG, USDT}:
		prec = 4
	case Pair{ENJ, USDT}:
		prec = 4
	case Pair{PAXG, USDT}:
		prec = 2
	default:
		panic("pair.OrderPricePrec not implemented for " + string(pair.Coin) + string(pair.Base))
		return
	}
	return
}

// the minimum precision of all exchanges for price of a pair to place an order
func (pair Pair) OrderPricePrec(price float64) string {
	prec := orderPricePrec(pair)
	return strconv.FormatFloat(price, 'f', prec, 64)
}

func (pair Pair) MinimumTick() float64 {
	prec := orderPricePrec(pair)
	return math.Pow10(-prec)
}

func AllCoins(pairs []Pair) (res Coins) {
	for _, p := range pairs {
		if !util.Contains(res, p.Base) {
			res = append(res, p.Base)
		}
		if !util.Contains(res, p.Coin) {
			res = append(res, p.Coin)
		}
	}
	return res
}

func RightPair(coin1, coin2 Coin) (Pair, bool) {
	if coin1 == USDT || coin1 == USD || coin1 == USDC || coin1 == PAX {
		return Pair{coin2, coin1}, true
	}
	if coin2 == USDT || coin2 == USD || coin2 == USDC || coin2 == PAX {
		return Pair{coin1, coin2}, true
	}
	if coin2 == KRW {
		return Pair{coin1, coin2}, true
	}
	if coin1 == KRW {
		return Pair{coin2, coin1}, false
	}
	if coin1 == BTC {
		return Pair{coin2, coin1}, true
	}
	if coin2 == BTC {
		return Pair{coin1, coin2}, true
	}
	if coin1 == ETH {
		if coin2 != BCH {
			return Pair{coin2, coin1}, true
		} else {
			return Pair{}, false
		}

	}
	if coin2 == ETH {
		if coin1 != BCH {
			return Pair{coin1, coin2}, true
		} else {
			return Pair{}, false
		}
	}
	if (coin1 == IOTX && coin2 == APOT) || (coin1 == APOT && coin2 == IOTX) {
		return Pair{IOTX, APOT}, true
	}

	// logger.Warn().Msg(("unknow pair convention " + string(coin1) + string(coin2)))
	return Pair{}, false
}

// all possible pairs for a given list of coins
func PossiblePairs(coins []Coin) (pairs []Pair) {
	for i := 0; i < len(coins); i++ {
		if coins[i] == IOTX {
			pairs = append(pairs, Pair{IOTX, BTC})
			pairs = append(pairs, Pair{IOTX, USDT})
		} else {
			for j := i + 1; j < len(coins); j++ {
				if coins[j] != IOTX {
					p, valid := RightPair(coins[i], coins[j])
					if valid {
						pairs = append(pairs, p)
					}
				}
			}
		}
	}
	return
}

// split p by "_", then construct the pair, only work for string like BTC_USDT
func ParsePair(p string) Pair {
	s := strings.Split(p, "_")
	return Pair{Coin(strings.ToUpper(s[0])), Coin(strings.ToUpper(s[1]))}
}

// parse the underlying pair by symbol name, symbol follow below formats:
// BTCUSD-ContractSpec, e.g., BTCUSD-PERPETUAL, BTCUSDT-PERPETUAL, BTCUSD-25SEP20
// BTCUSD_ContractSpec, e.g., BTCUSD_PERPETUAL, BTCUSDT_PERPETUAL, BTCUSD_25SEP20
// BTCUSD, BTCUSDT, IOTXBTC
// contract spec will be thrown away, the first string will be split into two coins based on the following ruls
// look for USDT, USD, BTC at the right, trim it, then convert the left to a coin
func SymbolToPair(symbol string) Pair {
	s := strings.Split(strings.Split(symbol, "_")[0], "-")[0]
	if strings.HasSuffix(s, "USDT") {
		return Pair{Coin(strings.Trim(s, "USDT")), USDT}
	} else if strings.HasSuffix(s, "USD") {
		// TODO: if we want to support BUSD as the quote ccy, we need to differentiate it here
		return Pair{Coin(strings.Trim(s, "USD")), USD}
	} else if strings.HasSuffix(s, "BTC") {
		return Pair{Coin(strings.Trim(s, "BTC")), BTC}
	} else if strings.HasSuffix(s, "ETH") {
		return Pair{Coin(strings.Trim(s, "ETH")), ETH}
	} else {
		panic("unknown format of the symbol: " + symbol)
	}
}