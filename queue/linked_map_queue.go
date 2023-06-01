package queue

import (
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"golang.org/x/exp/constraints"
)

type HasID interface {
	GetID() ids.ID
}

type Item[V HasID] struct {
	value V
	prevItem *Item[V]
	nextItem *Item[V]
}

type LinkedMapQueue[V HasID, S constraints.Ordered] struct {
	head *Item[V]
	tail *Item[V]
	hashMap map[ids.ID]*Item[V]
	priority S
	length int
}

func NewItem[V HasID](val V) *Item[V] {
	return &Item[V]{
		value: val,
		prevItem: nil,
		nextItem: nil,
	}
}

func NewLinkedMapQueue[V HasID,  S constraints.Ordered](val V, priority S) *LinkedMapQueue[V, S] {
	item := NewItem(val)
	hashmap := make(map[ids.ID]*Item[V])
	hashmap[val.GetID()] = item
	return &LinkedMapQueue[V, S]{
		head: item,
		tail: item,
		hashMap: hashmap,
		priority: priority,
		length: 1,
	}
}

func (lq *LinkedMapQueue[V, S]) Priority() S { return lq.priority }

func (lq *LinkedMapQueue[V, S]) Len() int { return lq.length }

func (lq *LinkedMapQueue[V, S]) Push(val V, id ids.ID) error {
	item := &Item[V]{
		value: val,
	}
	prevTail := lq.tail
	prevTail.nextItem = item
	item.prevItem = prevTail
	lq.tail = item
	lq.hashMap[id] = item
	lq.length++
	return nil
}

func (lq *LinkedMapQueue[V, S]) Peek() V {
	return lq.head.value
}

func (lq *LinkedMapQueue[V, S]) Pop() V {
	nextHead := lq.head.nextItem
	if nextHead != nil {
		nextHead.prevItem = nil
	}
	prevHead := lq.head
	lq.head = nextHead
	res := prevHead.value
	delete(lq.hashMap, res.GetID())
	lq.length--
	return res
}

func (lq *LinkedMapQueue[V, S]) Contains(id ids.ID) bool {
	_, ok := lq.hashMap[id]
	return ok
}

func (lq *LinkedMapQueue[V, S]) Remove(id ids.ID) error {
	item := lq.hashMap[id]
	if item == nil {
		return errors.New("Item not found")
	}
	prevItem := item.prevItem
	nextItem := item.nextItem

	if prevItem != nil {
		prevItem.nextItem = nextItem
	}
	if nextItem != nil {
		nextItem.prevItem = prevItem
	}

	if item == lq.head {
		lq.head = nextItem
	} else if item == lq.tail {
		lq.tail = prevItem
	}
	lq.length--
	return nil
}

func (lq *LinkedMapQueue[V, S]) Values() []V {
	var values []V
	for item := lq.head; item != nil; item = item.nextItem {
		values = append(values, item.value)
	}
	return values
}