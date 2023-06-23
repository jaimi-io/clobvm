package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	hutils "github.com/jaimi-io/hypersdk/utils"
)

type AddOrder struct {
	Pair              orderbook.Pair `json:"pair"`
	Quantity          uint64         `json:"quantity"`
	Side              bool           `json:"side"`
	Price             uint64         `json:"price"`
	BlockExpiryWindow uint64         `json:"blockExpiryWindow"`
}

func (ao *AddOrder) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (ao *AddOrder) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (ao *AddOrder) amount(obm *orderbook.OrderbookManager, blockHeight uint64) (uint64, ids.ID) {
	isFilled := false
	getAmount := orderbook.GetAmountFn(ao.Side, isFilled, ao.Pair)
	price := ao.Price
	if price == 0 {
		mid := obm.GetOrderbook(ao.Pair).GetMidPriceBlk(blockHeight)
		// 10% max slippage
		price = mid + mid/10
	}
	amt, tokenID := getAmount(ao.Quantity, price)
	return amt, tokenID
}

func (ao *AddOrder) StateKeys(auth chain.Auth, txID ids.ID) [][]byte {
	user := auth.PublicKey()
	return [][]byte{
		storage.BalanceKey(user, ao.Pair.BaseTokenID),
		storage.BalanceKey(user, ao.Pair.QuoteTokenID),
	}
}

func (ao *AddOrder) Fee(timestamp int64, blockHeight uint64, auth chain.Auth, memoryState any) (amount uint64) {
	obm := memoryState.(*orderbook.OrderbookManager)
	if obm == nil {
		return 0
	}
	user := auth.PublicKey()
	amt, _ := ao.amount(obm, blockHeight)
	return obm.GetOrderbook(ao.Pair).GetFee(user, timestamp, amt)
}

func (ao *AddOrder) Token(memoryState any) (tokenID ids.ID) {
	obm := memoryState.(*orderbook.OrderbookManager)
	if obm == nil {
		return ao.Pair.QuoteTokenID
	}
	_, tokenID = ao.amount(obm, 0)
	return tokenID
}

func (ao *AddOrder) Execute(
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
	if ao.Quantity == 0 {
		err = errors.New("amount cannot be zero")
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if baseBalance, err = storage.PullPendingBalance(ctx, db, obm, user, ao.Pair.BaseTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if quoteBalance, err = storage.PullPendingBalance(ctx, db, obm, user, ao.Pair.QuoteTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if ao.Price == 0 && obm.GetOrderbook(ao.Pair).GetMidPriceBlk(blockHeight) == 0 {
		err = errors.New("mid-price cannot be zero")
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	amount, tokenID := ao.amount(obm, blockHeight)
	var decBalance uint64
	if decBalance, err = storage.DecBalance(ctx, db, user, tokenID, amount); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if tokenID == ao.Pair.BaseTokenID {
		baseBalance = decBalance
	} else {
		quoteBalance = decBalance
	}
	output := utils.PackUpdatedBalance(user, baseBalance, user, quoteBalance)
	return &chain.Result{Success: true, Units: 0, Output: output}, nil
}

func (ao *AddOrder) Marshal(p *codec.Packer) {
	p.PackID(ao.Pair.BaseTokenID)
	p.PackID(ao.Pair.QuoteTokenID)
	p.PackUint64(ao.Quantity)
	p.PackBool(ao.Side)
	p.PackUint64(ao.Price)
	p.PackUint64(ao.BlockExpiryWindow)
}

func UnmarshalAddOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var ao AddOrder
	p.UnpackID(true, &ao.Pair.BaseTokenID)
	p.UnpackID(true, &ao.Pair.QuoteTokenID)
	ao.Quantity = p.UnpackUint64(true)
	ao.Side = p.UnpackBool()
	ao.Price = p.UnpackUint64(false)
	ao.BlockExpiryWindow = p.UnpackUint64(false)
	if ao.BlockExpiryWindow == 0 {
		ao.BlockExpiryWindow = consts.EvictionBlockWindow
	}
	return &ao, p.Err()
}