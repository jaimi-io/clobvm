package genesis

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"

	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Genesis struct {
	// Address prefix
	HRP string `json:"hrp"`

	// Block params
	MaxBlockTxs   int    `json:"maxBlockTxs"`
	MaxBlockUnits uint64 `json:"maxBlockUnits"` // must be possible to reach before block too large

	// Tx params
	BaseUnits      uint64 `json:"baseUnits"`
	ValidityWindow int64  `json:"validityWindow"` // seconds

	// Unit pricing
	MinUnitPrice               uint64 `json:"minUnitPrice"`
	UnitPriceChangeDenominator uint64 `json:"unitPriceChangeDenominator"`
	WindowTargetUnits          uint64 `json:"windowTargetUnits"` // 10s

	// Block pricing
	MinBlockCost               uint64 `json:"minBlockCost"`
	BlockCostChangeDenominator uint64 `json:"blockCostChangeDenominator"`
	WindowTargetBlocks         uint64 `json:"windowTargetBlocks"` // 10s

	Rules *Rules
}

func Default() *Genesis {
	return &Genesis{
		HRP: "clob",

		// Block params
		MaxBlockTxs:   20_000,    // rely on max block units
		MaxBlockUnits: math.MaxUint64, // 18 billion worth of order quantity (to review)

		// Tx params
		BaseUnits:      0,
		ValidityWindow: 60,

		// Unit pricing
		MinUnitPrice:               1,
		UnitPriceChangeDenominator: 48,
		WindowTargetUnits:          20_000_000_000,

		// Block pricing
		MinBlockCost:               0,
		BlockCostChangeDenominator: 48,
		WindowTargetBlocks:         1_000_000_000, // 10s
	}
}

func New(b []byte, _ []byte) (*Genesis, error) {
	g := Default()
	if len(b) > 0 {
		if err := json.Unmarshal(b, g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config %s: %w", string(b), err)
		}
	}
	return g, nil
}

func (g *Genesis) GetHRP() string {
	return consts.HRP
}

func (g *Genesis) GetRules() *Rules {
	if g.Rules == nil {
		g.Rules = g.NewRules()
	}
	return g.Rules
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
		decimals := uint64(math.Pow10(consts.BalanceDecimals))
		amt := uint64(10_000_000_000) * decimals
		for _, tokenID := range tokens {
			err = storage.SetBalance(ctx, db, addr, tokenID, amt)
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