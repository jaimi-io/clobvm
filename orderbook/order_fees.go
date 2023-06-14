package orderbook

import "math"

type Fee struct {
	Amount uint64
	Percent float32
}

var fees = []Fee{
	{Amount: 0, Percent: 0.004},
	{Amount: 10_000_000, Percent: 0.003},
	{Amount: 1_000_000_000, Percent: 0.002},
	{Amount: math.MaxUint64, Percent: 0.001},
}

func getFee(monthlyExecuted uint64) float64 {
	var currentFee float64
	for _, fee := range fees {
		if monthlyExecuted >= fee.Amount {
			currentFee = float64(fee.Percent)
		} else {
			break
		}
	}
	return currentFee
}

func CalculateFee(monthlyExecuted uint64, amount uint64) uint64 {
	fee := getFee(monthlyExecuted)
	cumulFee := float64(amount) * fee
	return uint64(math.Ceil(cumulFee))
}

func GetFeeRate(monthlyExecuted uint64) float64 {
	return getFee(monthlyExecuted)
}