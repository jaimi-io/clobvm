package chain

import (
	"github.com/jaimi-io/clobvm/actions"
)

type Mempool interface {
	Len() int
	PopMax() (*actions.Transaction, uint64)
	Add(*actions.Transaction) bool
	NewTxs(uint64) []*actions.Transaction
}