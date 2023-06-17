package auth

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	"github.com/jaimi-io/hypersdk/crypto"
)

type EIP712 struct {
	Signature crypto.Signature
	From 		  crypto.PublicKey
	TokenID	 ids.ID
}

func (e *EIP712) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (e *EIP712) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (e *EIP712) StateKeys() [][]byte {
	return [][]byte{
		storage.BalanceKey(e.From, e.TokenID),
	}
}

func (e *EIP712) AsyncVerify(msg []byte) error {
	if !crypto.Verify(msg, e.From, e.Signature) {
		return errors.New("invalid signature")
	}
	return nil
}

func (e *EIP712) Verify(ctx context.Context, r chain.Rules, db chain.Database, action chain.Action) (units uint64, err error) {
	switch a := action.(type) {
		case *actions.AddOrder:
			if a.Quantity % utils.MinQuantity() != 0 {
				return 0, errors.New("invalid quantity for order entered")
			}
	}
	return 0, nil
}

func (e *EIP712) Payer() []byte {
	return e.From[:]
}

func (e *EIP712) PublicKey() crypto.PublicKey {
	return e.From
}

func (e *EIP712) CanDeduct(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	_, bal, err := storage.GetBalance(ctx, db, e.From, tokenID)
	if err != nil {
		return err
	}
	if bal < amount {
		return errors.New("insufficient balance")
	}
	return nil
}

func (e *EIP712) Deduct(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	_, err := storage.DecBalance(ctx, db, e.From, tokenID, amount)
	return err
}

func (e *EIP712) Refund(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	_, err := storage.IncBalance(ctx, db, e.From, tokenID, amount)
	return err
}

func (e *EIP712) Marshal(p *codec.Packer) {
	p.PackSignature(e.Signature)
	p.PackPublicKey(e.From)
	p.PackID(e.TokenID)
}

func UnmarshalEIP712(p *codec.Packer, _ *warp.Message) (chain.Auth, error) {
	var d EIP712
	p.UnpackSignature(&d.Signature)
	p.UnpackPublicKey(true, &d.From)
	p.UnpackID(true, &d.TokenID)
	return &d, p.Err()
}

func NewEIP712Factory(priv crypto.PrivateKey) *EIP712Factory {
	return &EIP712Factory{priv}
}

type EIP712Factory struct {
	priv crypto.PrivateKey
}

func (d *EIP712Factory) Sign(msg []byte, a chain.Action) (chain.Auth, error) {
	sig := crypto.Sign(msg, d.priv)
	return &EIP712{sig, d.priv.PublicKey(), a.Token()}, nil
}

