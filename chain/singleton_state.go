package chain

import "github.com/ava-labs/avalanchego/database"

const (
	IsInitializedKey byte = iota
)

var (
	isInitializedKey                = []byte{IsInitializedKey}
	_                SingletonState = (*singletonState)(nil)
)

// SingletonState is a thin wrapper around a database to provide, caching,
// serialization, and deserialization of singletons.
type SingletonState interface {
	IsInitialized() (bool, error)
	SetInitialized() error
}

type singletonState struct {
	singletonDB database.Database
}

func NewSingletonState(db database.Database) SingletonState {
	return &singletonState{
		singletonDB: db,
	}
}

func (s *singletonState) IsInitialized() (bool, error) {
	return s.singletonDB.Has(isInitializedKey)
}

func (s *singletonState) SetInitialized() error {
	return s.singletonDB.Put(isInitializedKey, nil)
}
