package genesis

import (
	"context"

	"github.com/ava-labs/avalanchego/trace"

	"github.com/jaimi-io/hypersdk/chain"
)

type Genesis struct {}

func (g *Genesis) GetHRP() string {
	return ""
}

func (g *Genesis) Load(context.Context, trace.Tracer, chain.Database) error {
	return nil
}