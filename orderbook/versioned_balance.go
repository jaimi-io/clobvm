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
	items := ring.New(int(consts.PendingBlockWindow))
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
	for i := 0; i < vb.items.Len() - 1 && cur.Value != nil; i++ {
		item := cur.Value.(*VersionedItem)
		if item.blkHgt <= blockHeight {
			balance = item.bal
			break
		}
		cur = cur.Prev()
	}

	return balance, blockHeight
}

func (vb *VersionedBalance) Pull(blockHeight uint64) uint64 {
	if blockHeight > vb.lastBlockHeight {
		res := vb.lastBalance
		vb.lastBalance = 0
		return res
	}
	var balance uint64

	cur := vb.items
	for i := 0; i < vb.items.Len() - 1 && cur.Value != nil; i++ {
		item := cur.Value.(*VersionedItem)
		if item.blkHgt <= blockHeight {
			balance = item.bal
			break
		}
		cur = cur.Prev()
	}

	for cur != vb.items && cur.Next().Value != nil {
		cur = cur.Next()
		item := cur.Value.(*VersionedItem)
		item.bal -= balance
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