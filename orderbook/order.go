package orderbook

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Order struct {
	ID          ids.ID
	User        crypto.PublicKey
	Price       uint64
	Quantity    uint64
	Fee         float64
	Side        bool
	BlockExpiry uint64
}

func (o *Order) GetID() ids.ID {
	return o.ID
}

func (o *Order) GetQuantity() uint64 {
	return o.Quantity
}

func NewOrder (id ids.ID, user crypto.PublicKey, price uint64, quantity uint64, side bool, currentBlock uint64, blockWindow uint64) *Order {
	return &Order{
		ID: id,
		User: user,
		Price: price,
		Quantity: utils.BalanceToQuantity(quantity),
		Side: side,
		BlockExpiry: currentBlock + blockWindow,
	}
}

func (o *Order) String() string {
	format := "ID: %s, P: %." + fmt.Sprint(consts.PriceDecimals) + "f, Q: %." + fmt.Sprint(consts.QuantityDecimals) + "f"
	return fmt.Sprintf(format, o.ID.String(), utils.DisplayPrice(o.Price), utils.DisplayQuantity(o.Quantity))
}