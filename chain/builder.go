package chain

import (
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
)

func BuildBlock(vm VM) (snowman.Block, error) {
	// TODO: Build a block
	return vm.BuildBlock()
}