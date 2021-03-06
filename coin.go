package bean

import (
	util "bean/utils"
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
	BTC Coin = "BTC"
	XBT Coin = "XBT"
	USD Coin = "USD"
	ETH Coin = "ETH"
	DOT Coin = "DOT"

	XRP Coin = "XRP"
	EOS Coin = "EOS"
	LTC Coin = "LTC"
	BCH Coin = "BCH"
	XLM Coin = "XLM"
	ZEC Coin = "ZEC"

	USDT Coin = "USDT"
	USDC Coin = "USDC"
	TUSD Coin = "TUSD"
	PAX  Coin = "PAX"
	BUSD Coin = "BUSD"

	IOTX Coin = "IOTX"
	ZRX  Coin = "ZRX"
	ONT  Coin = "ONT"
	ETC  Coin = "ETC"
	NEO  Coin = "NEO"
	IOTA Coin = "IOTA"
	ADA  Coin = "ADA"
	DASH Coin = "DASH"

	FT  Coin = "FT"
	HT  Coin = "HT"
	XMX Coin = "XMX"
	NKN Coin = "NKN"
	KRW Coin = "KRW"

	TRX     Coin = "TRX"
	MFT     Coin = "MFT"
	MITH    Coin = "MITH"
	MDT     Coin = "MDT"
	BNB     Coin = "BNB"
	APOT    Coin = "APOT"
	GT      Coin = "GT"
	FMEX    Coin = "FMEX"
	DUO     Coin = "DUO"
	AITFACE Coin = "AITFACE"
	DOGE    Coin = "DOGE"
	MCO     Coin = "MCO"
	OMG     Coin = "OMG"
	ENJ     Coin = "ENJ"
	PROB    Coin = "PROB"
	GNTO    Coin = "GNTO" // GoldNugget
	DGX     Coin = "DGX"
	VND     Coin = "VND"
	IOST    Coin = "IOST"
	THETA   Coin = "THETA"
	MATIC   Coin = "MATIC"
	ATOM    Coin = "ATOM"
	BAT     Coin = "BAT"
	VET     Coin = "VET"
	QTUM    Coin = "QTUM"
	LINK    Coin = "LINK"
	XMR     Coin = "XMR"
	COMP    Coin = "COMP"
	XTZ     Coin = "XTZ"
	WAVES   Coin = "WAVES"
	UNI     Coin = "UNI"
	PAXG    Coin = "PAXG"
	VITA    Coin = "VITA"
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
	case USDT:
		prec = 3
		symb = "$"
	case PAX:
		prec = 3
		symb = "$"
	case USDC:
		prec = 3
		symb = "$"
	case TUSD:
		prec = 3
		symb = "$"
	case BTC:
		prec = 6
		symb = "฿"
	}
	return symb + " " + strconv.FormatFloat(v, 'f', prec, 64)
}

func (c Coin) RoundCoinAmount(amount float64) float64 {
	sgn := 1.0
	if amount < 0 {
		sgn = -1.0
	}
	amtAbs := math.Abs(amount)
	switch c {
	case IOTX:
		return math.Floor(amtAbs) * sgn
	case DUO:
		return math.Floor(amtAbs) * sgn
	case ETH:
		return math.Floor(amtAbs*1e3) / 1e3 * sgn
	case ZRX:
		return math.Floor(amtAbs) * sgn
	case ONT:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case TRX:
		return math.Floor(amtAbs) * sgn
	case FT:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case MFT:
		return math.Floor(amtAbs) * sgn
	case MDT:
		return math.Floor(amtAbs) * sgn
	case BTC:
		return math.Floor(amtAbs*1e6) / 1e6 * sgn
	case ETC:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case EOS:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case NEO:
		return math.Floor(amtAbs*1e6) / 1e6 * sgn
	case ADA:
		return math.Floor(amtAbs) * sgn
	case DASH:
		return math.Floor(amtAbs*1e5) / 1e5 * sgn
	case BNB:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case HT:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case PAX, BUSD, USDC, USDT:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case DGX:
		return math.Floor(amtAbs*1e3) / 1e3 * sgn
	case XRP:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case DOGE:
		return math.Floor(amtAbs) * sgn
	case MCO:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case OMG:
		return math.Floor(amtAbs*1e2) / 1e2 * sgn
	case ENJ:
		return math.Floor(amtAbs) * sgn
	default:
		//		logger.Fatal().Msg("RoundCoinAmount not implemented for " + string(coin))
		return math.NaN()
	}
}

func (coin Coin) RenderCoinAmount(amount float64) string {
	switch coin {
	case USDT, USDC, USD, TUSD, PAX:
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
		if math.Abs(amount) < 1e-4 {
			return ""
		} else {
			return string("`Ξ `") + util.RenderFloat("###.###", amount)
		}
	case DGX:
		return string("`G `") + util.RenderFloat("###.###", amount)
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
		if math.Abs(amount) < 1e-2 {
			return ""
		} else {
			return "`" + string(coin) + " `" + util.RenderFloat("#,###.", amount)
		}
	}
}
