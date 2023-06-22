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
	openOrders map[crypto.PublicKey]map[ids.ID]struct{}
	executionHistory map[crypto.PublicKey]*MonthlyExecuted
	midPrice *VersionedBalance
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
		openOrders: make(map[crypto.PublicKey]map[ids.ID]struct{}),
		midPrice: NewVersionedBalance(0, 0),
	}
}

func (ob *Orderbook) Add(order *Order, blockHeight uint64, blockTs int64, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	marketOrder := order.Price == 0
	if marketOrder {
		order.Price = ob.GetMidPrice()
	}
	if marketOrder && ((order.Side && ob.sellSideVolume < order.Quantity) || (!order.Side && ob.buySideVolume < order.Quantity)) {
		feeToReturn := ob.RefundMarketOrderFee(order.User, blockTs, order.Quantity)
		if feeToReturn > 0 {
			ob.toPendingAmount(order, feeToReturn, false, pendingAmounts)
		}
		return
	}
	ob.matchOrder(order, blockTs, marketOrder, pendingAmounts, metrics)
	if order.Quantity > 0 {
		feeToReturn := ob.RefundFee(order.User, blockTs, order.Quantity)
		if feeToReturn > 0 {
			ob.toPendingAmount(order, feeToReturn, false, pendingAmounts)
		}

		ob.volumeMap[order.Price] += order.Quantity
		ob.orderMap[order.ID] = order
		ob.AddToEviction(order.ID, blockHeight)
		order.Fee = ob.GetFeeRate(order.User, blockTs)

		if _, ok := ob.openOrders[order.User]; !ok {
			ob.openOrders[order.User] = make(map[ids.ID]struct{})
		}
		ob.openOrders[order.User][order.ID] = struct{}{}

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

func (ob *Orderbook) CancelAll(user crypto.PublicKey, pendingAmounts *[]PendingAmt, metrics *metrics.Metrics) {
	if _, ok := ob.openOrders[user]; !ok {
		return
	}
	for id := range ob.openOrders[user] {
		order := ob.orderMap[id]
		ob.Cancel(order, pendingAmounts, metrics)
	}
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
	metrics.OrderCancelNum()
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
	delete(ob.openOrders[order.User], order.ID)
	metrics.OrderNumDec()
	metrics.OrderAmountSub(order.Quantity)
}

func (ob *Orderbook) GetBuySide() [][]*Order {
	return ob.maxHeap.Values()
}

func (ob *Orderbook) GetMidPrice() uint64 {
	if ob.maxHeap.Len() == 0 && ob.minHeap.Len() == 0 {
		return 0
	}
	if ob.maxHeap.Len() == 0 {
		return ob.minHeap.Peek().Priority()
	}
	if ob.minHeap.Len() == 0 {
		return ob.maxHeap.Peek().Priority()
	}
	maxQueue := ob.maxHeap.Peek()
	minQueue := ob.minHeap.Peek()
	return (maxQueue.Priority() + minQueue.Priority()) / 2
}

func (obm *OrderbookManager) UpdateAllMidPrices(blockNumber uint64) {
	for pair := range obm.orderbooks {
		ob := obm.orderbooks[pair]
		ob.AddMidPriceBlk(blockNumber)
	}
}

func (ob *Orderbook) GetMidPriceBlk(blockHeight uint64) uint64 {
	if blockHeight < consts.PendingBlockWindow {
		return 0
	}
	mid, _ := ob.midPrice.Get(blockHeight - consts.PendingBlockWindow)
	return mid
}

func (ob *Orderbook) AddMidPriceBlk(blockHeight uint64) {
	ob.midPrice.Put(ob.GetMidPrice(), blockHeight)
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

