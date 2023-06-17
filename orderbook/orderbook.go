package orderbook

import (
	"fmt"
	"sort"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/heap"
	"github.com/jaimi-io/clobvm/metrics"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Orderbook struct {
	pair Pair
	minHeap *heap.PriorityQueueHeap[*Order, uint64]
	maxHeap *heap.PriorityQueueHeap[*Order, uint64]

	orderMap map[ids.ID]*Order
	buySideVolume uint64
	sellSideVolume uint64
	volumeMap map[uint64]uint64
	evictionMap map[uint64]map[ids.ID]struct{}
	executionHistory map[crypto.PublicKey]*MonthlyExecuted
}

func NewOrderbook(pair Pair) *Orderbook {
	return &Orderbook{
		pair: pair,
		minHeap: heap.NewPriorityQueueHeap[*Order, uint64](1024, true),
		maxHeap: heap.NewPriorityQueueHeap[*Order, uint64](1024, false),
		orderMap: make(map[ids.ID]*Order),
		volumeMap: make(map[uint64]uint64),
		evictionMap: make(map[uint64]map[ids.ID]struct{}),
		executionHistory: make(map[crypto.PublicKey]*MonthlyExecuted),
	}
}

func (ob *Orderbook) Add(order *Order, blockHeight uint64, blockTs int64, marketOrder bool, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	fmt.Printf("Add order: %v\n", order)
	fmt.Printf("Orderbook: %v\n", ob)
	if marketOrder && ((order.Side && ob.sellSideVolume < order.Quantity) || (!order.Side && ob.buySideVolume < order.Quantity)) {
		feeToReturn := ob.RefundMarketOrderFee(order.User, blockTs, order.Quantity)
		if feeToReturn > 0 {
			order.Price = ob.GetMidPrice()
			ob.toPendingAmount(order, feeToReturn, false, pendingAmounts)
		}
		return
	}
	ob.matchOrder(order, blockTs, pendingAmounts, metrics)
	if order.Quantity > 0 {
		feeToReturn := ob.RefundFee(order.User, blockTs, order.Quantity)
		if feeToReturn > 0 {
			ob.toPendingAmount(order, feeToReturn, false, pendingAmounts)
		}

		ob.volumeMap[order.Price] += order.Quantity
		ob.orderMap[order.ID] = order
		ob.AddToEviction(order.ID, blockHeight)
		order.Fee = ob.GetFeeRate(order.User, blockTs)

		if order.Side {
			ob.maxHeap.Add(order, order.ID, order.Price)
			ob.buySideVolume += order.Quantity
		} else {
			ob.minHeap.Add(order, order.ID, order.Price)
			ob.sellSideVolume += order.Quantity
		}

		metrics.OrderNumInc()
		metrics.OrderAmountAdd(order.Quantity)
	}
}

func (ob *Orderbook) Get(id ids.ID) *Order {
	return ob.orderMap[id]
}

func (ob *Orderbook) Cancel(order *Order, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	if order.Side {
		ob.maxHeap.Remove(order.ID, order.Price)
	} else {
		ob.minHeap.Remove(order.ID, order.Price)
	}
	ob.Remove(order, metrics)
	isFilled := false
	ob.toPendingAmount(order, order.Quantity, isFilled, pendingAmounts)
}

func (ob *Orderbook) Remove(order *Order, metrics *metrics.Metrics) {
	ob.volumeMap[order.Price] -= order.Quantity
	if order.Side {
		ob.buySideVolume -= order.Quantity
	} else {
		ob.sellSideVolume -= order.Quantity
	}
	delete(ob.orderMap, order.ID)
	delete(ob.evictionMap[order.BlockExpiry], order.ID)
	metrics.OrderNumDec()
	metrics.OrderAmountSub(order.Quantity)
}

func (ob *Orderbook) GetBuySide() [][]*Order {
	return ob.maxHeap.Values()
}

func (ob *Orderbook) GetMidPrice() uint64 {
	if ob.maxHeap.Len() == 0 || ob.minHeap.Len() == 0 {
		return 0
	}
	maxQueue := ob.maxHeap.Peek()
	minQueue := ob.minHeap.Peek()
	return (maxQueue.Priority() + minQueue.Priority()) / 2
}

func (ob *Orderbook) GetSellSide() [][]*Order {
	return ob.minHeap.Values()
}

func (ob *Orderbook) GetVolumes() string {
	priceLevels := len(ob.volumeMap)
	prices := make([]int, 0, priceLevels)
	for price := range ob.volumeMap {
		prices = append(prices, int(price))
	}
	sort.Ints(prices)
	var outputStr string
	format := "%." + fmt.Sprint(consts.PriceDecimals) + "f : %." + fmt.Sprint(consts.QuantityDecimals) + "f\n"
	for i := priceLevels - 1; i >= 0; i-- {
		outputStr += fmt.Sprintf(format, utils.DisplayPrice(uint64(prices[i])), utils.DisplayQuantity(ob.volumeMap[uint64(prices[i])]))
	}
	return fmt.Sprint(outputStr)
}

