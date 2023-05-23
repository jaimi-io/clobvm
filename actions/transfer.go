package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
)

type Transfer struct {
}

func (t *Transfer) MaxUnits(r chain.Rules) uint64 {
	return 0
}

func (t *Transfer) ValidRange(r chain.Rules) (start int64, end int64) {
	return 0, 0
}

func (t *Transfer) StateKeys() [][]byte {
	return nil
}

func (t *Transfer) Fee() (amount int64, tokenID ids.ID) {
	return 0, ids.ID{}
}

func (t *Transfer) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	timestamp int64,
	auth chain.Auth,
	txID ids.ID,
	warpVerified bool,
) (result *chain.Result, err error) {
	return nil, nil
}

func (t *Transfer) Marshal(p *codec.Packer) {

}