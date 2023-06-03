package orderbook

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/heap"
	"github.com/jaimi-io/hypersdk/crypto"
)

func GetAmountFn(side bool, isFilled bool) func (q, p uint64) uint64 {
	if side && !isFilled || !side && isFilled {
		return func(q, p uint64) uint64 { return q * p}
	}
	return func(q, p uint64) uint64 { return q }
}

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

func (ob *Orderbook) matchOrder(order *Order, tokenID ids.ID, oppTokenID ids.ID, pendingAmounts *[]PendingAmt) {
	var heap *heap.PriorityQueueHeap[*Order, uint64]
	if order.Side {
		heap = ob.minHeap
	} else {
		heap = ob.maxHeap
	}
	matchPriceFn := getMatchPriceFn(order.Side)
	prevQuantity := order.Quantity
	isFilled := true

	for heap.Len() > 0 && matchPriceFn(heap.Peek().Priority(), order.Price) && 0 < order.Quantity {
		queue := heap.Peek()
		for queue.Len() > 0 && 0 < order.Quantity {
			takerOrder := queue.Peek()
			toFill := min(takerOrder.Quantity, order.Quantity)
			takerOrder.Quantity -= toFill
			order.Quantity -= toFill
			ob.volumeMap[order.Price] -= toFill
			if takerOrder.Quantity == 0 {
				ob.Remove(queue.Pop())
			}
			// TODO: avg price for the order that gets added
			ob.toPendingAmount(takerOrder, tokenID, toFill, isFilled, pendingAmounts)
		}

		if queue.Len() == 0 {
			heap.Pop()
		}
	}
	if prevQuantity > order.Quantity {
		ob.toPendingAmount(order, oppTokenID, prevQuantity - order.Quantity, isFilled, pendingAmounts)
	} 
}

type PendingAmt struct {
	User    crypto.PublicKey
	TokenID ids.ID
	Amount  uint64
}

func (ob *Orderbook) toPendingAmount(order *Order, tokenID ids.ID, quantity uint64, isFilled bool, pendingAmounts *[]PendingAmt) {
	getAmount := GetAmountFn(order.Side, isFilled)
	*pendingAmounts = append(*pendingAmounts, PendingAmt{order.User, tokenID, getAmount(quantity, order.Price)})
}