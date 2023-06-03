package orderbook

type VersionedItem struct {
	bal uint64
	blkHgt uint64
}

type VersionedBalance struct {
	items []*VersionedItem
	recent uint64
}

func NewVersionedBalance(balance uint64, blockHeight uint64) *VersionedBalance {
	items := make([]*VersionedItem, 0, 10)
	items = append(items, &VersionedItem{balance, blockHeight})
	return &VersionedBalance{
		items: items,
		recent: balance,
	}
}

func (vb *VersionedBalance) Get(blockHeight uint64) uint64 {
	if len(vb.items) == 0 {
		return 0
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
	vb.recent -= balance
	return balance
}

func (vb *VersionedBalance) Put(amount uint64, blockHeight uint64) {
	if len(vb.items) == 10 {
		vb.items = vb.items[1:]
	}
	prevRecent := vb.recent
	newBalance := prevRecent + amount
	vb.items = append(vb.items, &VersionedItem{newBalance, blockHeight})
	vb.recent = newBalance
}