package orderbook

import (
	"fmt"

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

func (ob *Orderbook) matchOrder(order *Order, tokenID ids.ID, oppTokenID ids.ID, pendingAmounts *[]PendingAmt) {
	var heap *heap.PriorityQueueHeap[*Order, uint64]
	if order.Side {
		heap = ob.minHeap
	} else {
		heap = ob.maxHeap
	}
	matchPriceFn := getMatchPriceFn(order.Side)
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
				ob.Remove(queue.Pop())
			}
			// TODO: avg price for the order that gets added
			ob.toPendingAmount(takerOrder, oppTokenID, pendingAmounts)
		}

		if queue.Len() == 0 {
			heap.Pop()
		}
	}
	if prevQuantity > order.Quantity {
		ob.toPendingAmount(order, tokenID, pendingAmounts)
	} 
}

type PendingAmt struct {
	User    crypto.PublicKey
	TokenID ids.ID
	Amount  uint64
}

func (ob *Orderbook) toPendingAmount(order *Order, tokenID ids.ID, pendingAmounts *[]PendingAmt) {
	getAmount := GetAmountFn(order.Side, false)
	fmt.Println("toPendingAmount.append:")
	*pendingAmounts = append(*pendingAmounts, PendingAmt{order.User, tokenID, getAmount(order.Quantity, order.Price)})
	fmt.Println("length of pendingAmounts:", len(*pendingAmounts))
}