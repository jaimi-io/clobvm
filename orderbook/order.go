package orderbook

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Order struct {
	ID        ids.ID
	User      crypto.PublicKey
	Price     uint64
	Quantity  uint64
	Fee       float64
	Side      bool
}

func (o *Order) GetID() ids.ID {
	return o.ID
}

func NewOrder (id ids.ID, user crypto.PublicKey, price uint64, quantity uint64, side bool) *Order {
	return &Order{
		ID: id,
		User: user,
		Price: price,
		Quantity: utils.BalanceToQuantity(quantity),
		Side: side,
	}
}

func (o *Order) String() string {
	format := "ID: %s, P: %." + fmt.Sprint(consts.PriceDecimals) + "f, Q: %." + fmt.Sprint(consts.QuantityDecimals) + "f"
	return fmt.Sprintf(format, o.ID.String(), utils.DisplayPrice(o.Price), utils.DisplayQuantity(o.Quantity))
}