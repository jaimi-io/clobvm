package orderbook

import "github.com/jaimi-io/clobvm/heap"

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

func (ob *Orderbook) matchOrder(order *Order){
	var heap *heap.PriorityQueueHeap[*Order, uint64]
	if order.Side {
		heap = ob.minHeap
	} else {
		heap = ob.maxHeap
	}
	matchPriceFn := getMatchPriceFn(order.Side)

	for heap.Len() > 0 && matchPriceFn(heap.Peek().Priority(), order.Price) && 0 < order.Quantity {
		queue := heap.Peek()
		for queue.Len() > 0 && 0 < order.Quantity {
			takerOrder := queue.Peek()
			toFill := min(takerOrder.Quantity, order.Quantity)
			takerOrder.Quantity -= toFill
			order.Quantity -= toFill
			if takerOrder.Quantity == 0 {
				queue.Pop()
			}
		}
		if queue.Len() == 0 {
			heap.Pop()
		}
	}
}