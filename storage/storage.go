package storage

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/consts"
	"github.com/jaimi-io/hypersdk/crypto"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

func BalanceKey(pk crypto.PublicKey, tokenID ids.ID) []byte {
	key := make([]byte, crypto.PublicKeyLen+consts.IDLen)
	copy(key[0:crypto.PublicKeyLen], pk[:])
	copy(key[crypto.PublicKeyLen:crypto.PublicKeyLen+consts.IDLen], tokenID[:])
	return key
}

func SetBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID, amount uint64) error {
	key := BalanceKey(pk, tokenID)
	err := db.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, amount))
	return err
}

func innerGetBalance(
	v []byte,
	err error,
) (uint64, error) {
	if errors.Is(err, database.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(v), nil
}

func GetBalance(ctx context.Context, f ReadState, pk crypto.PublicKey, tokenID ids.ID) (uint64, error) {
	key := BalanceKey(pk, tokenID)
	values, errs := f(ctx, [][]byte{key})
	bal, err := innerGetBalance(values[0], errs[0])
	return bal, err
}

func getBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID) ([]byte, uint64, error) {
	key := BalanceKey(pk, tokenID)
	bal, err := db.GetValue(ctx, key)
	if errors.Is(err, database.ErrNotFound) {
		return key, 0, nil
	}
	return key, binary.BigEndian.Uint64(bal), err
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
