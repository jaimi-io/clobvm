package auth

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/hypersdk/chain"
)

type EIP712 struct {}

func (e *EIP712) MaxUnits(r chain.Rules) uint64 {
	return 0
}

func (e *EIP712) ValidRange(r chain.Rules) (start int64, end int64) {
	return 0, 0
}

func (e *EIP712) StateKeys() [][]byte {
	return nil
}

func (e *EIP712) AsyncVerify(msg []byte) error {
	return nil
}

func (e *EIP712) Verify(ctx context.Context, r chain.Rules, db chain.Database, action chain.Action) (units uint64, err error) {
	return 0, nil
}

func (e *EIP712) Payer() []byte {
	return nil
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


