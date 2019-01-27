package bean

import (
	"bean"
	"fmt"
	"math"
	"strings"
	"time"
)

func PositionSummary(mds RPCMDSConnC, t time.Time) string {
	ptf := bean.GetLiveContracts()

	var output strings.Builder
	pvSum, deltaSum, vegaSum := 0.0, 0.0, 0.0
	spotMid := SpotPrice(mds, t, ptf[0].Underlying())

	for i, opt := range ptf {
		futMid, optBid, optAsk := FutureOptionPrice(mds, t, opt)
		domRate := bean.USDiscountRate
		volBid := opt.ImpVol(t, spotMid, futMid, domRate, optBid)
		volAsk := opt.ImpVol(t, spotMid, futMid, domRate, optAsk)
		volMid := (volBid + volAsk) / 2.0

		pv := opt.PV(t, spotMid, futMid, domRate, volMid)
		delta := opt.Delta(t, spotMid, futMid, volMid, domRate)
		vega := opt.Vega(t, spotMid, futMid, volMid, domRate)

		pvSum += pv
		deltaSum += delta
		vegaSum += vega

		if i == 0 {
			fmt.Fprintf(&output, "%s   %6.1f\n", t.Format("Mon 02Jan06 15:04"), spotMid)
			fmt.Fprintf(&output, "Contract Qty\nPV(BTC) PV(USD) DELTA(BTC) VEGA(USD)\n")
		}
		fmt.Fprintf(&output, "%s %4.1f\n%6.3f %6.1f %6.3f %5.1f\n", opt.Name(), opt.Quantity(), pv/spotMid, pv, delta, vega)
	}
	fmt.Fprintf(&output, "TOTAL\n%6.3f %6.1f %6.3f %5.1f\n", pvSum/spotMid, pvSum, deltaSum, vegaSum)
	return output.String()
}

func RiskSummary(mds RPCMDSConnC, t time.Time) string {
	ptf := bean.GetLiveContracts()

	var output strings.Builder
	pvSum, deltaSum, gammaSum, vegaSum, thetaSum := 0.0, 0.0, 0.0, 0.0, 0.0
	spotMid := SpotPrice(mds, t, ptf[0].Underlying())

	for _, opt := range ptf {
		futMid, optBid, optAsk := FutureOptionPrice(mds, t, opt)
		domRate := bean.USDiscountRate
		volBid := opt.ImpVol(t, spotMid, futMid, domRate, optBid)
		volAsk := opt.ImpVol(t, spotMid, futMid, domRate, optAsk)
		volMid := (volBid + volAsk) / 2.0

		pv := opt.PV(t, spotMid, futMid, domRate, volMid)
		delta := opt.Delta(t, spotMid, futMid, volMid, domRate)
		vega := opt.Vega(t, spotMid, futMid, volMid, domRate)

		pvSum += pv
		deltaSum += delta
		vegaSum += vega
	}

	fmt.Fprintf(&output, "Spot:       %7.1f\n", spotMid)
	fmt.Fprintf(&output, "PV (BTC)    %6.3f\n", pvSum/spotMid)
	fmt.Fprintf(&output, "PV (USD)    %6.1f\n", pvSum)
	fmt.Fprintf(&output, "DELTA (BTC) %6.3f\n", deltaSum)
	fmt.Fprintf(&output, "GAMMA (BTC) %6.3f\n", gammaSum)
	fmt.Fprintf(&output, "VEGA (USD)  %6.1f\n", vegaSum)
	fmt.Fprintf(&output, "THETA (USD) %6.1f\n", thetaSum)

	return output.String()
}

func RiskLadder(mds RPCMDSConnC, t time.Time) string {
	ptf := bean.GetLiveContracts()
	var output strings.Builder
	spotBump := [10]float64{-0.50, -0.25, -0.10, -0.05, 0.0, 0.05, 0.10, 0.25, 0.50, 1.0}
	var pv, delta, vega [len(spotBump)]float64

	spotMid := SpotPrice(mds, t, ptf[0].Underlying())

	fmt.Fprintf(&output, "SPOT   PV     DELTA VEGA\n")
	for j, s := range spotBump {
		for _, opt := range ptf {
			futMid, optBid, optAsk := FutureOptionPrice(mds, t, opt)
			domRate := bean.USDiscountRate
			volBid := opt.ImpVol(t, spotMid, futMid, domRate, optBid)
			volAsk := opt.ImpVol(t, spotMid, futMid, domRate, optAsk)
			volMid := (volBid + volAsk) / 2.0

			pv[j] += opt.PV(t, (1.0+s)*spotMid, (1.0+s)*futMid, domRate, volMid)
			delta[j] += opt.Delta(t, (1.0+s)*spotMid, (1.0+s)*futMid, volMid, domRate)
			vega[j] += opt.Vega(t, (1.0+s)*spotMid, (1.0+s)*futMid, volMid, domRate)
		}
		fmt.Fprintf(&output, "%6.1f %6.0f %5.2f %4.1f\n", (1.0+s)*spotMid, pv[j], delta[j], vega[j])
	}

	return output.String()
}

