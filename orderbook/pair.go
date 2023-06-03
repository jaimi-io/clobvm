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