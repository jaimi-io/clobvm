package orderbook

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/heap"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Order struct {
	ID        ids.ID
	User      crypto.PublicKey
	Price     uint64
	Quantity  uint64
	Side      bool
}

func (o *Order) GetID() ids.ID {
	return o.ID
}

func NewOrder (id ids.ID, user crypto.PublicKey, price uint64, quantity uint64, side bool) *Order {
	return &Order{
		ID: id,
		User: user,
		Price: price,
		Quantity: quantity,
		Side: side,
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("ID: %s, Price: %d, Quantity: %d", o.ID.String(), o.Price, o.Quantity)
}

type Orderbook struct {
	minHeap *heap.PriorityQueueHeap[*Order, uint64]
	maxHeap *heap.PriorityQueueHeap[*Order, uint64]

	orderMap map[ids.ID]*Order
	volumeMap map[uint64]uint64
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		minHeap: heap.NewPriorityQueueHeap[*Order, uint64](1024, true),
		maxHeap: heap.NewPriorityQueueHeap[*Order, uint64](1024, false),
		orderMap: make(map[ids.ID]*Order),
		volumeMap: make(map[uint64]uint64),
	}
}

func (ob *Orderbook) Add(order *Order, tokenID ids.ID, oppTokenID ids.ID, pendingAmounts *[]PendingAmt) {
	ob.orderMap[order.ID] = order
	ob.matchOrder(order, tokenID, oppTokenID, pendingAmounts)
	if order.Quantity > 0 {
		ob.volumeMap[order.Price] += order.Quantity
		if order.Side {
			ob.maxHeap.Add(order, order.ID, order.Price)
		} else {
			ob.minHeap.Add(order, order.ID, order.Price)
		}
	}
}

func (ob *Orderbook) Get(id ids.ID) *Order {
	order := ob.orderMap[id]
	return order
}

func (ob *Orderbook) Cancel(order *Order, tokenID ids.ID, pendingAmounts *[]PendingAmt) {
	ob.volumeMap[order.Price] -= order.Quantity
	if order.Side {
		ob.maxHeap.Remove(order.ID, order.Price)
	} else {
		ob.minHeap.Remove(order.ID, order.Price)
	}
	delete(ob.orderMap, order.ID)
	ob.toPendingAmount(order, tokenID, pendingAmounts)
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

