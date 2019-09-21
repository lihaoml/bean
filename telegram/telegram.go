package telegram

import (
	. "bean"
	"bean/logger"
	"bean/utils"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"math"
	"os"
	"strconv"
	"strings"
)

/*
func SendMsg(text string) {
	ShakeMsg(text)
}

func ShakeMsg(text string) {
	SendMsgRaw("TG_SHAKE_BOT_TOKEN", text)
	SendMsgRaw("TG_SHAKEBOT_BOT_TOKEN", text)
}

func BoatMsg(text string) {
	SendMsgRaw("TG_BOAT_BOT_TOKEN", text)
}
*/

func SendCSV(s TradeLogS, pair Pair, filename string) {
	bot, _ := tgbotapi.NewBotAPI(os.Getenv("TG_TRADE_SUMMARY_BOT_TOKEN"))
	uids_env := os.Getenv("TG_TRADE_SUMMARY_RECIPIENT")
	// parse it
	uids_ := strings.Split(uids_env, ",")
	var uids []int64
	for _, v := range uids_ {
		uids = append(uids, util.ParseIntSafe64(v))
	}
	s.ToCSV(filename)
	for _, v := range uids {
		docConfig := tgbotapi.NewDocumentUpload(v, filename)
		bot.Send(docConfig)
	}
}

func ReportError(err error) {
	if err != nil {
		uids_env := os.Getenv("TG_TRADE_SUMMARY_RECIPIENT")
		// parse it
		uids_ := strings.Split(uids_env, ",")
		var uids []int64
		for _, v := range uids_ {
			uids = append(uids, util.ParseIntSafe64(v))
		}
		SendMsgRaw("TG_TRADE_SUMMARY_BOT_TOKEN", uids, err.Error())
	}
}

func ReportTradeActions(ts []TradeAction) {
	if len(ts) == 0 {
		return
	}
	uids_env := os.Getenv("TG_TRADE_SUMMARY_RECIPIENT")
	// parse it
	uids_ := strings.Split(uids_env, ",")
	var uids []int64
	for _, v := range uids_ {
		uids = append(uids, util.ParseIntSafe64(v))
	}

	msg := ""
	for _, t := range ts {
		msg += t.Show()
	}
	SendMsgRaw("TG_TRADE_SUMMARY_BOT_TOKEN", uids, msg)
}

func ReportTradeSummary(s TradeLogSummary, periodDesc string) {
	ReportTradeSummaryWithMid(s, periodDesc, 0)
}

func ReportTradeSummaryWithMid(s TradeLogSummary, periodDesc string, mid float64) {
	// uids := []int64{578987011} // , 130320527}
	uids_env := os.Getenv("TG_TRADE_SUMMARY_RECIPIENT")
	// parse it
	uids_ := strings.Split(uids_env, ",")
	var uids []int64
	for _, v := range uids_ {
		uids = append(uids, util.ParseIntSafe64(v))
	}
	msg := FormatTradeSummary(s, periodDesc, mid)

	SendMsgRaw("TG_TRADE_SUMMARY_BOT_TOKEN", uids, msg)
}

func ReportPortfolio(port Portfolio, pnl_msg string) {
	uids_env := os.Getenv("TG_TRADE_SUMMARY_RECIPIENT")
	uids_ := strings.Split(uids_env, ",")
	var uids []int64
	for _, v := range uids_ {
		uids = append(uids, util.ParseIntSafe64(v))
	}
	relevantCoins := port.Coins()
	msg := renderEx(relevantCoins, port)
	msg += pnl_msg
	SendMsgRaw("TG_TRADE_SUMMARY_BOT_TOKEN", uids, msg)
}

// FIXME: this function is duplicated
func renderEx(relevantCoins []Coin, port Portfolio) string {
	msg := "--- position ---\n"
	for _, c := range relevantCoins {
		if math.Abs(port.Balance(c)) > 0 {
			msg += c.RenderCoinAmount(port.Balance(c)) + "\n"
		}
	}
	return msg
}

// show unrealized pnl if mid isn't 0
func FormatTradeSummary(s TradeLogSummary, periodDesc string, mid float64) string {
	vol := " (" + periodDesc + ", " + "v: " + fmt.Sprintf("%.2f", s.BuyValue+s.SellValue) + " " + string(s.Pair.Base) + ")"
	msg := "*" + string(s.Pair.Coin) + string(s.Pair.Base) + "*" + vol + " \n"
	msg += "`b: " + s.Pair.FormatAvgPrice(s.AvgBuyPrice()) + " " + util.RenderFloat(s.Pair.Coin.Format(), s.BuyAmount) + "`\n"
	msg += "`s: " + s.Pair.FormatAvgPrice(s.AvgSellPrice()) + " " + util.RenderFloat(s.Pair.Coin.Format(), s.SellAmount) + "`\n"
	sn := "+"
	if s.BuyAmount < s.SellAmount {
		sn = ""
	}
	exposure := s.BuyAmount - s.SellAmount
	msg += "`net: " + sn + util.RenderFloat(s.Pair.Coin.Format(), exposure) + " " + string(s.Pair.Coin) + "`\n"
	realizedPnL := s.RealizedPL()
	msg += "`rpl: " + s.Pair.Base.RenderCoinAmount(realizedPnL) + "`\n"

	if mid > 0 {
		msg += "`upl: " + s.Pair.Base.RenderCoinAmount(s.UnrealizedPL(mid)) + ", mid: " + s.Pair.FormatAvgPrice(mid) + "` \n"
	}

	msg += "`fee: "
	for c, a := range s.Fee {
		msg += string(c) + " " + strconv.FormatFloat(a, 'f', 3, 64) + " "
	}
	msg += "`"
	return msg
}

func SendMsgRaw(envKey string, uids []int64, text string) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv(envKey))
	if err != nil {
		logger.Warn().Msg(err.Error())
		return
	}
	for _, u := range uids {
		msg := tgbotapi.NewMessage(u, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		_, err := bot.Send(msg)
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}
