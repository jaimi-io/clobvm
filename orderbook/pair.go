package orderbook

import "github.com/ava-labs/avalanchego/ids"

type Pair struct {
	BaseTokenID ids.ID
	QuoteTokenID ids.ID
}

func (p *Pair) TokenID(side bool, isAdd bool) ids.ID {
	if side && isAdd || !side && !isAdd {
		return p.QuoteTokenID
	} 
	return p.BaseTokenID
}

func GetAmountFn(side bool, isAdd bool) func (q, p uint64) uint64 {
	if side && isAdd || !side && !isAdd {
		return func(q, p uint64) uint64 { return q * p}
	}
	return func(q, p uint64) uint64 { return q }
}