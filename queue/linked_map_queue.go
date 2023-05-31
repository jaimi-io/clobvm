package queue

import (
	"github.com/ava-labs/avalanchego/ids"
	"golang.org/x/exp/constraints"
)

type Item[V any] struct {
	value V
	prevItem *Item[V]
	nextItem *Item[V]
}

type LinkedMapQueue[V any, S constraints.Ordered] struct {
	head *Item[V]
	tail *Item[V]
	hashMap map[ids.ID]*Item[V]
	priority S
	length int
}

func NewItem[V any](val V) *Item[V] {
	return &Item[V]{
		value: val,
		prevItem: nil,
		nextItem: nil,
	}
}

func NewLinkedMapQueue[V any,  S constraints.Ordered](val V, priority S) *LinkedMapQueue[V, S] {
	item := NewItem(val)
	return &LinkedMapQueue[V, S]{
		head: item,
		tail: item,
		hashMap: make(map[ids.ID]*Item[V]),
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
	nextHead.prevItem = nil
	curHead := lq.head
	lq.head = nextHead
	lq.length--
	return curHead.value
}

func (lq *LinkedMapQueue[V, S]) Remove(id ids.ID) error {
	item := lq.hashMap[id]
	prevItem := item.prevItem
	nextItem := item.nextItem
	prevItem.nextItem = nextItem
	nextItem.prevItem = prevItem
	lq.length--
	return nil
}