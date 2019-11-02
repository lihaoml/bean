package bean

type Ticker struct {
	BestBid       float64
	BestAsk       float64
	BestBidAmount float64
	BestAskAmount float64
	LastPrice     float64
	LastAmount    float64
	Change24H     float64
	Volume24H     float64 // volume in base
}
