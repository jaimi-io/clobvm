package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	"github.com/jaimi-io/hypersdk/utils"
)

type AddOrder struct {
	Pair     orderbook.Pair `json:"pair"`
	Quantity uint64  	      `json:"quantity"`
	Price    uint64  		    `json:"price"`
	Side     bool 			    `json:"side"`
}

func (ao *AddOrder) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (ao *AddOrder) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (ao *AddOrder) amount() (uint64, ids.ID) {
	isFilled := false
	getAmount := orderbook.GetAmountFn(ao.Side, isFilled, ao.Pair)
	return getAmount(ao.Quantity, ao.Price)
}

func (ao *AddOrder) StateKeys(cauth chain.Auth, txID ids.ID) [][]byte {
	user := auth.GetUser(cauth)
	return [][]byte{
		storage.BalanceKey(user, ao.Pair.BaseTokenID),
		storage.BalanceKey(user, ao.Pair.QuoteTokenID),
	}
}

func (ao *AddOrder) Fee() (amount int64, tokenID ids.ID) {
	_, tokenID = ao.amount()
	return 1, tokenID
}

func (ao *AddOrder) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	timestamp int64,
	cauth chain.Auth,
	txID ids.ID,
	warpVerified bool,
	memoryState any,
	blockHeight uint64,
) (result *chain.Result, err error) {
	obm := memoryState.(*orderbook.OrderbookManager)
	user := auth.GetUser(cauth)
	if err = storage.PullPendingBalance(ctx, db, obm, user, ao.Pair.BaseTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, nil
	}
	if err = storage.PullPendingBalance(ctx, db, obm, user, ao.Pair.QuoteTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, nil
	}

	if ao.Quantity == 0 {
		err = errors.New("amount cannot be zero")
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, nil
	}
	amount, tokenID := ao.amount()
	if err = storage.DecBalance(ctx, db, user, tokenID, amount); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, nil
	}
	return &chain.Result{Success: true, Units: 0}, nil
}

func (ao *AddOrder) Marshal(p *codec.Packer) {
	p.PackID(ao.Pair.BaseTokenID)
	p.PackID(ao.Pair.QuoteTokenID)
	p.PackUint64(ao.Quantity)
	p.PackUint64(ao.Price)
	p.PackBool(ao.Side)
}

func UnmarshalAddOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var ao AddOrder
	p.UnpackID(true, &ao.Pair.BaseTokenID)
	p.UnpackID(true, &ao.Pair.QuoteTokenID)
	ao.Quantity = p.UnpackUint64(true)
	ao.Price = p.UnpackUint64(true)
	ao.Side = p.UnpackBool()
	return &ao, p.Err()
}