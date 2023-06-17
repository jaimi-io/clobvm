package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	hutils "github.com/jaimi-io/hypersdk/utils"
)

type MarketOrder struct {
	Pair              orderbook.Pair `json:"pair"`
	Quantity          uint64         `json:"quantity"`
	Side              bool           `json:"side"`
}

func (ao *MarketOrder) MaxUnits(r chain.Rules) uint64 {
	return ao.Quantity / utils.MinQuantity()
}

func (ao *MarketOrder) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (ao *MarketOrder) amount(obm *orderbook.OrderbookManager) (uint64, ids.ID) {
	isFilled := false
	getAmount := orderbook.GetAmountFn(ao.Side, isFilled, ao.Pair)
	return getAmount(ao.Quantity, obm.GetOrderbook(ao.Pair).GetMidPrice())
}

func (ao *MarketOrder) StateKeys(auth chain.Auth, txID ids.ID) [][]byte {
	user := auth.PublicKey()
	return [][]byte{
		storage.BalanceKey(user, ao.Pair.BaseTokenID),
		storage.BalanceKey(user, ao.Pair.QuoteTokenID),
	}
}

func (ao *MarketOrder) Fee(timestamp int64, auth chain.Auth, memoryState any) (amount uint64) {
	obm := memoryState.(*orderbook.OrderbookManager)
	if obm == nil {
		return 0
	}
	user := auth.PublicKey()
	amt, _ := ao.amount(obm)
	return obm.GetOrderbook(ao.Pair).GetFee(user, timestamp, amt)
}

func (mo *MarketOrder) Token(memoryState any) (tokenID ids.ID) {
	obm := memoryState.(*orderbook.OrderbookManager)
	if obm == nil {
		return mo.Pair.QuoteTokenID
	}
	_, tokenID = mo.amount(obm)
	return tokenID
}

func (mo *MarketOrder) Execute(
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
	if mo.Quantity == 0 {
		err = errors.New("amount cannot be zero")
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if baseBalance, err = storage.PullPendingBalance(ctx, db, obm, user, mo.Pair.BaseTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if quoteBalance, err = storage.PullPendingBalance(ctx, db, obm, user, mo.Pair.QuoteTokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	amount, tokenID := mo.amount(obm)
	var decBalance uint64
	if decBalance, err = storage.DecBalance(ctx, db, user, tokenID, amount); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if tokenID == mo.Pair.BaseTokenID {
		baseBalance = decBalance
	} else {
		quoteBalance = decBalance
	}
	output := utils.PackUpdatedBalance(user, baseBalance, user, quoteBalance)
	return &chain.Result{Success: true, Units: mo.Quantity / utils.MinQuantity(), Output: output}, nil
}

func (mo *MarketOrder) Marshal(p *codec.Packer) {
	p.PackID(mo.Pair.BaseTokenID)
	p.PackID(mo.Pair.QuoteTokenID)
	p.PackUint64(mo.Quantity)
	p.PackBool(mo.Side)
}

func UnmarshalMarketOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var mo MarketOrder
	p.UnpackID(true, &mo.Pair.BaseTokenID)
	p.UnpackID(true, &mo.Pair.QuoteTokenID)
	mo.Quantity = p.UnpackUint64(true)
	mo.Side = p.UnpackBool()
	return &mo, p.Err()
}