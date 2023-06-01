package actions

import (
	"context"
	"fmt"

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
	TokenID ids.ID  	`json:"tokenID"`
	OrderID ids.ID  	`json:"orderID"`
}

func (co *CancelOrder) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (co *CancelOrder) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (co *CancelOrder) StateKeys(cauth chain.Auth, _ ids.ID) [][]byte {
	user := auth.GetUser(cauth)
	return [][]byte{
		storage.BalanceKey(user, co.TokenID),
		storage.OrderKey(co.OrderID),
	}
}

func (co *CancelOrder) Fee() (amount int64, tokenID ids.ID) {
	return 1, co.TokenID
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
	ob := memoryState.(*orderbook.Orderbook)
	user := auth.GetUser(cauth)
	fmt.Println("CancelOrder.Execute: user=", user, "co=", co.OrderID.String())
	remaining, err := storage.RemoveOrder(ctx, db, co.OrderID)
	if err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}
	if err = storage.IncBalance(ctx, db, user, co.TokenID, remaining); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}
	order := ob.Get(co.OrderID)
	ob.Remove(order)
	return &chain.Result{Success: true, Units: 0}, nil
}

func (co *CancelOrder) Marshal(p *codec.Packer) {
	p.PackID(co.TokenID)
	p.PackID(co.OrderID)
}

func UnmarshalCancelOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var co CancelOrder
	p.UnpackID(true, &co.TokenID)
	p.UnpackID(true, &co.OrderID)
	return &co, p.Err()
}