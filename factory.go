package vm

import (
	"github.com/jaimi-io/clobvm/chain"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/vms"
)

var _ vms.Factory = &Factory{}

// Factory ...
type Factory struct{}

// New ...
func (*Factory) New(logging.Logger) (interface{}, error) { return &chain.VM{}, nil }
