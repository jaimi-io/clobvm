package consts

import (
	"github.com/jaimi-io/hypersdk/crypto"

	"github.com/ava-labs/avalanchego/ids"
)

var (
	ChainID ids.ID
	KeyPath string
	PrivKey crypto.PrivateKey
	URI string
	URIS []string
	GetPair bool
	GetAddress bool
)