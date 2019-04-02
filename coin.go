package bean

import (
	"bean/utils"
	"math"
	"strconv"
	"strings"
)

// Coin is a constant representing a token
// reason of using coin instead token is token is a builtin package in golang
type Coin string
type Coins []Coin

// token's names
const (
	BTC  Coin = "BTC"
	XBT  Coin = "XBT"
	USD  Coin = "USD"
	ETH  Coin = "ETH"
	USDT Coin = "USDT"
	XRP  Coin = "XRP"
	EOS  Coin = "EOS"
	LTC  Coin = "LTC"
	BCH  Coin = "BCH"

	IOTX Coin = "IOTX"
	ZRX  Coin = "ZRX"
	ONT  Coin = "ONT"
	ETC  Coin = "ETC"
	NEO  Coin = "NEO"
	IOTA Coin = "IOTA"
	ADA  Coin = "ADA"

	BGG Coin = "BGG"
	FT  Coin = "FT"
	HT  Coin = "HT"
	XMX Coin = "XMX"
	NKN Coin = "NKN"

	TRX  Coin = "TRX"
	MFT  Coin = "MFT"
	MITH Coin = "MITH"
	MDT  Coin = "MDT"
	BNB  Coin = "BNB"
	APOT Coin = "APOT"
)

func (s Coins) Len() int {
	return len(s)
}
func (s Coins) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Coins) Less(i, j int) bool {
	return s[i] < s[j]
}

func (c Coin) Format() string {
	switch c {
	case IOTX:
		return "#,###."
	case ETH:
		return "#,###.###"
	case USDT:
		return "#,###.##"
	case BTC:
		return "#,###.####"
	default:
		return "#,###.####"
	}
}

func FormatProfit(v float64, base Coin) string {
	var prec int = 6
	symb := strings.ToLower(string(base))
	switch base {
	case ETH:
		prec = 5
		symb = "Ξ"
		break
	case USDT:
		prec = 2
		symb = "$"
		break
	}
	return symb + " " + strconv.FormatFloat(v, 'f', prec, 64)
}

func (c Coin) RoundCoinAmount(amount float64) float64 {
	switch c {
	case IOTX:
		return math.Floor(amount)
	case ETH:
		return math.Floor(amount*1e3) / 1e3
	case ZRX:
		return math.Floor(amount)
	case ONT:
		return math.Floor(amount*1e2) / 1e2
	case TRX:
		return math.Floor(amount)
	case FT:
		return math.Floor(amount*1e2) / 1e2
	case MFT:
		return math.Floor(amount)
	case MDT:
		return math.Floor(amount)
	case BTC:
		return math.Floor(amount*1e6) / 1e6
	case ETC:
		return math.Floor(amount*1e6) / 1e6
	case EOS:
		return math.Floor(amount*1e6) / 1e6
	case NEO:
		return math.Floor(amount*1e6) / 1e6
	case ADA:
		return math.Floor(amount)
	default:
		//		logger.Fatal().Msg("RoundCoinAmount not implemented for " + string(coin))
		return math.NaN()
	}
}

func (coin Coin) RenderCoinAmount(amount float64) string {
	switch coin {
	case USDT:
		if math.Abs(amount) < 0.01 {
			return ""
		} else {
			return string("`$ `") + util.RenderFloat("#,###.##", amount)
		}
	case BNB:
		if math.Abs(amount) < 0.1 {
			return ""
		} else {
			return string("`BNB `") + util.RenderFloat("#,###.#", amount)
		}
	case BTC:
		if math.Abs(amount) < 1e-4 {
			return ""
		} else {
			return string("`฿ `") + util.RenderFloat("###.####", amount)
		}
	case ETH:
		return string("`Ξ `") + util.RenderFloat("###.###", amount)
	case IOTX:
		if amount == 0 {
			return ""
		} else {
			return string("`I `") + util.RenderFloat("#,###.", amount)
		}
	case ZRX:
		if amount == 0 {
			return ""
		} else {
			return string("`Z `") + util.RenderFloat("#,###.", amount)
		}
	case ONT:
		if amount == 0 {
			return ""
		} else {
			return string("`O `") + util.RenderFloat("#,###.", amount)
		}
	case FT:
		if amount < 10 {
			return ""
		} else {
			return string("`F `") + util.RenderFloat("#,###.", amount)
		}
	default:
		if amount < 1e-2 {
			return ""
		} else {
			return "`" + string(coin) + " `" + util.RenderFloat("#,###.", amount)
		}
	}
}
