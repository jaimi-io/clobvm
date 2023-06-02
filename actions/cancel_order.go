package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/auth"
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

func (co *CancelOrder) TokenID() ids.ID {
	return co.Pair.TokenID(co.Side, true)
}

func (co *CancelOrder) StateKeys(cauth chain.Auth, _ ids.ID) [][]byte {
	user := auth.GetUser(cauth)
	return [][]byte{
		storage.BalanceKey(user, co.Pair.BaseTokenID),
		storage.BalanceKey(user, co.Pair.QuoteTokenID),
	}
}

func (co *CancelOrder) Fee() (amount int64, tokenID ids.ID) {
	return 1, co.TokenID()
}

func (co *CancelOrder) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	timestamp int64,
	cauth chain.Auth,
	txID ids.ID,
	warpVerified bool,
	memoryState any,
) (result *chain.Result, err error) {
	obm := memoryState.(*orderbook.OrderbookManager)
	ob := obm.GetOrderbook(co.Pair)
	user := auth.GetUser(cauth)
	if err = storage.RetrieveFilledBalance(ctx, db, ob, user, co.Pair); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
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