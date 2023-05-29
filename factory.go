package vm

import (
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/jaimi-io/clobvm/controller"
)

type Factory struct {}

func (*Factory) New(logging.Logger) (interface{}, error) {
	return controller.New(), nil
}