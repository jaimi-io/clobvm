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
	"github.com/jaimi-io/hypersdk/crypto"
	hutils "github.com/jaimi-io/hypersdk/utils"
)

type Transfer struct {
	To crypto.PublicKey `json:"to"`
	TokenID ids.ID  	`json:"tokenID"`
	Amount uint64  	 `json:"amount"`
}

func (t *Transfer) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (t *Transfer) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (t *Transfer) StateKeys(auth chain.Auth, _ ids.ID) [][]byte {
	user := auth.PublicKey()
	return [][]byte{
		storage.BalanceKey(user, t.TokenID),
		storage.BalanceKey(t.To, t.TokenID),
	}
}

func (t *Transfer) Fee(timestamp int64, blockHeight uint64, auth chain.Auth, memoryState any) (amount uint64) {
	return 1
}

func (t *Transfer) Token(memoryState any) (tokenID ids.ID) {
	return t.TokenID
}

func (t *Transfer) Execute(
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
	user := auth.PublicKey()
	obm := memoryState.(*orderbook.OrderbookManager)
	var baseBalance uint64
	var quoteBalance uint64
	if t.Amount == 0 {
		err = errors.New("amount cannot be zero")
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if baseBalance, err = storage.PullPendingBalance(ctx, db, obm, user, t.TokenID, blockHeight); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if baseBalance, err = storage.DecBalance(ctx, db, user, t.TokenID, t.Amount); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	if quoteBalance, err = storage.IncBalance(ctx, db, t.To, t.TokenID, t.Amount); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: hutils.ErrBytes(err)}, nil
	}
	output := utils.PackUpdatedBalance(user, baseBalance, t.To, quoteBalance)
	return &chain.Result{Success: true, Units: 0, Output: output}, nil
}

func (t *Transfer) Marshal(p *codec.Packer) {
	p.PackPublicKey(t.To)
	p.PackID(t.TokenID)
	p.PackUint64(t.Amount)
}

func UnmarshalTransfer(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var t Transfer
	p.UnpackPublicKey(true, &t.To)
	p.UnpackID(true, &t.TokenID)
	t.Amount = p.UnpackUint64(true)
	return &t, p.Err()
}