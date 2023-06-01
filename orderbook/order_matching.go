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
	Amount uint64
}

func (ob *Orderbook) matchOrder(order *Order) ([]*OrderStatus, int) {
	var heap *heap.PriorityQueueHeap[*Order, uint64]
	if order.Side {
		heap = ob.minHeap
	} else {
		heap = ob.maxHeap
	}
	matchPriceFn := getMatchPriceFn(order.Side)
	getAmount := GetAmountFn(order.Side, false)
	// TODO: max fills constant + throw error if exceeded?
	// TODO: group amounts by user
	orderStatuses := make([]*OrderStatus, 1024)
	matchIndex := 0
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
			orderStatuses[matchIndex] = &OrderStatus{takerOrder.ID, takerOrder.User, getAmount(toFill, takerOrder.Price)}
			matchIndex++
			if matchIndex == 1024 {
				return orderStatuses, -1
			}
		}
		if queue.Len() == 0 {
			heap.Pop()
		}
	}
	// i.e. a fill has occurred for this order
	if order.Quantity <= initialQuantity {
		orderStatuses = append(orderStatuses, &OrderStatus{order.ID, order.User, initialQuantity - order.Quantity})
	}
	return orderStatuses, matchIndex
}