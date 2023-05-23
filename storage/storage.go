package storage

import (
	"context"
	"encoding/binary"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/consts"
	"github.com/jaimi-io/hypersdk/crypto"
)

func BalanceKey(pk crypto.PublicKey, tokenID ids.ID) []byte {
	key := make([]byte, 0, crypto.PublicKeyLen+consts.IDLen)
	copy(key[0:crypto.PublicKeyLen], pk[:])
	copy(key[crypto.PublicKeyLen:crypto.PublicKeyLen+consts.IDLen], tokenID[:])
	return key
}

func getBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID) ([]byte, uint64, error) {
	key := BalanceKey(pk, tokenID)
	bal, err := db.GetValue(ctx, key)
	return key, binary.BigEndian.Uint64(bal), err
}

func GetBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID) (uint64, error) {
	_, bal, err := getBalance(ctx, db, pk, tokenID)
	return bal, err
}

func IncBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID, amount uint64) error {
	key, bal, err := getBalance(ctx, db, pk, tokenID)
	if err != nil {
		return err
	}
	newBal, err := math.Add64(bal, amount)
	if err != nil {
		return err
	}
	err = db.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, newBal))
	return err
}

func DecBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID, amount uint64) error {
	key, bal, err := getBalance(ctx, db, pk, tokenID)
	if err != nil {
		return err
	}
	newBal, err := math.Sub(bal, amount)
	if err != nil {
		return err
	}
	err = db.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, newBal))
	return err
}
