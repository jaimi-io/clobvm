package orderbook

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/metrics"
)

func (ob *Orderbook) AddToEviction(orderID ids.ID, blockHeight uint64) {
	blockExpiry := blockHeight + consts.EvictionBlockWindow
	if _, ok := ob.evictionMap[blockExpiry]; !ok {
		ob.evictionMap[blockExpiry] = make(map[ids.ID]struct{})
	}
	ob.evictionMap[blockExpiry][orderID] = struct{}{}
}

func (ob *Orderbook) Evict(blockNumber uint64, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	ordersToEvict := ob.evictionMap[blockNumber]
	if ordersToEvict == nil {
		return
	}
	var i int;
	for orderID := range ordersToEvict {
		order := ob.Get(orderID)
		if order == nil {
			continue
		}
		i++;
		ob.Cancel(order, pendingAmounts, metrics)
	}
	delete(ob.evictionMap, blockNumber)
}

func (obm *OrderbookManager) EvictAllPairs(blockNumber uint64, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	for pair := range obm.orderbooks {
		ob := obm.orderbooks[pair]
		ob.Evict(blockNumber, pendingAmounts, metrics)
	}
}