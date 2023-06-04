package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	"github.com/jaimi-io/hypersdk/utils"
)

type CancelOrder struct {
	Pair    orderbook.Pair `json:"pair"`
	OrderID ids.ID  	     `json:"orderID"`
	Side    bool 			     `json:"side"`
}

func (co *CancelOrder) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (co *CancelOrder) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (co *CancelOrder) StateKeys(auth chain.Auth, _ ids.ID) [][]byte {
	user := auth.PublicKey()
	return [][]byte{
		storage.BalanceKey(user, co.Pair.BaseTokenID),
		storage.BalanceKey(user, co.Pair.QuoteTokenID),
	}
}

func (co *CancelOrder) Fee() (amount int64, tokenID ids.ID) {
	return 1, co.Pair.TokenID(co.Side)
}

func (co *CancelOrder) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	timestamp int64,
	auth chain.Auth,
	txID ids.ID,
	warpVerified bool,
	memoryState any,
	blockHeight uint64,
) (result *chain.Result, err error) {
	obm := memoryState.(*orderbook.OrderbookManager)
	user := auth.PublicKey()
	if err = storage.PullPendingBalance(ctx, db, obm, user, co.Pair.BaseTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, nil
	}
	if err = storage.PullPendingBalance(ctx, db, obm, user, co.Pair.QuoteTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, nil
	}
	return &chain.Result{Success: true, Units: 0}, nil
}

func (co *CancelOrder) Marshal(p *codec.Packer) {
	p.PackID(co.Pair.BaseTokenID)
	p.PackID(co.Pair.QuoteTokenID)
	p.PackID(co.OrderID)
	p.PackBool(co.Side)
}

func UnmarshalCancelOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var co CancelOrder
	p.UnpackID(true, &co.Pair.BaseTokenID)
	p.UnpackID(true, &co.Pair.QuoteTokenID)
	p.UnpackID(true, &co.OrderID)
	co.Side = p.UnpackBool()
	return &co, p.Err()
}