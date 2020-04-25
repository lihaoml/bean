package mds

import (
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
		"expiry": strings.ToUpper(p.Expiry.Format("2Jan06")),
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

func (mdss *MDSSink) OrderBook(p ConOBPoint) {
	go func() {
		mdss.lockbp.Lock()
		writeOBBatchPoints(mdss.bp, p.ExName, p.Instrument, p.Symbol, "BID", p.OB.Bids(), p.TimeStamp, p.Lag)
		writeOBBatchPoints(mdss.bp, p.ExName, p.Instrument, p.Symbol, "ASK", p.OB.Asks(), p.TimeStamp, p.Lag)
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
