package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	hutils "github.com/jaimi-io/hypersdk/utils"
)

type CancelOrder struct {
	Pair    orderbook.Pair `json:"pair"`
	OrderID ids.ID  	     `json:"orderID"`
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

func (co *CancelOrder) Fee(timestamp int64, auth chain.Auth, memoryState any) (amount uint64) {
	return 1
}

func (co *CancelOrder) Token(memoryState any) (tokenID ids.ID) {
	return co.Pair.QuoteTokenID
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
	var baseBalance uint64
	var quoteBalance uint64
	if baseBalance, err = storage.PullPendingBalance(ctx, db, obm, user, co.Pair.BaseTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if quoteBalance, err = storage.PullPendingBalance(ctx, db, obm, user, co.Pair.QuoteTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	output := utils.PackUpdatedBalance(user, baseBalance, user, quoteBalance)
	return &chain.Result{Success: true, Units: 0, Output: output}, nil
}

func (co *CancelOrder) Marshal(p *codec.Packer) {
	p.PackID(co.Pair.BaseTokenID)
	p.PackID(co.Pair.QuoteTokenID)
	p.PackID(co.OrderID)
}

func UnmarshalCancelOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var co CancelOrder
	p.UnpackID(true, &co.Pair.BaseTokenID)
	p.UnpackID(true, &co.Pair.QuoteTokenID)
	p.UnpackID(false, &co.OrderID)
	return &co, p.Err()
}