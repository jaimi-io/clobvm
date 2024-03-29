package rpc

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Controller interface {
	Genesis() (*genesis.Genesis)
	GetBalance(ctx context.Context, address crypto.PublicKey, tokenID ids.ID) (uint64, error)
	GetMidPrice(ctx context.Context, pair orderbook.Pair) (uint64, error)
	GetOrderbook(ctx context.Context, pair orderbook.Pair, numPriceLevels int) (string, string, error)
	GetPendingFunds(ctx context.Context, user crypto.PublicKey, tokenID ids.ID, blockHeight uint64) (uint64, uint64)
	GetVolumes(ctx context.Context, pair orderbook.Pair, numPriceLevels int) (string, error)
	Tracer() trace.Tracer
}