package controller

import (
	"github.com/ava-labs/avalanchego/ids"
)

type StateManager struct{}

func (*StateManager) HeightKey() []byte {
	return []byte{0x1}
}

func (*StateManager) IncomingWarpKey(sourceChainID ids.ID, msgID ids.ID) []byte {
	return []byte{}
}

func (*StateManager) OutgoingWarpKey(txID ids.ID) []byte {
	return []byte{}
}
