package orderbook

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/heap"
)

type Orderbook struct {
	pair Pair
	minHeap *heap.PriorityQueueHeap[*Order, uint64]
	maxHeap *heap.PriorityQueueHeap[*Order, uint64]

	orderMap map[ids.ID]*Order
	volumeMap map[uint64]uint64
	evictionMap map[uint64]map[ids.ID]struct{}
}

func NewOrderbook(pair Pair) *Orderbook {
	return &Orderbook{
		pair: pair,
		minHeap: heap.NewPriorityQueueHeap[*Order, uint64](1024, true),
		maxHeap: heap.NewPriorityQueueHeap[*Order, uint64](1024, false),
		orderMap: make(map[ids.ID]*Order),
		volumeMap: make(map[uint64]uint64),
		evictionMap: make(map[uint64]map[ids.ID]struct{}),
	}
}

func (ob *Orderbook) Add(order *Order, blockExpiry uint64, pendingAmounts *[]PendingAmt) {
	ob.matchOrder(order, pendingAmounts)
	if order.Quantity > 0 {
		ob.volumeMap[order.Price] += order.Quantity
		ob.orderMap[order.ID] = order
		ob.AddToEviction(order.ID, blockExpiry)
		if order.Side {
			ob.maxHeap.Add(order, order.ID, order.Price)
		} else {
			ob.minHeap.Add(order, order.ID, order.Price)
		}
	}
}

func (ob *Orderbook) Get(id ids.ID) *Order {
	return ob.orderMap[id]
}

func (ob *Orderbook) Cancel(order *Order, pendingAmounts *[]PendingAmt) {
	ob.volumeMap[order.Price] -= order.Quantity
	if order.Side {
		ob.maxHeap.Remove(order.ID, order.Price)
	} else {
		ob.minHeap.Remove(order.ID, order.Price)
	}
	delete(ob.orderMap, order.ID)
	isFilled := false
	ob.toPendingAmount(order, order.Quantity, isFilled, pendingAmounts)
}

func (ob *Orderbook) Remove(order *Order) {
	ob.volumeMap[order.Price] -= order.Quantity
	delete(ob.orderMap, order.ID)
}

func (ob *Orderbook) GetBuySide() [][]*Order {
	return ob.maxHeap.Values()
}

func (ob *Orderbook) GetSellSide() [][]*Order {
	return ob.minHeap.Values()
}

