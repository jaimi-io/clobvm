package controller

import (
	"github.com/ava-labs/avalanchego/ids"
)

type StateManager struct{}

func (*StateManager) HeightKey() []byte {
	return nil
}

func (*StateManager) IncomingWarpKey(sourceChainID ids.ID, msgID ids.ID) []byte {
	return nil
}

func (*StateManager) OutgoingWarpKey(txID ids.ID) []byte {
	return nil
}
