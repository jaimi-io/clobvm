package actions

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/utils"
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

func (t *Transfer) StateKeys(cauth chain.Auth, _ ids.ID) [][]byte {
	user := auth.GetUser(cauth)
	return [][]byte{
		storage.BalanceKey(user, t.TokenID),
		storage.BalanceKey(t.To, t.TokenID),
	}
}

func (t *Transfer) Fee() (amount int64, tokenID ids.ID) {
	return 1, t.TokenID
}

func (t *Transfer) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	timestamp int64,
	cauth chain.Auth,
	txID ids.ID,
	warpVerified bool,
) (result *chain.Result, err error) {
	user := auth.GetUser(cauth)
	if t.Amount == 0 {
		err = errors.New("amount cannot be zero")
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}
	if err = storage.DecBalance(ctx, db, user, t.TokenID, t.Amount); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}
	if err = storage.IncBalance(ctx, db, t.To, t.TokenID, t.Amount); err != nil {
		return &chain.Result{Success: false, Units: 0, Output: utils.ErrBytes(err)}, err
	}
	return &chain.Result{Success: true, Units: 0}, nil
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