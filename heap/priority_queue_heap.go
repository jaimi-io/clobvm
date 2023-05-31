package heap

import (
	"container/heap"

	"github.com/jaimi-io/clobvm/queue"
	"golang.org/x/exp/constraints"
)

type PriorityQueueHeap[V any, S constraints.Ordered] struct {
	items []*queue.LinkedMapQueue[V, S]
	hashMap map[S]*queue.LinkedMapQueue[V, S]
	isMinHeap bool
}

func NewPriorityQueueHeap[V any, S constraints.Ordered](size int, isMinHeap bool) *PriorityQueueHeap[V, S] {
	ph := &PriorityQueueHeap[V, S]{
		items: make([]*queue.LinkedMapQueue[V, S], 0, size),
		hashMap: make(map[S]*queue.LinkedMapQueue[V, S]),
		isMinHeap: isMinHeap,
	}
	heap.Init(ph)
	return ph
}

func (ph *PriorityQueueHeap[V, S]) Len() int { return len(ph.items) }

func (ph *PriorityQueueHeap[V, S]) Less(i, j int) bool {
	if ph.isMinHeap {
		return ph.items[i].Priority() > ph.items[j].Priority()
	}
	return ph.items[i].Priority() < ph.items[j].Priority()
}

func (ph *PriorityQueueHeap[V, S]) Swap(i, j int) {
	ph.items[i], ph.items[j] = ph.items[j], ph.items[i]
}

func (ph *PriorityQueueHeap[V, S]) Push(x any) {
	item := x.(*queue.LinkedMapQueue[V, S])
	ph.items = append(ph.items, item)
	ph.hashMap[item.Priority()] = item
}

func (ph *PriorityQueueHeap[V, S]) Pop() any {
	n := len(ph.items)
	item := ph.items[n-1]
	ph.items = ph.items[0:n-1]
	delete(ph.hashMap, item.Priority())
	return item
}

func (ph *PriorityQueueHeap[V, S]) Peek() *queue.LinkedMapQueue[V, S] {
	n := len(ph.items)
	return ph.items[n-1]
}

func (ph *PriorityQueueHeap[V, S]) Contains(priority S) bool {
	_, ok := ph.hashMap[priority]
	return ok
}

func (ph *PriorityQueueHeap[V, S]) Get(priority S) *queue.LinkedMapQueue[V, S] {
	return ph.hashMap[priority]
}