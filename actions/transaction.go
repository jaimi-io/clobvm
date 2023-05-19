package actions

import (
	"crypto/rand"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	Sender  common.Address
	Volume int    `json:"volume"`
	Pair	 string `json:"pair"`
	Price  int    `json:"price"`
	Side bool		`json:"side"`
	Likelihood float64 `json:"likelihood"`
}

func NewTx(sender common.Address, volume int, pair string, price int, side bool) *Transaction {
	randVolume, _ := rand.Int(rand.Reader, big.NewInt(int64(volume)))
	likelihood := float64(randVolume.Int64()) / 10000000
	return &Transaction{
		Sender: sender,
		Volume: volume,
		Pair: pair,
		Price: price,
		Side: side,
		Likelihood: likelihood,
	}
}

func (t *Transaction) Execute() error {
	return nil
}