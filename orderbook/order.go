package orderbook

import (
	"fmt"
	"math"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Order struct {
	ID        ids.ID
	User      crypto.PublicKey
	Price     uint64
	Quantity  uint64
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
		Quantity: quantity,
		Side: side,
	}
}

func (o *Order) String() string {
	decimals := float64(-9)
	quantity := math.Pow(float64(o.Quantity), decimals)
	return fmt.Sprintf("ID: %s, Price: %d, Quantity: %7.9f", o.ID.String(), o.Price, quantity)
}