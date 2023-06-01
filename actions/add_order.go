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

func (ao *AddOrder) TokenID() ids.ID {
	return ao.Pair.TokenID(ao.Side, true)
}

func (ao *AddOrder) Amount() uint64 {
	getAmount := orderbook.GetAmountFn(ao.Side, true)
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
	return 1, ao.TokenID()
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
) (result *chain.Result, err error) {
	obm := memoryState.(*orderbook.OrderbookManager)
	ob := obm.GetOrderbook(ao.Pair)
	user := auth.GetUser(cauth)
	if err = storage.RetrieveFilledBalance(ctx, db, ob, user, ao.Pair); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}

	if ao.Quantity == 0 {
		err = errors.New("amount cannot be zero")
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}
	if err = storage.DecBalance(ctx, db, user, ao.TokenID(), ao.Amount()); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}

	order := orderbook.NewOrder(txID, user, ao.Price, ao.Quantity, ao.Side)
	ob.Add(order)
	if order.Quantity < ao.Quantity {
		getAmount := orderbook.GetAmountFn(order.Side, false)
		filled := ao.Quantity - order.Quantity
		amount := getAmount(filled, order.Price)
		oppTokenID := ao.Pair.TokenID(ao.Side, false)
		if err = storage.IncBalance(ctx, db, user, oppTokenID, amount); err != nil {
			return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
		}
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