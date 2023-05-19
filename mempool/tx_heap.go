package mempool

import "github.com/jaimi-io/clobvm/actions"

type txHeap struct {
	txs []*actions.Transaction
}

func newTxHeap() *txHeap {
	return &txHeap{
		txs: []*actions.Transaction{},
	}
}

func (h *txHeap) Len() int {
	return len(h.txs)
}

func (h *txHeap) Less(i, j int) bool {
	return h.txs[i].Likelihood > h.txs[j].Likelihood
}

func (h *txHeap) Swap(i, j int) {
	h.txs[i], h.txs[j] = h.txs[j], h.txs[i]
}

func (h *txHeap) Push(x interface{}) {
	h.txs = append(h.txs, x.(*actions.Transaction))
}

func (h *txHeap) Pop() interface{} {
	old := h.txs
	n := len(old)
	x := old[n-1]
	h.txs = old[0 : n-1]
	return x
}