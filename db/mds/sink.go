package mds

import (
	"bean"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type MDSSink struct {
	mds    *MDS
	bp     client.BatchPoints
	lockbp sync.Mutex
	stop   chan struct{}
}

func NewMDSSink(mds *MDS) (mdss *MDSSink) {
	mdss = &MDSSink{mds: mds, stop: make(chan struct{})}
	mdss.Empty()

	timer := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-mdss.stop:
				return
			case <-timer.C:
				mdss.Empty()
			}
		}
	}()
	return
}

func (mdss *MDSSink) Stop() {
	mdss.stop <- struct{}{}
}

func (mdss *MDSSink) Empty() (err error) {
	mdss.lockbp.Lock()
	bp2 := mdss.bp
	mdss.bp, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "us",
	})
	mdss.lockbp.Unlock()
	if err != nil {
		return
	}
	if bp2 != nil {
		err = mdss.mds.WriteBatchPoints(bp2)
	}
	return
}

func (mdss *MDSSink) SmilePoint(p SmilePoint) (err error) {
	tags := map[string]string{
		"expiry": strings.ToUpper(p.Expiry.Format(bean.ContractDateFormat)),
		"pair":   p.Pair.String()}
	fields := map[string]interface{}{
		"Atm":   p.Atm * 100.0,
		"RR25":  p.RR25 * 100.0,
		"RR10":  p.RR10 * 100.0,
		"Fly25": p.Fly25 * 100.0,
		"Fly10": p.Fly10 * 100.0}
	pt, err := client.NewPoint("SMILE", tags, fields, p.TimeStamp)
	if err != nil {
		return
	}
	go func() {
		mdss.lockbp.Lock()
		mdss.bp.AddPoint(pt)
		mdss.lockbp.Unlock()
	}()
	return
}

func (mdss *MDSSink) OrderBook(timeStamp time.Time, exName string, instr string, side bean.Side, bestOrder bean.Order, lag time.Duration) {
	go func() {
		mdss.lockbp.Lock()
		tags := map[string]string{
			"index":      "0",
			"instrument": instr,
			"exchange":   exName,
			"side":       string(side),
			"symbol":     "",
		}
		fields := map[string]interface{}{
			"Price":  bestOrder.Price,
			"Amount": bestOrder.Amount,
			"Lag":    lag.Seconds()}
		pt1, _ := client.NewPoint(MT_CONTRACT_ORDERBOOK, tags, fields, timeStamp)
		mdss.bp.AddPoint(pt1)

		mdss.lockbp.Unlock()
	}()
}

// TEMP - THIS SHOULD MOVE TO TDS
func (mdss *MDSSink) Risk(timeStamp time.Time, exName string, con *bean.Contract, pair bean.Pair, pos, spot, vol, pv, delta, gamma, vega, theta float64) {
	go func() {
		mdss.lockbp.Lock()
		var tags map[string]string
		if con == nil {
			tags = map[string]string{"instrument": "CASH",
				"exchange": exName,
				//				"pair":       pair.String(),
				//				"expiryDays": "0",
			}
		} else if con.IsOption() {
			tags = map[string]string{
				"instrument": con.Name(),
				"exchange":   exName,
				//				"pair":       pair.String(),
				//				"expiryDays": strconv.Itoa(con.ExpiryDays(timeStamp)),
				//				"strike":     strconv.Itoa(int(con.Strike())),
				//				"callPut":    string(con.CallPut()),
			}
		} else {
			tags = map[string]string{
				"instrument": con.Name(),
				"exchange":   exName,
				//				"pair":       pair.String(),
				//				"expiryDays": strconv.Itoa(con.ExpiryDays(timeStamp)),
				//				"strike":     "",
				//				"callPut":    "",
			}
		}
		fields := map[string]interface{}{
			"Position": pos,
			"Spot":     spot,
			"Vol":      vol,
			"Pv":       pv,
			"Delta":    delta,
			"Gamma":    gamma,
			"Vega":     vega,
			"Theta":    theta,
		}
		pt, _ := client.NewPoint(MT_CONTRACT_RISK, tags, fields, timeStamp)
		mdss.bp.AddPoint(pt)

		mdss.lockbp.Unlock()
	}()
}

func (mdss *MDSSink) ArbPoint(p ArbPoint) {
	tags := map[string]string{
		"instr": p.Instrument}
	fields := map[string]interface{}{
		"lag":     p.Lag.Seconds(),
		"arb":     p.Arb,
		"arbsize": p.ArbSize}
	pt, err := client.NewPoint("SMILE", tags, fields, p.TimeStamp)
	if err != nil {
		return
	}
	go func() {
		mdss.lockbp.Lock()
		mdss.bp.AddPoint(pt)
		mdss.lockbp.Unlock()
	}()
	return
}

func (mdss *MDSSink) TxnPoint(p ConTxnPoint) {
	go func() {
		mdss.lockbp.Lock()
		writeConTxnBatchPoints(mdss.bp, p)
		mdss.lockbp.Unlock()
	}()
}
