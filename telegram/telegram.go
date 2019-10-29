package telegram

import (
	. "bean"
	"bean/logger"
	"bean/utils"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// webhook address and port
const (
	WebHook_Addr = "WEBHOOK_ADDRESS"
	WebHook_Port = "WEBHOOK_PORT"
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

// moved from notifictaion/telegram/tele.go

func SendMsg(sender string, receiver string, msg string) {
	godotenv.Load(BeanexAccountPath() + "tgbot.env")
	godotenv.Overload()
	if os.Getenv(sender) == "" {
		panic("Error loading sender token")
	}
	bot, err := tgbotapi.NewBotAPI(os.Getenv(sender))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	uids_env := os.Getenv(receiver)
	// parse it
	uids_ := strings.Split(uids_env, ",")
	var uids []int64
	for _, v := range uids_ {
		uids = append(uids, util.ParseIntSafe64(v))
	}
	for _, u := range uids {
		msg := tgbotapi.NewMessage(u, msg)
		msg.ParseMode = tgbotapi.ModeMarkdown
		bot.Send(msg)
	}
}

func SendFigs(sender string, receiver string, figpath string) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv(sender))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	uids_env := os.Getenv(receiver)
	// parse it
	uids_ := strings.Split(uids_env, ",")
	var uids []int64
	for _, v := range uids_ {
		uids = append(uids, util.ParseIntSafe64(v))
	}
	for _, u := range uids {
		msg := tgbotapi.NewPhotoUpload(u, figpath)
		bot.Send(msg)
	}
}

func SendFile(sender string, receiver string, filepath string) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv(sender))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	uids_env := os.Getenv(receiver)
	// parse it
	uids_ := strings.Split(uids_env, ",")
	var uids []int64
	for _, v := range uids_ {
		uids = append(uids, util.ParseIntSafe64(v))
	}
	for _, u := range uids {
		msg := tgbotapi.NewDocumentShare(u, filepath)
		bot.Send(msg)
	}
}

//this part will need kep.pem and pub.pem to setup a webhook for receiving message.
// Here we will use self-signed certificate to set webhook. See more: https://core.telegram.org/bots/self-signed
func TGWebhook(servant string) tgbotapi.UpdatesChannel {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv(servant))
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	link := os.Getenv(WebHook_Addr) + os.Getenv(WebHook_Port) + `/`
	_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert(link+bot.Token, "cert.pem"))
	if err != nil {
		log.Fatal(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	updates := bot.ListenForWebhook("/" + bot.Token)
	addr := "0.0.0.0" + `:` + os.Getenv(WebHook_Port)
	go http.ListenAndServeTLS(addr, "cert.pem", "key.pem", nil)
	//for update := range updates {
	//	log.Printf("%+v\n", update)
	//}
	return updates
}

const maxMsgPerMin = 20

// SendChannel opens a buffered string channel and listens for messages on that channel which it relays
// to the telegram bot. Stop the bot with a true on the unbuffered stop channel
func NotifyChannel(sender string, receiver string) (teleChan chan string, stop chan bool, err error) {
	var bot *tgbotapi.BotAPI
	bot, err = tgbotapi.NewBotAPI(os.Getenv(sender))
	if err != nil {
		log.Print(err)
		return
	}

	notifyID, err := strconv.ParseInt(os.Getenv(receiver), 10, 64)
	if err != nil {
		fmt.Println(err)
		log.Panic(err)
	}

	teleChan = make(chan string, 100)
	stop = make(chan bool)

	go func() {
		var msgCount int64
		var msgMin int
		for {
			select {
			case msgText := <-teleChan:
				log.Print(msgText)
				if msgMin == time.Now().Minute() {
					msgCount++
				} else {
					msgCount = 0
					msgMin = time.Now().Minute()
				}
				// Limit the maximum number of messages in a minute
				if msgCount < maxMsgPerMin {
					bot.Send(tgbotapi.NewMessage(notifyID, msgText))
				} else if msgCount == maxMsgPerMin {
					bot.Send(tgbotapi.NewMessage(notifyID, "..."))
				}
			case <-stop:
				// empty message channel first
				log.Print("Teleprinter instructed to stop")
				for len(teleChan) > 0 {
					msgText := <-teleChan
					log.Print(msgText)
					bot.Send(tgbotapi.NewMessage(notifyID, msgText))
				}
				log.Print("Teleprinter stopping")
				return
			}
		}
	}()

	return
}
