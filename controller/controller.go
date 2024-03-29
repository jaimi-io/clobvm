package controller

import (
	"context"
	"time"

	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/version"
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/clobvm/metrics"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/registry"
	"github.com/jaimi-io/clobvm/rpc"
	"github.com/jaimi-io/hypersdk/config"
	"github.com/jaimi-io/hypersdk/crypto"

	"github.com/jaimi-io/hypersdk/builder"
	"github.com/jaimi-io/hypersdk/chain"
	hconsts "github.com/jaimi-io/hypersdk/consts"
	"github.com/jaimi-io/hypersdk/gossiper"
	"github.com/jaimi-io/hypersdk/pebble"
	hrpc "github.com/jaimi-io/hypersdk/rpc"
	"github.com/jaimi-io/hypersdk/utils"
	"github.com/jaimi-io/hypersdk/vm"
)

type Controller struct {
	inner *vm.VM
	orderbookManager *orderbook.OrderbookManager

	metrics *metrics.Metrics

	snowCtx *snow.Context
	stateManager *StateManager
	config *config.Config
	genesis *genesis.Genesis
}

func New() *vm.VM {
	return vm.New(&Controller{}, &version.Semantic{Major: 0,
		Minor: 0,
		Patch: 1,})
}

func (c *Controller) Initialize(
	inner *vm.VM, // hypersdk VM
	snowCtx *snow.Context,
	gatherer ametrics.MultiGatherer,
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
	any,
	error,
) {
	c.inner = inner
	c.stateManager = &StateManager{}
	c.config = &config.Config{}
	var err error
	c.metrics, err = metrics.NewMetrics(gatherer)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	c.genesis, err = genesis.New(genesisBytes, upgradeBytes)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	c.orderbookManager = orderbook.NewOrderbookManager()
	bcfg := builder.DefaultTimeConfig()
	bcfg.PreferredBlocksPerSecond = 3
	build := builder.NewTime(inner, bcfg)
	gcfg := gossiper.DefaultProposerConfig()
	gcfg.GossipInterval = 1 * time.Second
	gcfg.GossipMaxSize = hconsts.NetworkSizeLimit
	gcfg.GossipProposerDiff = 3
	gcfg.GossipProposerDepth = 1
	gcfg.BuildProposerDiff = 1
	gcfg.VerifyTimeout = 5
	gossip := gossiper.NewProposer(inner, gcfg)
	blockPath, err := utils.InitSubDirectory(snowCtx.ChainDataDir, "block")
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	cfg := pebble.NewDefaultConfig()
	blockDB, err := pebble.New(blockPath, cfg)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	statePath, err := utils.InitSubDirectory(snowCtx.ChainDataDir, "state")
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	stateDB, err := pebble.New(statePath, cfg)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	apis := map[string]*common.HTTPHandler{}
	jsonRPCHandler, err := hrpc.NewJSONRPCHandler(
		"clobvm",
		rpc.NewRPCServer(c),
		common.NoLock,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	apis[consts.JSONRPCEndpoint] = jsonRPCHandler
	inner.Logger().Info("Returning from controller.Initialize")
	return c.config, c.genesis, build, gossip, blockDB, stateDB, apis, registry.ActionRegistry, registry.AuthRegistry, c.orderbookManager, err
}

func (c *Controller) Rules(t int64) chain.Rules {
	return c.genesis.GetRules()
}

func (c *Controller) StateManager() chain.StateManager {
	return c.stateManager
}

func (c *Controller) Accepted(ctx context.Context, blk *chain.StatelessBlock) error {
	start := time.Now()
	results := blk.Results()
	var pendingAmounts []orderbook.PendingAmt
	pendingAmtPtr := &pendingAmounts

	c.orderbookManager.EvictAllPairs(blk.Hght, pendingAmtPtr, c.metrics)

	for i, tx := range blk.Txs {
		result := results[i]
		if result.Success {
			addr := tx.Auth.PublicKey()
			switch action := tx.Action.(type) {
			case *actions.AddOrder:
				c.metrics.AddOrder()
				order := orderbook.NewOrder(tx.ID(), addr, action.Price, action.Quantity, action.Side, blk.Hght, action.BlockExpiryWindow)
				ob := c.orderbookManager.GetOrderbook(action.Pair)
				ob.Add(order, blk.Hght, blk.Tmstmp, pendingAmtPtr, c.metrics)
			case *actions.CancelOrder:
				c.metrics.CancelOrder()
				orderbook := c.orderbookManager.GetOrderbook(action.Pair)
				if action.OrderID == ids.Empty {
					orderbook.CancelAll(addr, pendingAmtPtr, c.metrics)
				} else {
					order := orderbook.Get(action.OrderID)
					if order != nil || order.User == addr {
						orderbook.Cancel(order, pendingAmtPtr, c.metrics)
					}
				}
			case *actions.Transfer:
				c.metrics.Transfer()
			}
		}
	}

	fundsPerUser := make(map[crypto.PublicKey]map[ids.ID]uint64)
	for _, pendingAmt := range pendingAmounts {
		if _, ok := fundsPerUser[pendingAmt.User]; !ok {
			fundsPerUser[pendingAmt.User] = make(map[ids.ID]uint64)
		}
		fundsPerUser[pendingAmt.User][pendingAmt.TokenID] += pendingAmt.Amount
	}

	for user, tokenBalances := range fundsPerUser {
		for tokenID, balance := range tokenBalances {
			c.orderbookManager.AddPendingFunds(user, tokenID, balance, blk.Hght)
		}
	}
	c.orderbookManager.UpdateAllMidPrices(blk.Hght)
	c.orderbookManager.UpdateLastBlockHeight(blk.Hght)
	c.metrics.ObserverOrderProcessing(time.Since(start))
	return nil
}

func (c *Controller) Rejected(ctx context.Context, blk *chain.StatelessBlock) error {
	return nil
}

func (c *Controller) Shutdown(context.Context) error {
	return nil
}

func (c *Controller) Tracer() trace.Tracer {
	return c.inner.Tracer()
}
