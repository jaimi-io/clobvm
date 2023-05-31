package heap

import (
	"container/heap"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/queue"
	"golang.org/x/exp/constraints"
)

type Item[V any, S constraints.Ordered] struct {
	Queue *queue.LinkedMapQueue[V, S]
	Index int
}

type PriorityQueueHeap[V any, S constraints.Ordered] struct {
	items []*Item[V, S]
	hashMap map[S]*Item[V, S]
	isMinHeap bool
}

func NewPriorityQueueHeap[V any, S constraints.Ordered](size int, isMinHeap bool) *PriorityQueueHeap[V, S] {
	ph := &PriorityQueueHeap[V, S]{
		items: make([]*Item[V, S], 0, size),
		hashMap: make(map[S]*Item[V, S]),
		isMinHeap: isMinHeap,
	}
	heap.Init(ph)
	return ph
}

func (ph *PriorityQueueHeap[V, S]) Len() int { return len(ph.items) }

func (ph *PriorityQueueHeap[V, S]) Less(i, j int) bool {
	if ph.isMinHeap {
		return ph.items[i].Queue.Priority() > ph.items[j].Queue.Priority()
	}
	return ph.items[i].Queue.Priority() < ph.items[j].Queue.Priority()
}

func (ph *PriorityQueueHeap[V, S]) Swap(i, j int) {
	ph.items[i], ph.items[j] = ph.items[j], ph.items[i]
	ph.items[i].Index = i
	ph.items[j].Index = j
}

func (ph *PriorityQueueHeap[V, S]) Push(x any) {
	queue := x.(*queue.LinkedMapQueue[V, S])
	item := &Item[V, S]{Queue: queue, Index: len(ph.items)}
	ph.items = append(ph.items, item)
	ph.hashMap[queue.Priority()] = item
}

func (ph *PriorityQueueHeap[V, S]) Pop() any {
	n := len(ph.items)
	item := ph.items[n-1]
	ph.items = ph.items[0:n-1]
	delete(ph.hashMap, item.Queue.Priority())
	return item.Queue
}

func (ph *PriorityQueueHeap[V, S]) Peek() *queue.LinkedMapQueue[V, S] {
	n := len(ph.items)
	return ph.items[n-1].Queue
}

func (ph *PriorityQueueHeap[V, S]) Contains(priority S) bool {
	_, ok := ph.hashMap[priority]
	return ok
}

func (ph *PriorityQueueHeap[V, S]) Get(priority S) (*queue.LinkedMapQueue[V, S], *Item[V, S]) {
	item, ok := ph.hashMap[priority]
	if !ok {
		return nil, nil
	}
	return item.Queue, item
}

func (ph *PriorityQueueHeap[V, S]) Add(value V, id ids.ID, priority S) {
	lq, _ := ph.Get(priority)
	if lq == nil {
		lq = queue.NewLinkedMapQueue(value, priority)
		heap.Push(ph, lq)
	} else {
		lq.Push(value, id)
	}
}

func (ph *PriorityQueueHeap[V, S]) Remove(id ids.ID, priority S) error {
	lq, item := ph.Get(priority)
	if lq == nil {
		return errors.New("PriorityQueueHeap.Remove: priority not found")
	}
	err := lq.Remove(id)
	if err != nil {
		return err
	}
	if lq.Len() == 0 {
		heap.Remove(ph, item.Index)
	}
	return nil
}

func (ph *PriorityQueueHeap[V, S]) Values() [][]V {
	values := make([][]V, 0, len(ph.items))
	for _, item := range ph.items {
		values = append(values, item.Queue.Values())
	}
	return values
}