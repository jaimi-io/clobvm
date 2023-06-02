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

func distributeTokens(ctx context.Context, db chain.Database) error {
	b := make([]byte, 32)
	copy(b, []byte("AVAX"))
	avaxID, _ := ids.ToID(b)
	copy(b, []byte("USDC"))
	usdcID, _ := ids.ToID(b)
	tokens := []ids.ID {
		avaxID,
		usdcID,
	}
	addressess := []string{
		"clob1wve8exfmxhzfd2y0jl4aaqyg5yyj3z64k6v5fpe75ck7h7lkg5rscxxzlz",
		"clob12l2xyad754fu3s9rqdwq4mkllnl0vc5yerygn2aw5xasmjhzmtwspkx4ek",
	}
	for _, strAddr := range addressess {
		addr, err := crypto.ParseAddress("clob", strAddr)
		if err != nil {
			return err
		}
		for _, tokenID := range tokens {
			err = storage.SetBalance(ctx, db, addr, tokenID, 1000000000000)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Genesis) Load(ctx context.Context, tracer trace.Tracer, db chain.Database) error {
	ctx, span := tracer.Start(ctx, "genesis.Load")
	defer span.End()

	return distributeTokens(ctx, db)
}