package orderbook

import (
	"fmt"
	"sort"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/heap"
	"github.com/jaimi-io/clobvm/utils"
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

func (ob *Orderbook) Add(order *Order, blockHeight uint64, pendingAmounts *[]PendingAmt) {
	ob.matchOrder(order, pendingAmounts)
	if order.Quantity > 0 {
		ob.volumeMap[order.Price] += order.Quantity
		ob.orderMap[order.ID] = order
		ob.AddToEviction(order.ID, blockHeight)
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

