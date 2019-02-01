package bean

import (
	"bean"
	"fmt"
	"math"
	"strings"
	"time"
)

// risk module generates various portfolio risk reports
// it sits in the mds section because it needs to go and get it's own market data - DOESN'T SEEM OPTIMAL TBH

// all reports need to be extended to describe the portfolio cash/coin risk better

func PortfolioPositionSummary(ptf bean.Portfolio, mds RPCMDSConnC, t time.Time) string {

	var output strings.Builder
	pvSum, pvUSDSum, deltaSum, vegaSum := 0.0, 0.0, 0.0, 0.0

	fmt.Fprintf(&output, "%s\n", t.Format("Mon 02Jan06 15:04"))
	fmt.Fprintf(&output, "Contract Qty\nPV(BTC) PV(USD) DELTA(BTC) VEGA(USD)\n")

	for c, q := range ptf.Contracts() {
		spotMid, futMid, domRate, volBid, volAsk := ContractMarket(mds, t, c)
		volMid := (volBid + volAsk) / 2.0

		pv := q * c.PV(t, spotMid, futMid, domRate, volMid)
		delta := q * c.Delta(t, spotMid, futMid, volMid, domRate)
		vega := q * c.Vega(t, spotMid, futMid, volMid, domRate)

		pvSum += pv
		pvUSDSum += pv * spotMid
		deltaSum += delta
		vegaSum += vega

		fmt.Fprintf(&output, "%s %4.1f\n%6.3f %6.1f %6.3f %5.1f\n", c.Name(), q, pv/spotMid, pv, delta, vega)
	}
	fmt.Fprintf(&output, "TOTAL\n%6.3f %6.1f %6.3f %5.1f\n", pvUSDSum, pvSum, deltaSum, vegaSum)
	return output.String()
}

func PortfolioRiskSummary(ptf bean.Portfolio, mds RPCMDSConnC, t time.Time) string {

	var output strings.Builder
	pvSum, deltaSum, gammaSum, vegaSum, thetaSum := 0.0, 0.0, 0.0, 0.0, 0.0

	// REALLY DIRTY. come back here.
	_, _, btcSpot := SpotPrice(mds, t, bean.Pair{bean.BTC, bean.USDT})
	pvSum = ptf.Balance(bean.BTC) * btcSpot
	deltaSum = ptf.Balance(bean.BTC)

	for c, q := range ptf.Contracts() {
		spotMid, futMid, domRate, volBid, volAsk := ContractMarket(mds, t, c)
		volMid := (volBid + volAsk) / 2.0

		pv := q * c.PV(t, spotMid, futMid, domRate, volMid)
		delta := q * c.Delta(t, spotMid, futMid, volMid, domRate)
		gamma := q * c.Gamma(t, spotMid, futMid, volMid, domRate)
		theta := q * c.Theta(t, spotMid, futMid, volMid, domRate)
		vega := q * c.Vega(t, spotMid, futMid, volMid, domRate)

		pvSum += pv
		deltaSum += delta
		vegaSum += vega
		gammaSum += gamma
		thetaSum += theta
	}

	fmt.Fprintf(&output, "Spot:       %7.1f\n", btcSpot)
	fmt.Fprintf(&output, "PV (BTC)    %6.3f\n", pvSum/btcSpot)
	fmt.Fprintf(&output, "PV (USD)    %6.1f\n", pvSum)
	fmt.Fprintf(&output, "DELTA (BTC) %6.3f\n", deltaSum)
	fmt.Fprintf(&output, "GAMMA (BTC) %6.3f\n", gammaSum)
	fmt.Fprintf(&output, "VEGA (USD)  %6.1f\n", vegaSum)
	fmt.Fprintf(&output, "THETA (USD) %6.1f\n", thetaSum)

	return output.String()
}

