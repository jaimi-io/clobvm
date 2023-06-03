package genesis

import "github.com/ava-labs/avalanchego/ids"

type Rules struct {
	g *Genesis
}

func (g *Genesis) NewRules() *Rules {
	return &Rules{g}
}

func (r *Rules) GetMaxBlockTxs() int {
	return r.g.MaxBlockTxs
}

func (r *Rules) GetMaxBlockUnits() uint64 {
	return r.g.MaxBlockUnits
}

func (r *Rules) GetValidityWindow() int64 {
	return r.g.ValidityWindow
}

func (r *Rules) GetBaseUnits() uint64 {
	return r.g.BaseUnits
}

func (r *Rules) GetMinUnitPrice() uint64 {
	return r.g.MinUnitPrice
}

func (r *Rules) GetUnitPriceChangeDenominator() uint64 {
	return r.g.UnitPriceChangeDenominator
}

func (r *Rules) GetWindowTargetUnits() uint64 {
	return r.g.WindowTargetUnits
}

func (r *Rules) GetMinBlockCost() uint64 {
	return r.g.MinBlockCost
}

func (r *Rules) GetBlockCostChangeDenominator() uint64 {
	return r.g.BlockCostChangeDenominator
}

func (r *Rules) GetWindowTargetBlocks() uint64 {
	return r.g.WindowTargetBlocks
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