package orderbook

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
)

func (ob *Orderbook) AddToEviction(orderID ids.ID, blockHeight uint64) {
	blockExpiry := blockHeight + consts.EvictionBlockWindow
	if _, ok := ob.evictionMap[blockExpiry]; !ok {
		ob.evictionMap[blockExpiry] = make(map[ids.ID]struct{})
	}
	ob.evictionMap[blockExpiry][orderID] = struct{}{}
}

func (ob *Orderbook) Evict(blockNumber uint64, pendingAmounts *[]PendingAmt) {
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
		ob.Cancel(order, pendingAmounts)
	}
	delete(ob.evictionMap, blockNumber)
}

func (obm *OrderbookManager) EvictAllPairs(blockNumber uint64, pendingAmounts *[]PendingAmt) {
	for _, pair := range obm.pairs {
		ob := obm.orderbooks[pair]
		ob.Evict(blockNumber, pendingAmounts)
	}
}