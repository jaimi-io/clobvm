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

type ED25519 struct {
	Signature crypto.Signature
	From 		  crypto.PublicKey
}

func (e *ED25519) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (e *ED25519) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (e *ED25519) StateKeys() [][]byte {
	return [][]byte{}
}

func (e *ED25519) AsyncVerify(msg []byte) error {
	if !crypto.Verify(msg, e.From, e.Signature) {
		return errors.New("invalid signature")
	}
	return nil
}

func (e *ED25519) Verify(ctx context.Context, r chain.Rules, db chain.Database, action chain.Action) (units uint64, err error) {
	switch a := action.(type) {
		case *actions.AddOrder:
			if a.Quantity % utils.MinQuantity() != 0 {
				return 0, errors.New("invalid quantity for order entered")
			}
	}
	return 0, nil
}

func (e *ED25519) Payer() []byte {
	return e.From[:]
}

func (e *ED25519) PublicKey() crypto.PublicKey {
	return e.From
}

func (e *ED25519) CanDeduct(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	_, bal, err := storage.GetBalance(ctx, db, e.From, tokenID)
	if err != nil {
		return err
	}
	if bal < amount {
		return errors.New("insufficient balance")
	}
	return nil
}

func (e *ED25519) Deduct(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	_, err := storage.DecBalance(ctx, db, e.From, tokenID, amount)
	return err
}

func (e *ED25519) Refund(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	_, err := storage.IncBalance(ctx, db, e.From, tokenID, amount)
	return err
}

func (e *ED25519) Marshal(p *codec.Packer) {
	p.PackSignature(e.Signature)
	p.PackPublicKey(e.From)
}

func UnmarshalEIP712(p *codec.Packer, _ *warp.Message) (chain.Auth, error) {
	var d ED25519
	p.UnpackSignature(&d.Signature)
	p.UnpackPublicKey(true, &d.From)
	return &d, p.Err()
}

func NewE25519Factory(priv crypto.PrivateKey) *E25519Factory {
	return &E25519Factory{priv}
}

type E25519Factory struct {
	priv crypto.PrivateKey
}

func (d *E25519Factory) Sign(msg []byte, a chain.Action) (chain.Auth, error) {
	sig := crypto.Sign(msg, d.priv)
	return &ED25519{sig, d.priv.PublicKey()}, nil
}

