package orderbook

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/heap"
	"github.com/jaimi-io/hypersdk/crypto"
)

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func getMatchPriceFn(side bool) func(a, b uint64) bool {
	var matchPriceFn func(a, b uint64) bool
	if side {
		matchPriceFn = func (a, b uint64) bool {
			return a <= b
		}
	} else {
		matchPriceFn = func (a, b uint64) bool {
			return a >= b
		}
	}
	return matchPriceFn
}

type OrderStatus struct {
	ID     ids.ID
	User   crypto.PublicKey
	Filled uint64
}

func (ob *Orderbook) matchOrder(order *Order) []*OrderStatus {
	var heap *heap.PriorityQueueHeap[*Order, uint64]
	if order.Side {
		heap = ob.minHeap
	} else {
		heap = ob.maxHeap
	}
	matchPriceFn := getMatchPriceFn(order.Side)
	var orderStatuses []*OrderStatus
	initialQuantity := order.Quantity

	for heap.Len() > 0 && matchPriceFn(heap.Peek().Priority(), order.Price) && 0 < order.Quantity {
		queue := heap.Peek()
		for queue.Len() > 0 && 0 < order.Quantity {
			takerOrder := queue.Peek()
			toFill := min(takerOrder.Quantity, order.Quantity)
			takerOrder.Quantity -= toFill
			order.Quantity -= toFill
			ob.volumeMap[order.Price] -= toFill
			if takerOrder.Quantity == 0 {
				queue.Pop()
			}
			orderStatuses = append(orderStatuses, &OrderStatus{takerOrder.ID, takerOrder.User, toFill})
		}
		if queue.Len() == 0 {
			heap.Pop()
		}
	}
	// i.e. a fill has occurred for this order
	if order.Quantity <= initialQuantity {
		orderStatuses = append(orderStatuses, &OrderStatus{order.ID, order.User, initialQuantity - order.Quantity})
	}
	return orderStatuses
}