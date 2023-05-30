package genesis

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"

	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/crypto"
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

	addr, err := crypto.ParseAddress("clob", "clob1wve8exfmxhzfd2y0jl4aaqyg5yyj3z64k6v5fpe75ck7h7lkg5rscxxzlz")
	if err != nil {
		return err
	}
	c := make([]byte, 32)
	copy(c, []byte("USDC"))
	id, err := ids.ToID(c)
	if err != nil {
		return err
	}
	return storage.SetBalance(ctx, db, addr, id, 1000000000000)
}