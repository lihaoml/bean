package mds

import (
	. "bean"
	"beanex/risk"
	"github.com/influxdata/influxdb/client/v2"
	"strings"
	"time"
)

// stream contains functions that allow the streaming of data to be written to mds through channels

const bufferSize = 5000

// ConOBPoint allows sending of the orderbook for writing to the MDS ORDERBOOK table
type ConOBPoint struct {
	TimeStamp  time.Time
	Instrument string
	OB         OrderBook
	Lag        time.Duration
}

// ConTxnPoint allows sending of a reported trade for writing to the MDS TRANSACTIONS table
type ConTxnPoint struct {
	TimeStamp  time.Time
	Instrument string
	Side       Side
	Price      float64
	Amount     float64
	IndexPrice float64
	Vol        float64
}

type MessagePoint struct {
	TimeStamp  time.Time
	Instrument string
	Type       string
	Message    string
}

// ArbPoint allows a detected arbitrage to be written to the MDS ARB table
type ArbPoint struct {
	TimeStamp  time.Time
	Instrument string
	Arb        float64
	ArbSize    float64
	Lag        time.Duration
}

// SmilePoint allows fitted smile parameters to be written to the MDS SMILE table
type SmilePoint struct {
	TimeStamp time.Time
	Pair      Pair
	Expiry    time.Time
	VolSmile  *risk.FivePointSmile
}

// Writer creates a DataPtCh channel into which various structures can be sent.
// These are decoded and written to the appropriate MDS tables in a batch format
// stopCh is bi-directional and indicates the dataptch is closed
// errCh allows reporting of errors from the writer
func (mds MDS) Writer() (dataPtCh chan interface{}, stopCh chan bool, errCh chan error) {
	var err error
	dataPtCh = make(chan interface{}, bufferSize)
	stopCh = make(chan bool)
	errCh = make(chan error)

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MDS_DBNAME,
		Precision: "ms",
	})
	writeDbTicker := time.NewTicker(2 * time.Second)

	go func() {
		for {
			if bp == nil {
				bp, err = client.NewBatchPoints(client.BatchPointsConfig{
					Database:  MDS_DBNAME,
					Precision: "ms",
				})
			}
			select {
			case <-writeDbTicker.C:
				err = mds.c.Write(bp)
				if err != nil {
					errCh <- err
					stopCh <- true
					return
				}
				bp = nil
			case <-stopCh:
				mds.c.Write(bp)
				mds.c.Close()
				return
			case dataPt := <-dataPtCh:
				switch p := dataPt.(type) {
				case ConOBPoint:
					writeOBBatchPoints(bp, NameDeribit, p.Instrument, "BID", p.OB.Bids(), p.TimeStamp, p.Lag)
					writeOBBatchPoints(bp, NameDeribit, p.Instrument, "ASK", p.OB.Asks(), p.TimeStamp, p.Lag)

				case MessagePoint:
					tags := map[string]string{
						"instrument": p.Instrument,
						"type":       p.Type}
					fields := map[string]interface{}{
						"Message": p.Message}
					pt, err := client.NewPoint("MSG", tags, fields, p.TimeStamp)
					if err != nil {
						errCh <- err
						stopCh <- true
						return
					}
					bp.AddPoint(pt)

				case ConTxnPoint:
					writeTxnBatchPoints(bp, NameDeribit, p.Instrument, p.Side, p.Price, p.Amount, p.IndexPrice, p.Vol, p.TimeStamp)

				case ArbPoint:
					tags := make(map[string]string)
					fields := make(map[string]interface{})

					tags["instr"] = p.Instrument
					fields["lag"] = p.Lag.Seconds()
					fields["arb"] = p.Arb
					fields["arbsize"] = p.ArbSize
					pt, err := client.NewPoint("ARB", tags, fields, p.TimeStamp)
					if err != nil {
						errCh <- err
						stopCh <- true
						return
					}
					bp.AddPoint(pt)

				case SmilePoint:
					tags := map[string]string{
						"expiry": strings.ToUpper(p.Expiry.Format("2Jan06")),
						"pair":   p.Pair.String()}
					fields := map[string]interface{}{
						"Atm":   p.VolSmile.Atm * 100.0,
						"RR25":  p.VolSmile.RR25 * 100.0,
						"RR10":  p.VolSmile.RR10 * 100.0,
						"Fly25": p.VolSmile.Fly25 * 100.0,
						"Fly10": p.VolSmile.Fly10 * 100.0}
					pt, err := client.NewPoint("SMILE", tags, fields, p.TimeStamp)
					if err != nil {
						errCh <- err
						stopCh <- true
						return
					}
					bp.AddPoint(pt)
				}
			}
		}
	}()
	return
}
