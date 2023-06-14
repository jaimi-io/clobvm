package orderbook

import (
	"math"

	"github.com/jaimi-io/clobvm/utils"
)

type Fee struct {
	Amount uint64
	MakerRate float64
	TakerRate float64
}

var fees = []Fee{
	{Amount: 0 * utils.MinBalance(), MakerRate: 0.001, TakerRate: 0.0015},
	{Amount: 100_000 * utils.MinBalance(), MakerRate: 0.0009, TakerRate: 0.001},
	{Amount: 1_000_000 * utils.MinBalance(), MakerRate: 0.0008, TakerRate: 0.001},
	{Amount: 10_000_000 * utils.MinBalance(), MakerRate: 0.0007, TakerRate: 0.0009},
}

func getFeeRates(monthlyExecuted uint64) (float64, float64) {
	var currentMakerRate, currentTakerRate float64
	for _, fee := range fees {
		if monthlyExecuted >= fee.Amount {
			currentMakerRate = fee.MakerRate
			currentTakerRate = fee.TakerRate
		} else {
			break
		}
	}
	return currentMakerRate, currentTakerRate
}

func CalculateTakerFee(monthlyExecuted uint64, amount uint64) uint64 {
	_, takerRate := getFeeRates(monthlyExecuted)
	takerFee := float64(amount) * takerRate
	return uint64(math.Ceil(takerFee))
}

func GetMakerRate(monthlyExecuted uint64) float64 {
	makerRate, _ := getFeeRates(monthlyExecuted)
	return makerRate
}

func RefundTakerFee(monthlyExecuted uint64, amount uint64) uint64 {
	makerRate, takerRate := getFeeRates(monthlyExecuted)
	refundRate := takerRate - makerRate
	refund := float64(amount) * refundRate
	return uint64(math.Ceil(refund))
}