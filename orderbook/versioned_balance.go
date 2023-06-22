package orderbook

import (
	"container/ring"

	"github.com/jaimi-io/clobvm/consts"
)

type VersionedItem struct {
	bal uint64
	blkHgt uint64
}

type VersionedBalance struct {
	items *ring.Ring
	lastBalance uint64
	lastBlockHeight uint64
}


func NewVersionedBalance(balance uint64, blockHeight uint64) *VersionedBalance {
	items := ring.New(int(consts.PendingBlockWindow+1))
	items.Value = &VersionedItem{balance, blockHeight}
	return &VersionedBalance{
		items: items,
		lastBalance: balance,
		lastBlockHeight: blockHeight,
	}
}

func (vb *VersionedBalance) Get(blockHeight uint64) (uint64, uint64) {
	if blockHeight > vb.lastBlockHeight {
		return vb.lastBalance, vb.lastBlockHeight
	}
	var balance uint64

	cur := vb.items
	for i := 0; i < vb.items.Len() && cur.Value != nil && cur.Value.(*VersionedItem).blkHgt > blockHeight; i++ {
		cur = cur.Prev()
	}

	if cur.Value != nil {
		balance = cur.Value.(*VersionedItem).bal
	}

	return balance, blockHeight
}

func (vb *VersionedBalance) Pull(blockHeight uint64) uint64 {
	if blockHeight >= vb.lastBlockHeight {
		vb.items.Value = nil
		res := vb.lastBalance
		vb.lastBalance = 0
		return res
	}
	var balance uint64

	cur := vb.items
	for i := 0; i < vb.items.Len() && cur.Value != nil && cur.Value.(*VersionedItem).blkHgt > blockHeight; i++ {
		cur = cur.Prev()
	}

	if cur.Value != nil {
		balance = cur.Value.(*VersionedItem).bal
		cur.Value = nil
	}

	cur = vb.items
	for cur.Value != nil {
		item := cur.Value.(*VersionedItem)
		item.bal -= balance
		cur = cur.Prev()
	}

	vb.lastBalance -= balance
	return balance
}

func (vb *VersionedBalance) Put(amount uint64, blockHeight uint64) {
	newBalance := vb.lastBalance + amount
	next := vb.items.Next()
	next.Value = &VersionedItem{newBalance, blockHeight}

	vb.items = next
	vb.lastBalance = newBalance
	vb.lastBlockHeight = blockHeight
}