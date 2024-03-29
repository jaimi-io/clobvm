package orderbook

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/hypersdk/crypto"
)

type OrderbookManager struct{
	orderbooks map[Pair]*Orderbook
	pendingFunds map[crypto.PublicKey]map[ids.ID]*VersionedBalance
	lastBlockHeight uint64
}

func NewOrderbookManager() *OrderbookManager {
	return &OrderbookManager{
		orderbooks: make(map[Pair]*Orderbook),
		pendingFunds: make(map[crypto.PublicKey]map[ids.ID]*VersionedBalance),
	}
}

func (obm *OrderbookManager) GetOrderbook(pair Pair) *Orderbook {
	if ob, ok := obm.orderbooks[pair]; ok {
		return ob
	}
	ob := NewOrderbook(pair)
	obm.orderbooks[pair] = ob
	return ob
}

func(obm *OrderbookManager) AddPendingFunds(user crypto.PublicKey, tokenID ids.ID, balance uint64, blockHeight uint64) {
	if _, ok := obm.pendingFunds[user]; !ok {
		obm.pendingFunds[user] = make(map[ids.ID]*VersionedBalance)
	}
	if _, ok := obm.pendingFunds[user][tokenID]; !ok {
		obm.pendingFunds[user][tokenID] = NewVersionedBalance(balance, blockHeight)
		return
	}
	obm.pendingFunds[user][tokenID].Put(balance, blockHeight)
}

func (obm *OrderbookManager) PullPendingFunds(user crypto.PublicKey, tokenID ids.ID, blockHeight uint64) uint64 {
	if blockHeight <= consts.PendingBlockWindow {
		return 0
	}
	if _, ok := obm.pendingFunds[user]; !ok {
		return 0
	}
	if _, ok := obm.pendingFunds[user][tokenID]; !ok {
		return 0
	}
	blockHeight -= consts.PendingBlockWindow
	if blockHeight > obm.lastBlockHeight {
		s := fmt.Sprintf("blockHeight %d is greater than lastBlockHeight %d", blockHeight, obm.lastBlockHeight)
		panic(s)
	}
	return obm.pendingFunds[user][tokenID].Pull(blockHeight)
}

func (obm *OrderbookManager) GetPendingFunds(user crypto.PublicKey, tokenID ids.ID, blockHeight uint64) (uint64, uint64) {
	if _, ok := obm.pendingFunds[user]; !ok {
		return 0, blockHeight
	}
	if _, ok := obm.pendingFunds[user][tokenID]; !ok {
		return 0, blockHeight
	}
	return obm.pendingFunds[user][tokenID].Get(blockHeight)
}

func (obm *OrderbookManager) UpdateLastBlockHeight(blockHeight uint64) {
	obm.lastBlockHeight = blockHeight
}
