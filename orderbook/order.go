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

func toDecimal(value uint64) float64 {
	decimals := math.Pow(10, float64(6))
	return float64(value) / decimals
}

func (o *Order) String() string {
	return fmt.Sprintf("ID: %s, Price: %.6f, Quantity: %.6f", o.ID.String(), toDecimal(o.Price), toDecimal(o.Quantity))
}