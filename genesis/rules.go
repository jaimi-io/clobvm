package genesis

import "github.com/ava-labs/avalanchego/ids"

type Rules struct {}

func (r *Rules) GetMaxBlockTxs() int {
	return 20_000
}

func (r *Rules) GetMaxBlockUnits() uint64 {
	return 1_800_000
}

func (r *Rules) GetValidityWindow() int64 {
	return 60
}

func (r *Rules) GetBaseUnits() uint64 {
	return 48
}

func (r *Rules) GetMinUnitPrice() uint64 {
	return 1
}

func (r *Rules) GetUnitPriceChangeDenominator() uint64 {
	return 48
}

func (r *Rules) GetWindowTargetUnits() uint64 {
	return 20_000_000
}

func (r *Rules) GetMinBlockCost() uint64 {
	return 0
}

func (r *Rules) GetBlockCostChangeDenominator() uint64 {
	return 48
}

func (r *Rules) GetWindowTargetBlocks() uint64 {
	return 20
}

func (r *Rules) GetWarpConfig(sourceChainID ids.ID) (bool, uint64, uint64) {
	return true, 4, 5
}

func (r *Rules) GetWarpBaseFee() uint64 {
	return 1_024
}

func (r *Rules) GetWarpFeePerSigner() uint64 {
	return 128
}

func (r *Rules) FetchCustom(string) (any, bool) {
	return nil, false
}