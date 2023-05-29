package auth

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	"github.com/jaimi-io/hypersdk/crypto"
)

type EIP712 struct {
	Signature crypto.Signature
	From 		  crypto.PublicKey
}

func (e *EIP712) MaxUnits(r chain.Rules) uint64 {
	return 1
}

func (e *EIP712) ValidRange(r chain.Rules) (start int64, end int64) {
	return -1, -1
}

func (e *EIP712) StateKeys() [][]byte {
	return [][]byte{}
}

func (e *EIP712) AsyncVerify(msg []byte) error {
	return nil
}

func (e *EIP712) Verify(ctx context.Context, r chain.Rules, db chain.Database, action chain.Action) (units uint64, err error) {
	return 0, nil
}

func (e *EIP712) Payer() []byte {
	return []byte{}
}

func (e *EIP712) CanDeduct(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	return nil
}

func (e *EIP712) Deduct(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	return nil
}

func (e *EIP712) Refund(ctx context.Context, db chain.Database, amount uint64, tokenID ids.ID) error {
	return nil
}

func (e *EIP712) Marshal(p *codec.Packer) {
	p.PackSignature(e.Signature)
	p.PackPublicKey(e.From)
}

func UnmarshalEIP712(p *codec.Packer, _ *warp.Message) (chain.Auth, error) {
	var d EIP712
	p.UnpackPublicKey(true, &d.From)
	p.UnpackSignature(&d.Signature)
	return &d, p.Err()
}


