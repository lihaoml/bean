package bean

// note that portfolio algebra does not carry locked portfolio, only clone() does
type Portfolio interface {
	// Log(string)
	Clone() Portfolio
	Add(Portfolio) Portfolio
	Minus(Portfolio) Portfolio
	Subtract(Portfolio) Portfolio
	Filter(Coins) Portfolio
	Balances() map[Coin]float64
	Balance(Coin) float64
	AddBalance(Coin, float64)
	RemoveBalance(Coin, float64)
	SetBalance(Coin, float64)
	AvailableBalance(c Coin) float64
	SetLockedBalance(Coin, float64)
	Coins() Coins
	AddContract(Contract)
	Contracts() []Contract
}

// Portfolio a Portfolio for an account
type portfolio struct {
	balances       map[Coin]float64 // total balance of each coin
	lockedBalances map[Coin]float64 // locked blance by exchange, used to calculate the free blance for placing order
	contracts      []Contract
}

func NewPortfolio(bal ...interface{}) Portfolio {
	if len(bal) == 0 {
		return portfolio{
			balances:       make(map[Coin]float64),
			lockedBalances: make(map[Coin]float64),
			contracts:      make([]Contract, 0),
		}
	} else if len(bal) == 1 {
		return portfolio{
			balances:       bal[0].(map[Coin]float64),
			lockedBalances: make(map[Coin]float64),
			contracts:      make([]Contract, 0),
		}
	} else if len(bal) == 2 {
		return portfolio{
			balances:       bal[0].(map[Coin]float64),
			lockedBalances: bal[1].(map[Coin]float64),
			contracts:      make([]Contract, 0),
		}
	} else {
		panic("invalid number of input for NewPortfolio")
	}
}

func (p portfolio) Balances() map[Coin]float64 {
	return p.balances
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

// Add - add two portfolios and return a new one
func (p portfolio) Add(p2 Portfolio) Portfolio {
	port := p.Clone()
	for c, v := range p2.Balances() {
		port.AddBalance(c, v)
	}
	return port
}

// Add - add two portfolios and return a new one
func (p portfolio) Minus(p2 Portfolio) Portfolio {
	port := p.Clone()
	for c, v := range p2.Balances() {
		port.AddBalance(c, -v)
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

func (p portfolio) AddContract(c Contract) {
	p.contracts = append(p.contracts, c)
}

func (p portfolio) Contracts() []Contract {
	return p.contracts
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