func OptMarketSummary(mds RPCMDSConnC, t time.Time) string {
	benchmarkNames := []string{
		"BTC-29MAR19-2000-P",
		"BTC-29MAR19-2500-P",
		"BTC-29MAR19-3000-P",
		"BTC-29MAR19-3500-P",
		"BTC-29MAR19-3500-C",
		"BTC-29MAR19-4000-C",
		"BTC-29MAR19-4500-C",
		"BTC-29MAR19-5000-C",
		"BTC-29MAR19-6000-C"}

	benchmarkContracts := bean.ContractsFromNames(benchmarkNames, nil)

	var output strings.Builder
	spotMid := SpotPrice(mds, t, benchmarkContracts[0].Underlying())

	for i, opt := range benchmarkContracts {
		futMid, optBid, optAsk := FutureOptionPrice(mds, t, opt)
		domRate := bean.USDiscountRate
		volBid := opt.ImpVol(t, spotMid, futMid, domRate, optBid)
		volAsk := opt.ImpVol(t, spotMid, futMid, domRate, optAsk)
		if i == 0 {
			fmt.Fprintf(&output, "%s   %6.1f\n", t.Format("Mon 02Jan06 15:04"), spotMid)
			fmt.Fprintf(&output, "Vol          Prem (BTC)\n")
		}
		fmt.Fprintf(&output, "%s\n%5.1f/%5.1f   %6.4f/%6.4f\n", opt.Name(), volBid*100.0, volAsk*100.0, optBid, optAsk)
	}
	return output.String()
}

func ContractHistory(mds RPCMDSConnC, cName string) string {
	var output strings.Builder
	c := bean.ContractFromName(cName, 0.0)
	n := time.Now()
	en := time.Date(n.Year(), n.Month(), n.Day(), 10, 0, 0, 0, time.UTC)
	if en.After(n) {
		en = en.Add(-24 * time.Hour)
	}
	st := en.Add(-30 * 24 * time.Hour)

	for t := st; t.Before(en); t = t.Add(24 * time.Hour) {
		spotMid := SpotPrice(mds, t, c.Underlying())
		futMid, optBid, optAsk := FutureOptionPrice(mds, t, c)
		domRate := bean.USDiscountRate
		volBid := c.ImpVol(t, spotMid, futMid, domRate, optBid)
		volAsk := c.ImpVol(t, spotMid, futMid, domRate, optAsk)
		if t.Equal(st) {
			fmt.Fprintf(&output, "%s   %6.1f\n", c.Name(), spotMid)
		}
		fmt.Fprintf(&output, "%s %5.1f/%5.1f   %6.4f/%6.4f\n", t.Format("02Jan06"), volBid*100.0, volAsk*100.0, optBid, optAsk)
	}
	return output.String()
}

func SpotPrice(mds RPCMDSConnC, asof time.Time, p bean.Pair) float64 {
	st := asof.Add(time.Duration(-1) * time.Minute)
	en := asof.Add(time.Duration(1) * time.Minute)

	spotobts, _ := mds.GetOrderBookTS(p, st, en, 20)

	return midPriceAt(spotobts, asof)
}

func FutureOptionPrice(mds RPCMDSConnC, asof time.Time, c bean.Contract) (futMid, optionBid, optionAsk float64) {
	st := asof.Add(time.Duration(-1) * time.Minute)
	en := asof.Add(time.Duration(1) * time.Minute)

	futContract := c.Name()[0:11]

	futobts, _ := mds.GetFutOrderBookTS(futContract, st, en, 20)
	optobts, _ := mds.GetOptOrderBookTS(c.Name(), st, en, 20)

	futMid = midPriceAt(futobts, asof)
	optionBid, optionAsk = priceAt(optobts, asof)
	return
}
func midPriceAt(obts bean.OrderBookTS, fix time.Time) float64 {
	i := 0
	for ; i < len(obts) && obts[i].Time.Before(fix); i++ {
	}
	if i >= len(obts) {
		return math.NaN()
	}
	return (obts[i].OB.Bids[0].Price + obts[i].OB.Asks[0].Price) / 2.0
}

func priceAt(obts bean.OrderBookTS, fix time.Time) (bid, ask float64) {
	i := 0
	for ; i < len(obts) && obts[i].Time.Before(fix); i++ {
	}
	if i >= len(obts) {
		return math.NaN(), math.NaN()
	}
	bid = obts[i].OB.Bids[0].Price
	ask = obts[i].OB.Asks[0].Price
	return
}
