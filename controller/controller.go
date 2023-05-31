package controller

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/version"
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/registry"
	"github.com/jaimi-io/clobvm/rpc"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/hypersdk/config"
	"github.com/jaimi-io/hypersdk/crypto"

	"github.com/jaimi-io/hypersdk/builder"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/gossiper"
	"github.com/jaimi-io/hypersdk/pebble"
	hyperrpc "github.com/jaimi-io/hypersdk/rpc"
	"github.com/jaimi-io/hypersdk/utils"
	"github.com/jaimi-io/hypersdk/vm"
)

type Controller struct {
	inner *vm.VM
	orderbook *orderbook.Orderbook

	snowCtx *snow.Context
	stateManager *StateManager
	config *config.Config
	rules *genesis.Rules
}

func New() *vm.VM {
	return vm.New(&Controller{}, &version.Semantic{Major: 0,
		Minor: 0,
		Patch: 1,})
}

func (c *Controller) Initialize(
	inner *vm.VM, // hypersdk VM
	snowCtx *snow.Context,
	gatherer metrics.MultiGatherer,
	genesisBytes []byte,
	upgradeBytes []byte,
	configBytes []byte,
) (
	vm.Config,
	vm.Genesis,
	builder.Builder,
	gossiper.Gossiper,
	database.Database,
	database.Database,
	vm.Handlers,
	chain.ActionRegistry,
	chain.AuthRegistry,
	error,
) {
	c.inner = inner
	c.stateManager = &StateManager{}
	c.config = &config.Config{}
	c.rules = &genesis.Rules{}
	c.orderbook = orderbook.NewOrderbook()
	bcfg := builder.DefaultTimeConfig()
	//bcfg.PreferredBlocksPerSecond = c.config.GetPreferredBlocksPerSecond()
	build := builder.NewTime(inner, bcfg)
	gcfg := gossiper.DefaultProposerConfig()
	// gcfg.GossipInterval = c.config.GossipInterval
	// gcfg.GossipMaxSize = c.config.GossipMaxSize
	// gcfg.GossipProposerDiff = c.config.GossipProposerDiff
	// gcfg.GossipProposerDepth = c.config.GossipProposerDepth
	// gcfg.BuildProposerDiff = c.config.BuildProposerDiff
	// gcfg.VerifyTimeout = c.config.VerifyTimeout
	gossip := gossiper.NewProposer(inner, gcfg)
	blockPath, err := utils.InitSubDirectory(snowCtx.ChainDataDir, "block")
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	cfg := pebble.NewDefaultConfig()
	blockDB, err := pebble.New(blockPath, cfg)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	statePath, err := utils.InitSubDirectory(snowCtx.ChainDataDir, "state")
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	stateDB, err := pebble.New(statePath, cfg)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	apis := map[string]*common.HTTPHandler{}
	jsonRPCHandler, err := hyperrpc.NewJSONRPCHandler(
		"clobvm",
		rpc.NewRPCServer(c),
		common.NoLock,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	apis[rpc.JSONRPCEndpoint] = jsonRPCHandler
	inner.Logger().Info("Returning from controller.Initialize")
	return c.config, genesis.New(), build, gossip, blockDB, stateDB, apis, registry.ActionRegistry, registry.AuthRegistry, err
}

func (c *Controller) Rules(t int64) chain.Rules {
	return c.rules
}

func (c *Controller) StateManager() chain.StateManager {
	return c.stateManager
}

func (c *Controller) Accepted(ctx context.Context, blk *chain.StatelessBlock) error {
	results := blk.Results()
	for i, tx := range blk.Txs {
		result := results[i]
		if result.Success {
			switch action := tx.Action.(type) {
			case *actions.AddOrder:
				fmt.Println("AddOrder: ", tx.ID())
				order := orderbook.NewOrder(tx.ID(), action.Price, action.Quantity, action.Side)
				c.orderbook.Add(order)
			case *actions.CancelOrder:
				fmt.Println("CancelOrder: ", tx.ID())
				order := c.orderbook.Get(action.OrderID)
				c.orderbook.Remove(order)
			}
		}
		fmt.Println("Res: ", result, " Tx: ", tx.ID())
	}
	return nil
}

func (c *Controller) Rejected(ctx context.Context, blk *chain.StatelessBlock) error {
	return nil
}

func (c *Controller) Shutdown(context.Context) error {
	return nil
}

func (c *Controller) GetBalance(ctx context.Context, pk crypto.PublicKey, tokenID ids.ID) (uint64, error) {
	return storage.GetBalanceFromState(ctx, c.inner.ReadState, pk, tokenID)
}

func (c *Controller) GetOrderbook(ctx context.Context) (string, string, error) {
	buySide := c.orderbook.GetBuySide()
	sellSide := c.orderbook.GetSellSide()
	return fmt.Sprint(buySide), fmt.Sprint(sellSide), nil
}

func (c *Controller) Tracer() trace.Tracer {
	return c.inner.Tracer()
}
