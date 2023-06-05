package orderbook

import "github.com/jaimi-io/clobvm/consts"

type VersionedItem struct {
	bal uint64
	blkHgt uint64
}

type VersionedBalance struct {
	items []*VersionedItem
	lastBalance uint64
	lastBlockHeight uint64
}


func NewVersionedBalance(balance uint64, blockHeight uint64) *VersionedBalance {
	items := make([]*VersionedItem, 0, consts.PendingBlockWindow)
	items = append(items, &VersionedItem{balance, blockHeight})
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

	for i := len(vb.items)-1; i >= 0; i-- {
		if vb.items[i].blkHgt <= blockHeight {
			balance = vb.items[i].bal
			break
		}
	}

	return balance, blockHeight
}

func (vb *VersionedBalance) Pull(blockHeight uint64) uint64 {
	if blockHeight > vb.lastBlockHeight {
		vb.items = vb.items[0:]
		res := vb.lastBalance
		vb.lastBalance = 0
		return res
	}
	var balance uint64
	var nextIndex int

	for i := len(vb.items)-1; i >= 0; i-- {
		if vb.items[i].blkHgt <= blockHeight {
			balance = vb.items[i].bal
			if i > len(vb.items)-1 {
				// only slice if not the last item
				nextIndex = i+1
			}
			break
		}
	}

	vb.items = vb.items[nextIndex:]
	for j := 0; j < len(vb.items); j++ {
		vb.items[j].bal -= balance
	}
	vb.lastBalance -= balance
	return balance
}

func (vb *VersionedBalance) Put(amount uint64, blockHeight uint64) {
	if len(vb.items) == int(consts.PendingBlockWindow) {
		vb.items = vb.items[1:]
	}
	newBalance := vb.lastBalance + amount
	vb.items = append(vb.items, &VersionedItem{newBalance, blockHeight})
	vb.lastBalance = newBalance
	vb.lastBlockHeight = blockHeight
}