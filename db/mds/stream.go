package mds

import (
	. "bean"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

// stream contains functions that allow the streaming of data to be written to mds through channels

const bufferSize = 10000

// ConOBPoint allows sending of the orderbook for writing to the MDS ORDERBOOK table
type ConOBPoint struct {
	TimeStamp  time.Time
	ExName     string
	Instrument string // instrument name, e.g., BTC-PERPETUAL, BTC-27DEC19
	OB         OrderBook
	Lag        time.Duration
	Symbol     string // contract symbol in exchange, not necessarily for all exchanges, e.g., XBTUSD, XBTZ19
}

// ConTxnPoint allows sending of a reported trade for writing to the MDS TRANSACTIONS table
type ConTxnPoint struct {
	TimeStamp  time.Time
	ExName     string
	Instrument string // instrument name, e.g., BTC-PERPETUAL, BTC-27DEC19
	Side       Side
	Price      float64
	Amount     float64
	IndexPrice float64
	Vol        float64
	TxnID      string
	Symbol     string // contract symbol in exchange, not necessarily for all exchanges, e.g., XBTUSD, XBTZ19
}

type SpotTxnPoint struct {
	ExName string
	Txn    Transaction
}

type SpotOBPoint struct {
	TimeStamp time.Time
	ExName    string
	Pair      Pair
	OB        OrderBook
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
	TimeStamp                     time.Time
	Pair                          Pair
	Expiry                        time.Time
	Atm, RR25, RR10, Fly25, Fly10 float64
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
		Precision: "us",
	})
	writeDbTicker := time.NewTicker(2 * time.Second)

	go func() {
		for {
			if bp == nil {
				bp, err = client.NewBatchPoints(client.BatchPointsConfig{
					Database:  MDS_DBNAME,
					Precision: "us",
				})
			}
			select {
			case <-writeDbTicker.C:
				// allow failing writes as long as there are succeeding ones
				errMsgs := []string{}
				for _, c := range mds.cs {
					err = c.Write(bp)
					if err != nil {
						errMsgs = append(errMsgs, "err writing db points "+err.Error())
					}
				}
				if len(errMsgs) == len(mds.cs) {
					// all writing failed, stop the process
					errCh <- errors.New(fmt.Sprint(errMsgs))
					stopCh <- true
					return
				}
				bp = nil
			case <-stopCh:
				for _, c := range mds.cs {
					c.Write(bp)
					c.Close()
				}
				return
			case dataPt := <-dataPtCh:
				switch p := dataPt.(type) {
				case ConOBPoint:
					writeOBBatchPoints(bp, p.ExName, p.Instrument, p.Symbol, "BID", p.OB.Bids(), p.TimeStamp, p.Lag)
					writeOBBatchPoints(bp, p.ExName, p.Instrument, p.Symbol, "ASK", p.OB.Asks(), p.TimeStamp, p.Lag)

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
					writeConTxnBatchPoints(bp, p)
				case SpotTxnPoint:
					writeSpotTxnBatchPoints(bp, p.ExName, p.Txn)
				case SpotOBPoint:
					writeSpotOBBatchPoints(bp, p.ExName, p.Pair, p.OB, p.TimeStamp)

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
						"Atm":   p.Atm * 100.0,
						"RR25":  p.RR25 * 100.0,
						"RR10":  p.RR10 * 100.0,
						"Fly25": p.Fly25 * 100.0,
						"Fly10": p.Fly10 * 100.0}
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
