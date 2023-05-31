package rpc

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Controller interface {
	GetBalance(ctx context.Context, address crypto.PublicKey, tokenID ids.ID) (uint64, error)
	Tracer() trace.Tracer
}