func PortfolioRiskLadder(ptf bean.Portfolio, mds RPCMDSConnC, t time.Time) string {
	var output strings.Builder
	spotBump := [10]float64{-0.50, -0.25, -0.10, -0.05, 0.0, 0.05, 0.10, 0.25, 0.50, 1.0}
	var pv, delta, vega [len(spotBump)]float64
	var spotMid float64

	fmt.Fprintf(&output, "SPOT   PV     DELTA VEGA\n")
	for j, s := range spotBump {
		for c, q := range ptf.Contracts() {
			spotMid, futMid, domRate, volBid, volAsk := ContractMarket(mds, t, c)
			volMid := (volBid + volAsk) / 2.0

			pv[j] += q * c.PV(t, (1.0+s)*spotMid, (1.0+s)*futMid, domRate, volMid)
			delta[j] += q * c.Delta(t, (1.0+s)*spotMid, (1.0+s)*futMid, volMid, domRate)
			vega[j] += q * c.Vega(t, (1.0+s)*spotMid, (1.0+s)*futMid, volMid, domRate)
		}
		fmt.Fprintf(&output, "%6.1f %6.0f %5.2f %4.1f\n", (1.0+s)*spotMid, pv[j], delta[j], vega[j])
	}

	return output.String()
}

func ContractMarketSummary(mds RPCMDSConnC, t time.Time) string {
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
	benchmarkContracts, _ := bean.ContractsFromNames(benchmarkNames, nil)
	var output strings.Builder

	_, _, btcspot := SpotPrice(mds, t, bean.Pair{bean.BTC, bean.USDT})
	fmt.Fprintf(&output, "%s   %6.1f\n", t.Format("Mon 02Jan06 15:04"), btcspot)
	fmt.Fprintf(&output, "Vol          Prem (BTC)\n")
	for c, _ := range benchmarkContracts {
		_, _, _, volBid, volAsk := ContractMarket(mds, t, c)
		fmt.Fprintf(&output, "%s\n%5.1f/%5.1f\n", c.Name(), volBid*100.0, volAsk*100.0)
	}
	return output.String()
}

func ContractHistory(mds RPCMDSConnC, cName string) (string, error) {
	var output strings.Builder
	c, err := bean.ContractFromName(cName)

	if err != nil {
		return "", err
	}
	n := time.Now()
	en := time.Date(n.Year(), n.Month(), n.Day(), 10, 0, 0, 0, time.UTC)
	if en.After(n) {
		en = en.Add(-24 * time.Hour)
	}
	st := en.Add(-30 * 24 * time.Hour)

	for t := st; t.Before(en); t = t.Add(24 * time.Hour) {
		spotMid, _, _, volBid, volAsk := ContractMarket(mds, t, c)
		if t.Equal(st) {
			fmt.Fprintf(&output, "%s   %6.1f\n", c.Name(), spotMid)
		}
		fmt.Fprintf(&output, "%s %5.1f/%5.1f\n", t.Format("02Jan06"), volBid*100.0, volAsk*100.0)
	}
	return output.String(), nil
}

func ContractMarket(mds RPCMDSConnC, asof time.Time, c bean.Contract) (spotMid, futMid, domRate, volBid, volAsk float64) {
	_, _, spotMid = SpotPrice(mds, asof, c.Underlying())
	_, _, futMid = ContractPrice(mds, asof, c.UnderFuture())
	optBid, optAsk, _ := ContractPrice(mds, asof, c)
	domRate = bean.USDiscountRate
	volBid = c.ImpVol(asof, spotMid, futMid, domRate, optBid)
	volAsk = c.ImpVol(asof, spotMid, futMid, domRate, optAsk)
	return
}

func SpotPrice(mds RPCMDSConnC, asof time.Time, p bean.Pair) (bid, ask, mid float64) {
	st := asof.Add(time.Duration(-1) * time.Minute)
	en := asof.Add(time.Duration(1) * time.Minute)

	spotobts, _ := mds.GetOrderBookTS(p, st, en, 20)

	bid, ask = priceAt(spotobts, asof)
	mid = (bid + ask) / 2.0
	return
}

func ContractPrice(mds RPCMDSConnC, asof time.Time, c bean.Contract) (bid, ask, mid float64) {
	var obts bean.OrderBookTS
	st := asof.Add(time.Duration(-1) * time.Minute)
	en := asof.Add(time.Duration(1) * time.Minute)

	if c.IsOption() {
		obts, _ = mds.GetOptOrderBookTS(c.Name(), st, en, 20)
	} else {
		obts, _ = mds.GetFutOrderBookTS(c.Name(), st, en, 20)
	}
	bid, ask = priceAt(obts, asof)
	mid = (bid + ask) / 2.0
	return
}

// given an orderbook time series, find the bidask at a specific time
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
