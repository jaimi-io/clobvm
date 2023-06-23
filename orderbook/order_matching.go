package orderbook

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/heap"
	"github.com/jaimi-io/clobvm/metrics"
	"github.com/jaimi-io/clobvm/queue"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/crypto"
)

func GetAmountFn(side bool, isFilled bool, pair Pair) func (q, p uint64) (uint64, ids.ID) {
	if side && !isFilled || !side && isFilled {
		return func(q, p uint64) (uint64, ids.ID) { return p * q / utils.MinPrice(), pair.QuoteTokenID }
	}
	return func(q, p uint64) (uint64, ids.ID) { return q, pair.BaseTokenID }
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

func (ob *Orderbook) getOppositeHeap(side bool) *heap.PriorityQueueHeap[*Order, uint64] {
	if side {
		return ob.minHeap
	}
	return ob.maxHeap
}

func (ob *Orderbook) fillPriceLevel(heap *heap.PriorityQueueHeap[*Order, uint64], order *Order, blockTs int64, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) uint64 {
	queue := heap.Peek()
	var filledQuote uint64
	for queue.Len() > 0 && 0 < order.Quantity {
		takerOrder := queue.Peek()
		toFill := min(takerOrder.Quantity, order.Quantity)
		takerOrder.Quantity -= toFill
		order.Quantity -= toFill
		ob.volumeMap[takerOrder.Price] -= toFill

		if takerOrder.Quantity == 0 {
			ob.Remove(queue.Pop(), metrics)
		}

		ob.fillAmount(takerOrder, toFill, pendingAmounts)
		ob.addExec(takerOrder.User, blockTs, toFill)
		filledQuote += takerOrder.Price * toFill
		metrics.OrderAmountSub(toFill)
		metrics.OrderFillsNum()
		metrics.OrderFillsAmount(toFill)
	}

	if queue.Len() == 0 {
		heap.PopQueue()
	}

	return filledQuote
}

func (ob *Orderbook) fillMaker(order *Order, blockTs int64, prevQuantity uint64, filledQuote uint64, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	filledQuantity := prevQuantity - order.Quantity
	oldOrderPrice := order.Price
	order.Price = filledQuote / filledQuantity
	ob.fillAmount(order, filledQuantity, pendingAmounts)
	order.Price = oldOrderPrice
	ob.addExec(order.User, blockTs, filledQuantity)
	metrics.OrderFillsNum()
	metrics.OrderFillsAmount(filledQuantity)
}

func (ob *Orderbook) matchLimitOrder(order *Order, blockTs int64, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	heap := ob.getOppositeHeap(order.Side)
	matchPriceFn := getMatchPriceFn(order.Side)
	var filledQuote uint64
	prevQuantity := order.Quantity

	for heap.Len() > 0 && matchPriceFn(heap.Peek().Priority(), order.Price) && 0 < order.Quantity {
		filledQuote += ob.fillPriceLevel(heap, order, blockTs, pendingAmounts, metrics)
	}

	if prevQuantity > order.Quantity {
		ob.fillMaker(order, blockTs, prevQuantity, filledQuote, pendingAmounts, metrics)
	} 
}

func (ob *Orderbook) checkSlippage(heap *heap.PriorityQueueHeap[*Order, uint64], order *Order, pendingAmounts *[]PendingAmt) bool {
	initial := order.Quantity * order.Price
	qty := order.Quantity
	var potential uint64
	var foo []*queue.LinkedMapQueue[*Order, uint64]
	for heap.Len() > 0 && 0 < qty {
		queue := heap.Peek()
		price := queue.Priority()
		vol := ob.volumeMap[queue.Priority()]
		toFill := min(vol, qty)
		qty -= toFill
		potential += toFill * price
		if qty != 0 {
			foo = append(foo, heap.PopQueue())
		}
	}
	for _, q := range foo {
		heap.PushQueue(q)
	}
	if initial < potential {
		ob.refundAmount(order, order.Quantity, pendingAmounts)
		return false
	}
	return true
}

func (ob *Orderbook) matchMarketOrder(order *Order, blockTs int64, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	heap := ob.getOppositeHeap(order.Side)
	var filledQuote uint64
	prevQuantity := order.Quantity

	if order.Side && !ob.checkSlippage(heap, order, pendingAmounts){
		return
	}

	for heap.Len() > 0 && 0 < order.Quantity {
		filledQuote += ob.fillPriceLevel(heap, order, blockTs, pendingAmounts, metrics)
	}

	if order.Side {
		toRefund := order.Price * prevQuantity - filledQuote
		ob.refundAmount(order, toRefund, pendingAmounts)
	}

	ob.fillMaker(order, blockTs, prevQuantity, filledQuote, pendingAmounts, metrics)
}

type PendingAmt struct {
	User    crypto.PublicKey
	TokenID ids.ID
	Amount  uint64
}

func (ob *Orderbook) toPendingAmount(order *Order, quantity uint64, isFilled bool, pendingAmounts *[]PendingAmt) {
	getAmount := GetAmountFn(order.Side, isFilled, ob.pair)
	if !isFilled && order.Fee > 0 {
		quantity += uint64(float64(quantity) * order.Fee)
	}
	amount, tokenID := getAmount(quantity, order.Price)
	*pendingAmounts = append(*pendingAmounts, PendingAmt{order.User, tokenID, amount * utils.MinQuantity()})
}

func (ob *Orderbook) fillAmount(order *Order, quantity uint64, pendingAmounts *[]PendingAmt) {
	ob.toPendingAmount(order, quantity, true, pendingAmounts)
}

func (ob *Orderbook) refundAmount(order *Order, quantity uint64, pendingAmounts *[]PendingAmt) {
	ob.toPendingAmount(order, quantity, false, pendingAmounts)
}