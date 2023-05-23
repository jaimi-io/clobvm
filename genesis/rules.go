package genesis

import "github.com/ava-labs/avalanchego/ids"

type Rules struct {}

func (r *Rules) GetMaxBlockTxs() int {
	return 0
}

func (r *Rules) GetMaxBlockUnits() uint64 {
	return 0
}

func (r *Rules) GetValidityWindow() int64 {
	return 0
}

func (r *Rules) GetBaseUnits() uint64 {
	return 0
}

func (r *Rules) GetMinUnitPrice() uint64 {
	return 0
}

func (r *Rules) GetUnitPriceChangeDenominator() uint64 {
	return 0
}

func (r *Rules) GetWindowTargetUnits() uint64 {
	return 0
}

func (r *Rules) GetMinBlockCost() uint64 {
	return 0
}

func (r *Rules) GetBlockCostChangeDenominator() uint64 {
	return 0
}

func (r *Rules) GetWindowTargetBlocks() uint64 {
	return 0
}

func (r *Rules) GetWarpConfig(sourceChainID ids.ID) (bool, uint64, uint64) {
	return false, 0, 0
}

func (r *Rules) GetWarpBaseFee() uint64 {
	return 0
}

func (r *Rules) GetWarpFeePerSigner() uint64 {
	return 0
}

func (r *Rules) FetchCustom(string) (any, bool) {
	return nil, false
}