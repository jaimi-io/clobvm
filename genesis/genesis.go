package genesis

import (
	"context"

	"github.com/ava-labs/avalanchego/trace"

	"github.com/jaimi-io/hypersdk/chain"
)

type Genesis struct {}

func New() *Genesis {
	return &Genesis{}
}

func (g *Genesis) GetHRP() string {
	return "clob"
}

func (g *Genesis) Load(ctx context.Context, tracer trace.Tracer, db chain.Database) error {
	ctx, span := tracer.Start(ctx, "genesis.Load")
	defer span.End()

	// TODO: add token amounts
	return nil
}