package bean

import (
	"bean/utils"
	"fmt"
	"sort"
)

// note that portfolio algebra does not carry locked portfolio, only clone() does
type Portfolio interface {
	// Log(string)
	Clone() Portfolio
	Add(Portfolio) Portfolio
	Subtract(Portfolio) Portfolio
	Age(TradeLogS) Portfolio
	Filter(Coins) Portfolio
	Balances() map[Coin]float64
	LockedBalances() map[Coin]float64
	Balance(Coin) float64
	AddBalance(Coin, float64)
	RemoveBalance(Coin, float64)
	SetBalance(Coin, float64)
	AvailableBalance(c Coin) float64
	SetLockedBalance(Coin, float64)
	Coins() Coins
	AddContract(Contract, float64)
	SetContracts(Contracts)
	Contracts() Contracts
	ShowBrief()
}

// Portfolio a Portfolio for an account
type portfolio struct {
	balances       map[Coin]float64 // total balance of each coin
	lockedBalances map[Coin]float64 // locked blance by exchange, used to calculate the free blance for placing order
	contracts      map[Contract]float64
}

func NewPortfolio(bal ...interface{}) Portfolio {
	if len(bal) == 0 {
		return portfolio{
			balances:       make(map[Coin]float64),
			lockedBalances: make(map[Coin]float64),
			contracts:      make(map[Contract]float64),
			//			contracts:      make([]Contract, 0),
		}
	} else if len(bal) == 1 {
		return portfolio{
			balances:       bal[0].(map[Coin]float64),
			lockedBalances: make(map[Coin]float64),
			contracts:      make(map[Contract]float64),
		}
	} else if len(bal) == 2 {
		return portfolio{
			balances:       bal[0].(map[Coin]float64),
			lockedBalances: bal[1].(map[Coin]float64),
			contracts:      make(map[Contract]float64),
		}
	} else {
		panic("invalid number of input for NewPortfolio")
	}
}

func (p portfolio) Balances() map[Coin]float64 {
	return p.balances
}

func (p portfolio) LockedBalances() map[Coin]float64 {
	return p.lockedBalances
}

func (p portfolio) Balance(c Coin) float64 {
	return p.balances[c]
}

func (p portfolio) AddBalance(c Coin, v float64) {
	if _, coinExist := p.balances[c]; coinExist {
		p.balances[c] += v
	} else {
		p.balances[c] = v
	}
}

func (p portfolio) RemoveBalance(c Coin, v float64) {
	p.AddBalance(c, -v)
}

func (p portfolio) SetBalance(c Coin, v float64) {
	p.balances[c] = v
}

func (p portfolio) SetLockedBalance(c Coin, v float64) {
	p.lockedBalances[c] = v
}

func (p portfolio) AvailableBalance(c Coin) float64 {
	v, ok := p.lockedBalances[c]
	if ok {
		return p.Balance(c) - v
	} else {
		return p.Balance(c)
	}
}

func (p portfolio) Coins() (cs Coins) {
	for c, _ := range p.balances {
		cs = append(cs, c)
	}
	return cs
}

func (p portfolio) Clone() Portfolio {
	r := NewPortfolio()
	for k, v := range p.balances {
		r.SetBalance(k, v)
	}
	for k, v := range p.lockedBalances {
		r.SetLockedBalance(k, v)
	}
	return r
}

// Log - log to logger, note that we do not log locked balance since it's only for exchange
func (p portfolio) ShowBrief() {
	cs := p.Coins()
	sort.Sort(cs)
	for _, c := range cs {
		v := p.balances[c]
		if v != 0 {
			fmt.Println(string(c), util.RenderFloat("#,###.####", v), "[LOCKED]", util.RenderFloat("#,###.####", p.lockedBalances[c]))
		}
	}
}

// Add - add two portfolios and return a new one
func (p portfolio) Add(p2 Portfolio) Portfolio {
	port := p.Clone()
	for c, v := range p2.Balances() {
		port.AddBalance(c, v)
	}
	return port
}

// Subtract - subtract one Portfolio from another and return a new one
func (p portfolio) Subtract(p2 Portfolio) Portfolio {
	port := p.Clone()
	for c, v := range p2.Balances() {
		port.AddBalance(c, -v)
	}
	return port
}

func (p_ portfolio) Age(s TradeLogS) Portfolio {
	p := p_.Clone()
	// put it in current snapshot
	for _, t := range s {
		if t.Side == BUY {
			p.AddBalance(t.Pair.Coin, t.Quantity)
			p.RemoveBalance(t.Pair.Base, t.Quantity*t.Price)
		} else {
			p.RemoveBalance(t.Pair.Coin, t.Quantity)
			p.AddBalance(t.Pair.Base, t.Quantity*t.Price)
		}
		// remove commission
		// FIXME: hack for FCoin maker fee, should be put into FCOIN
		if t.Commission == 0 {
			if t.Side == BUY {
				p.AddBalance(t.Pair.Coin, t.Quantity*0.0005)
			} else {
				p.AddBalance(t.Pair.Base, t.Quantity*t.Price*0.0005)
			}
		}
		p.RemoveBalance(t.CommissionAsset, t.Commission)
	}
	return p
}

func (p portfolio) Filter(coins Coins) Portfolio {
	r := NewPortfolio()
	for _, c := range coins {
		r.SetBalance(c, p.Balance(c))
		lockedC, locked := p.lockedBalances[c]
		if locked {
			r.SetLockedBalance(c, lockedC)
		}
	}
	return r
}

func (p portfolio) AddContract(c Contract, qty float64) {
	p.contracts[c] += qty
	//	p.contracts = append(p.contracts, c)
}

func (p portfolio) Contracts() Contracts {
	return p.contracts
}

func (p portfolio) SetContracts(cs Contracts) {
	for c, q := range cs {
		p.AddContract(c, q)
	}
}

/*
// Log - log to logger, note that we do not log locked balance since it's only for exchange
func (p portfolio) Log(msg string) {
	event := logger.Info()
	for k, v := range p.Balances() {
		if v != 0 {
			event.Str(string(k), util.RenderFloat("#,###.####", v))
		}
	}
	for k, v := range p.lockedBalances {
		if v != 0 {
			event.Str("[LOCKED]"+string(k), util.RenderFloat("#,###.####", v))
		}
	}
	event.Msg(msg)
}

func WalletBalance() Portfolio {
	err := godotenv.Load(".env.wallet")
	if err != nil {
		logger.Fatal().Msg("Error loading .env.wallet file")
	}

	resp, err := http.Get("https://etherscan.io/tokenholdings?a=0x1c5D0222ED85fe8c91b1F3c935A3Ed948681B346")

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var objmap map[string]*json.RawMessage
	_ = json.Unmarshal(body, &objmap)

	fmt.Print(objmap)
	return NewPortfolio()
}
*/
