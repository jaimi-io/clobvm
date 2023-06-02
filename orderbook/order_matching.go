package orderbook

import (
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

func (ob *Orderbook) addToFilled(side bool, user crypto.PublicKey, amount uint64) {
	if side {
		ob.filledBuys[user] += amount
	} else {
		ob.filledSells[user] += amount
	}
}

func (ob *Orderbook) matchOrder(order *Order) {
	var heap *heap.PriorityQueueHeap[*Order, uint64]
	if order.Side {
		heap = ob.minHeap
	} else {
		heap = ob.maxHeap
	}
	matchPriceFn := getMatchPriceFn(order.Side)
	getAmount := GetAmountFn(!order.Side, false)
	prevQuantity := order.Quantity

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
			// TODO: avg price for the order that gets added
			ob.addToFilled(takerOrder.Side, takerOrder.User, getAmount(toFill, takerOrder.Price))
		}

		if queue.Len() == 0 {
			heap.Pop()
		}
	}
	if prevQuantity > order.Quantity {
		getAmount := GetAmountFn(order.Side, false)
		// TODO: avg price for the order that gets added
		ob.addToFilled(order.Side, order.User, getAmount(prevQuantity - order.Quantity, order.Price))
	} 
}