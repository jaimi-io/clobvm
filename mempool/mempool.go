package mempool

import (
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/chain"
)

var _ chain.Mempool = &Mempool{}

type Mempool struct {
	tx *actions.Transaction
}

func (mem *Mempool) Len() int {
	return 0
}
// Prune(ids.Set)
func (mem *Mempool) PopMax() (*actions.Transaction, uint64) {
	return mem.tx, 0
}

func (mem *Mempool) Add(tx *actions.Transaction) bool {
	mem.tx = tx
	return true
}

func (mem *Mempool) NewTxs(uint64) ([]*actions.Transaction) {
	return []*actions.Transaction{}
}
