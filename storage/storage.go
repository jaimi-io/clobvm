package storage

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/consts"
	"github.com/jaimi-io/hypersdk/crypto"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

var (
	balancePrefix = byte(0x1)
	orderPrefix   = byte(0x2)
)

func BalanceKey(pk crypto.PublicKey, tokenID ids.ID) []byte {
	key := make([]byte, 1+crypto.PublicKeyLen+consts.IDLen)
	key[0] = balancePrefix
	copy(key[1:1+crypto.PublicKeyLen], pk[:])
	copy(key[1+crypto.PublicKeyLen:1+crypto.PublicKeyLen+consts.IDLen], tokenID[:])
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

func GetBalanceFromState(ctx context.Context, f ReadState, pk crypto.PublicKey, tokenID ids.ID) (uint64, error) {
	key := BalanceKey(pk, tokenID)
	values, errs := f(ctx, [][]byte{key})
	bal, err := innerGetBalance(values[0], errs[0])
	return bal, err
}

func GetBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID) ([]byte, uint64, error) {
	key := BalanceKey(pk, tokenID)
	bal, err := db.GetValue(ctx, key)
	if errors.Is(err, database.ErrNotFound) {
		return key, 0, nil
	}
	return key, binary.BigEndian.Uint64(bal), err
}

func IncBalance(ctx context.Context, db chain.Database, pk crypto.PublicKey, tokenID ids.ID, amount uint64) error {
	key, bal, err := GetBalance(ctx, db, pk, tokenID)
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
	key, bal, err := GetBalance(ctx, db, pk, tokenID)
	if err != nil {
		return err
	}
	newBal, err := math.Sub(bal, amount)
	if err != nil {
		return fmt.Errorf("invalid subtract balance (token=%s, bal=%d, addr=%v, amount=%d)", tokenID, bal, crypto.Address("clob", pk), amount)
	}
	err = db.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, newBal))
	return err
}

func OrderKey(orderID ids.ID) []byte {
	key := make([]byte, 1+consts.IDLen)
	key[0] = orderPrefix
	copy(key[1:1+consts.IDLen], orderID[:])
	return key
}

func SetOrder(ctx context.Context, db chain.Database, orderID ids.ID, amount uint64) error{
	key := OrderKey(orderID)
	return db.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, amount))
}

func UpdateOrder(ctx context.Context, db chain.Database, orderID ids.ID, amount uint64) error {
	key := OrderKey(orderID)
	return db.Insert(ctx, key, binary.BigEndian.AppendUint64(nil, amount))
}

func RemoveOrder(ctx context.Context, db chain.Database, orderID ids.ID) (uint64, error){
	key := OrderKey(orderID)
	val, err := db.GetValue(ctx, key)
	if err != nil {
		return 0, err
	}
	db.Remove(ctx, key)
	return binary.BigEndian.Uint64(val), nil
}
