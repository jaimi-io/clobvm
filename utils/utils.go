package utils

import (
	"math"

	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/hypersdk/codec"
	"github.com/jaimi-io/hypersdk/crypto"
)

func MinBalance() uint64 {
	return uint64(math.Pow10(consts.BalanceDecimals))
}

func MinQuantity() uint64 {
	return uint64(math.Pow10(consts.BalanceDecimals - consts.QuantityDecimals))
}

func MinPrice() uint64 {
	return uint64(math.Pow10(consts.PriceDecimals))
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

func PackUpdatedBalance(baseUser crypto.PublicKey, baseBal uint64, quoteUser crypto.PublicKey, quoteBal uint64) []byte {
	p := codec.NewWriter(math.MaxInt)
	p.PackPublicKey(baseUser)
	p.PackUint64(baseBal)
	p.PackPublicKey(quoteUser)
	p.PackUint64(quoteBal)
	return p.Bytes()
}