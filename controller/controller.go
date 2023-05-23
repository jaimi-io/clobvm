package controller

import (
	"context"

	"github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/jaimi-io/hypersdk/builder"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/gossiper"
	"github.com/jaimi-io/hypersdk/vm"
)

type Controller struct {

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
	return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil
}

func (c *Controller) Rules(t int64) chain.Rules {
	return nil
}

func (c *Controller) StateManager() chain.StateManager {
	return nil
}

func (c *Controller) Accepted(ctx context.Context, blk *chain.StatelessBlock) error {
	return nil
}

func (c *Controller) Rejected(ctx context.Context, blk *chain.StatelessBlock) error {
	return nil
}

func (c *Controller) Shutdown(context.Context) error {
	return nil
}
