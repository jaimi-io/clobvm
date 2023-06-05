package utils

import (
	"math"

	"github.com/jaimi-io/clobvm/consts"
)

func MinQuantity() uint64 {
	return uint64(math.Pow10(consts.BalanceDecimals - consts.QuantityDecimals))
}

func QuantityToBalance(quantity uint64) uint64 {
	return quantity * MinQuantity()
}

func BalanceToQuantity(balance uint64) uint64 {
	return balance / MinQuantity()
}

func toDecimal(value uint64, decimals int) float64 {
	return float64(value) / math.Pow(10, float64(decimals))
}

func DisplayPrice(price uint64) float64 {
	return toDecimal(price, consts.PriceDecimals)
}

func DisplayQuantity(quantity uint64) float64 {
	return toDecimal(quantity, consts.QuantityDecimals)
}

func DisplayBalance(balance uint64) float64 {
	return toDecimal(balance, consts.BalanceDecimals)
}