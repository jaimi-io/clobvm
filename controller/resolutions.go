package controller

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/storage"
	"github.com/jaimi-io/hypersdk/crypto"
)

func (c *Controller) Genesis() (*genesis.Genesis) {
	return c.genesis
}

func (c *Controller) GetBalance(ctx context.Context, pk crypto.PublicKey, tokenID ids.ID) (uint64, error) {
	return storage.GetBalanceFromState(ctx, c.inner.ReadState, pk, tokenID)
}

func (c *Controller) GetOrderbook(ctx context.Context, pair orderbook.Pair) (string, string, error) {
	ob := c.orderbookManager.GetOrderbook(pair)
	if ob == nil {
		return "", "", fmt.Errorf("orderbook not found for pair %s", pair)
	}
	buySide := ob.GetBuySide()
	sellSide := ob.GetSellSide()
	return fmt.Sprint(buySide), fmt.Sprint(sellSide), nil
}

func (c *Controller) GetMidPrice(ctx context.Context, pair orderbook.Pair) (uint64, error) {
	ob := c.orderbookManager.GetOrderbook(pair)
	if ob == nil {
		return 0, fmt.Errorf("orderbook not found for pair %s", pair)
	}
	midPrice := ob.GetMidPrice()
	return midPrice, nil
}


func (c *Controller) GetPendingFunds(ctx context.Context, user crypto.PublicKey, tokenID ids.ID, blockHeight uint64) (uint64, uint64) {
	return c.orderbookManager.GetPendingFunds(user, tokenID, blockHeight)
}

func (c *Controller) GetVolumes(ctx context.Context, pair orderbook.Pair) (string, error) {
	ob := c.orderbookManager.GetOrderbook(pair)
	if ob == nil {
		return "", fmt.Errorf("orderbook not found for pair %s", pair)
	}
	return ob.GetVolumes(), nil
}
