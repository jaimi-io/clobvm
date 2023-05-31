package registry

import (
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
)

var (
	ActionRegistry *codec.TypeParser[chain.Action, *warp.Message, bool]
	AuthRegistry   *codec.TypeParser[chain.Auth, *warp.Message, bool]
)

func init() {
	ActionRegistry = codec.NewTypeParser[chain.Action, *warp.Message]()
	AuthRegistry = codec.NewTypeParser[chain.Auth, *warp.Message]()

	_ = ActionRegistry.Register(&actions.Transfer{}, actions.UnmarshalTransfer, false)
	_ = ActionRegistry.Register(&actions.AddOrder{}, actions.UnmarshalAddOrder, false)
	_ = AuthRegistry.Register(&auth.EIP712{}, auth.UnmarshalEIP712, false)
}