package test

import (
	"reflect"
	"testing"

	"bean"
	"github.com/stretchr/testify/assert"
)

func TestPorfolio(t *testing.T) {
	p := bean.NewPortfolio()
	p.AddBalance(bean.BTC, 1)
	p.AddBalance(bean.ETH, 2)
	p.AddBalance(bean.IOTX, 10)
	p.AddBalance(bean.USDT, 20)

	q := p.Clone()
	assert.True(t, reflect.DeepEqual(p, q), "Clone should be the same")
	/*
		q.AddBalance(bean.ETH, 3)
		assert.True(t, reflect.DeepEqual(p.Add(q), bean.Portfolio{map[Coin]float64{BTC: 2, ETH: 7, IOTX: 20, USDT: 40}}))
		q.RemoveBalance(bean.ETH, 3)
		assert.True(t, reflect.DeepEqual(p, q), "Should be the same after removing balance")

		assert.True(t, reflect.DeepEqual(p.Add(q), bean.portfolio{map[Coin]float64{BTC: 2, ETH: 4, IOTX: 20, USDT: 40}}))
		assert.True(t, reflect.DeepEqual(p.Subtract(q), portfolio{map[Coin]float64{BTC: 0, ETH: 0, IOTX: 0, USDT: 0}}))
		assert.True(t, reflect.DeepEqual(p.Filter(Coins{BTC, ETH}), portfolio{map[Coin]float64{BTC: 1, ETH: 2}}))
	*/
}
